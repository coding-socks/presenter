package presenter

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(
		extension.Strikethrough,
		extension.TaskList,
		NewTwitterLinkExtension(),
		NewAnchorTargetBlankExtension(),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

type anchorTargetBlankTransformer struct{}

func NewAnchorTargetBlankTransformer() *anchorTargetBlankTransformer {
	return &anchorTargetBlankTransformer{}
}

func (t *anchorTargetBlankTransformer) Transform(n *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch n.Kind() {
			case ast.KindLink, ast.KindAutoLink:
				n.SetAttributeString("target", []byte("_blank"))
				// https://developers.google.com/web/tools/lighthouse/audits/noopener
				n.SetAttributeString("rel", []byte("noopener"))
			}
		}
		return ast.WalkContinue, nil
	})
}

type anchorTargetBlankExtension struct{}

func NewAnchorTargetBlankExtension() *anchorTargetBlankExtension {
	return &anchorTargetBlankExtension{}
}

func (e *anchorTargetBlankExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(NewAnchorTargetBlankTransformer(), 10)))
}

type twitterLinkExtension struct{}

// NewTwitterLinkExtension returns a new extension.
func NewTwitterLinkExtension() goldmark.Extender {
	return &twitterLinkExtension{}
}

func (e *twitterLinkExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewTwitterLinkParser(), 999),
		),
	)
}

var twitterDestinationPrefix = []byte("https://twitter.com/")

type twitterLinkParser struct {
}

var defaultTwitterLinkParser = &twitterLinkParser{}

// NewTwitterLinkParser return a new InlineParser that parses links.
func NewTwitterLinkParser() parser.InlineParser {
	return defaultTwitterLinkParser
}

func (s *twitterLinkParser) Trigger() []byte {
	return []byte{'<'}
}

func (s *twitterLinkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()
	stop := FindTwitterIndex(line[1:])
	if stop < 0 {
		return nil
	}
	stop++
	if stop >= len(line) || line[stop] != '>' {
		return nil
	}
	value := ast.NewTextSegment(text.NewSegment(segment.Start+1, segment.Start+stop))
	block.Advance(stop + 1)
	link := ast.NewLink()
	link.AppendChild(link, value)
	link.Destination = append(twitterDestinationPrefix, line[1:stop]...)
	return link
}

// FindTwitterIndex returns a stop index value if the given bytes seem an email address.
func FindTwitterIndex(b []byte) int {
	i := 0
	if b[i] != '@' {
		return -1
	}
	i++
	for ; i < len(b); i++ {
		if !util.IsAlphaNumeric(b[i]) && b[i] != '_' {
			break
		}
	}
	return i
}
