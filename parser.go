package presenter

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"html/template"
	"io"
	"log"
	"strings"
	"time"
)

// ParseMode represents flags for the Parse function.
type ParseMode int

const (
	// TitlesOnly mode parse only the title and subtitle.
	TitlesOnly ParseMode = 1 << (1 + iota)
)

func ParseSlide(r io.Reader, mode ParseMode) (*Doc, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	reader := text.NewReader(content)
	ndoc := markdown.Parser().Parse(reader)
	title, metadata, sections := parseSlide(ndoc, mode)
	doc := new(Doc)
	if title != nil {
		doc.Title = string(title.Text(content))
	}
	for _, meta := range metadata {
		if meta.Kind() != ast.KindParagraph {
			return nil, fmt.Errorf("unexpected node type in metadata: %s", meta.Kind())
		}
		if !meta.HasBlankPreviousLines() {
			if err := parseHeader(doc, meta, content); err != nil {
				return nil, err
			}
		} else {
			doc.Authors = append(doc.Authors, parseAuthor(meta, content))
		}
	}
	for _, section := range sections {
		d := ast.NewDocument()
		for _, node := range section.nodes {
			d.AppendChild(d, node)
		}
		if !d.HasChildren() {
			section.title.SetAttributeString("class", []byte("empty"))
		}
		doc.Sections = append(doc.Sections, Section{
			title:   section.title,
			content: content,
			doc:     d,
		})
	}
	return doc, nil
}

func parseHeader(doc *Doc, meta ast.Node, content []byte) error {
	lines := meta.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		text := strings.TrimSpace(string(line.Value(content)))
		if strings.HasPrefix(text, "Tags:") {
			tags := strings.Split(text[len("Tags:"):], ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
			doc.Tags = tags
		} else if strings.HasPrefix(text, "Summary:") {
			doc.Summary = strings.TrimSpace(text[len("Summary:"):])
		} else if t, ok := parseTime(text); ok {
			doc.Time = t
		} else if doc.Subtitle == "" {
			doc.Subtitle = text
		} else {
			return fmt.Errorf("unexpected header line: %q", text)
		}
	}
	return nil
}

func parseTime(text string) (t time.Time, ok bool) {
	t, err := time.Parse("15:04 2 Jan 2006", text)
	if err == nil {
		return t, true
	}
	t, err = time.Parse("2 Jan 2006", text)
	if err == nil {
		// at 11am UTC it is the same date everywhere
		t = t.Add(time.Hour * 11)
		return t, true
	}
	return time.Time{}, false
}

// Doc represents an entire document.
type Doc struct {
	Title      string
	Subtitle   string
	Summary    string
	Time       time.Time
	Authors    []Author
	TitleNotes []string
	Sections   []Section
	Tags       []string
}

func parseAuthor(n ast.Node, content []byte) Author {
	var nodes []ast.Node
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		nodes = append(nodes, c)
	}
	return Author{
		content: content,
		node:    n,
		nodes:   nodes,
	}
}

// Author represents the person who wrote and/or is presenting the document.
type Author struct {
	content []byte
	node    ast.Node
	nodes   []ast.Node
}

func (a Author) RenderText() template.HTML {
	p := ast.NewParagraph()
	for i := range a.nodes {
		k := a.nodes[i].Kind()
		if k != ast.KindText {
			continue
		}
		p.AppendChild(p, a.nodes[i])
	}
	var b bytes.Buffer
	if err := markdown.Renderer().Render(&b, a.content, p); err != nil {
		log.Print(err)
	}
	return template.HTML(bytes.ReplaceAll(b.Bytes(), []byte{'\n'}, []byte("<br>\n")))
}

func (a Author) Render() template.HTML {
	p := ast.NewParagraph()
	for i := range a.nodes {
		p.AppendChild(p, a.nodes[i])
	}
	var b bytes.Buffer
	if err := markdown.Renderer().Render(&b, a.content, p); err != nil {
		log.Print(err)
	}
	return template.HTML(bytes.ReplaceAll(b.Bytes(), []byte{'\n'}, []byte("<br>\n")))
}

// Section represents a section of a document (such as a presentation slide)
// comprising a title and a list of elements.
type Section struct {
	Number []int
	title  ast.Node
	Notes  []string

	content []byte
	doc     *ast.Document
}

func (s Section) Empty() bool {
	return !s.doc.HasChildren()
}

func (s Section) RenderTitle() template.HTML {
	var b strings.Builder
	if err := markdown.Renderer().Render(&b, s.content, s.title); err != nil {
		log.Print(err)
	}
	return template.HTML(b.String())
}

func (s Section) Render() template.HTML {
	var b strings.Builder
	if err := markdown.Renderer().Render(&b, s.content, s.doc); err != nil {
		log.Print(err)
	}
	return template.HTML(b.String())
}

type astSection struct {
	title ast.Node
	nodes []ast.Node
}

func parseSlide(ndoc ast.Node, mode ParseMode) (title ast.Node, metadata []ast.Node, sections []astSection) {
	processor := func(n ast.Node) {}
	ast.Walk(ndoc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering || n.Type() == ast.TypeDocument {
			return ast.WalkContinue, nil
		}
		if n, ok := n.(*ast.Heading); ok {
			if n.Level == 1 {
				title = n
				processor = func(n ast.Node) {
					metadata = append(metadata, n)
				}
				if mode&TitlesOnly > 0 {
					return ast.WalkStop, nil
				}
			} else if n.Level <= 3 {
				sections = append(sections, astSection{
					title: n,
				})
				processor = func(n ast.Node) {
					if len(sections) > 0 {
						section := sections[len(sections)-1]
						section.nodes = append(section.nodes, n)
						sections[len(sections)-1] = section
					}
				}
			}
			return ast.WalkSkipChildren, nil
		}
		processor(n)
		return ast.WalkSkipChildren, nil
	})
	return title, metadata, sections
}
