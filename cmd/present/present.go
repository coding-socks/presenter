package main

import (
	"embed"
	"errors"
	"flag"
	"github.com/coding-socks/presenter"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed static/* templates/*
var base embed.FS

var (
	httpAddr    = flag.String("addr", "127.0.0.1:3777", "HTTP service address")
	contentPath = flag.String("content", ".", "base path for presentation content")
	basePath    = flag.String("base", "", "base path for slide template and static resources")
)

func main() {
	err := flag.CommandLine.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		return
	}

	var b fs.FS = base
	if *basePath != "" {
		b = os.DirFS(*basePath)
	}
	if err := initTemplates(b); err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
		return
	}
	http.Handle("/static/", http.FileServer(http.FS(base)))

	content := os.DirFS(*contentPath)
	http.Handle("/", presentationServer(content))

	scheme := "http"
	log.Printf("Open your web browser and visit %s://%s", scheme, *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Print(err)
	}
}

func parse(content fs.FS, name string, mode presenter.ParseMode) (*presenter.Doc, error) {
	f, err := content.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return presenter.ParseSlide(f, mode)
}

type dirListData struct {
	License      string
	Breadcrumbs  []breadcrumb
	Dirs, Slides dirEntrySlice
}

type breadcrumb struct {
	Link, Dir string
}

func parseBreadcrumbs(p string) []breadcrumb {
	if p == "." {
		p = ""
	}
	p = strings.Trim(p, "/")
	dirs := strings.Split(p, "/")
	breadcrumbs := make([]breadcrumb, 0, len(dirs))
	if p != "" {
		for i := range dirs {
			breadcrumbs = append(breadcrumbs, breadcrumb{
				Link: "/" + strings.Join(dirs[:i+1], "/"),
				Dir:  dirs[i],
			})
		}
	}
	return breadcrumbs
}

type dirEntry struct {
	Name, Path, Title string
}

type dirEntrySlice []dirEntry

func (s dirEntrySlice) Len() int           { return len(s) }
func (s dirEntrySlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s dirEntrySlice) Less(i, j int) bool { return s[i].Name < s[j].Name }

var (
	// contentTemplates holds the content templates.
	contentTemplates *template.Template
)

func initTemplates(base fs.FS) error {
	tmpl := template.New("").Funcs(map[string]interface{}{
		"pagenum": pageNum,
	})
	var err error
	if contentTemplates, err = tmpl.ParseFS(base, "templates/*.tmpl"); err != nil {
		return err
	}
	return nil
}

// pageNum derives a page number from a section.
func pageNum(index int, offset int) int {
	return index + offset
}

var indexName = "_index.slide"

type fsDirEntry interface {
	Name() string
	IsDir() bool
}

func presentationServer(content fs.FS) http.Handler {
	contentFileServer := http.FileServer(http.FS(content))

	var license string
	if l, err := fs.ReadFile(content, "LICENSE"); err == nil && !os.IsNotExist(err) {
		license = string(l)
	}

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Redirect(resp, req, "/static/favicon.ico", http.StatusTemporaryRedirect)
			return
		}

		p := path.Clean(strings.TrimPrefix(req.URL.Path, "/"))
		if p == "" {
			p = "."
		}
		stat, err := fs.Stat(content, p)
		if os.IsNotExist(err) {
			http.NotFound(resp, req)
			return
		}
		if err != nil {
			http.Error(resp, http.StatusText(500), 500)
			return
		}
		if stat.IsDir() {
			d := &dirListData{Breadcrumbs: parseBreadcrumbs(p)}

			files, err := fs.ReadDir(content, p)
			if err != nil {
				log.Print(err)
				http.Error(resp, http.StatusText(500), 500)
				return
			}
			for _, ff := range files {
				f := ff.(fsDirEntry)
				e := dirEntry{
					Name: f.Name(),
					Path: filepath.ToSlash(path.Join(p, f.Name())),
				}
				if strings.HasPrefix(f.Name(), ".") {
					// Ignore hidden files
					continue
				}
				if f.IsDir() {
					s, err := fs.Stat(content, path.Join(e.Path, indexName))
					if errors.Is(err, fs.ErrNotExist) {
						// Directory has no index
					} else if err != nil {
						log.Printf("cound not read stat of %s: %v", path.Join(f.Name(), indexName), err)
					} else {
						e.Name = path.Join(e.Name, indexName)
						e.Path = path.Join(e.Path, indexName)
						f = s
					}
				}
				switch {
				case f.IsDir():
					d.Dirs = append(d.Dirs, e)
				case strings.HasSuffix(f.Name(), ".slide"):
					if p, err := parse(content, e.Path, presenter.TitlesOnly); err != nil {
						log.Printf("parse(%q, present.TitlesOnly): %v", f.Name(), err)
					} else {
						e.Title = p.Title
					}
					d.Slides = append(d.Slides, e)
				}
			}
			sort.Sort(d.Dirs)
			sort.Sort(d.Slides)
			d.License = license
			contentTemplates.ExecuteTemplate(resp, "dir.tmpl", d)
		} else if path.Ext(p) == ".slide" {
			d := &presenter.Doc{}
			if p, err := parse(content, p, 0); err != nil {
				log.Printf("parse(%q, present.TitlesOnly): %v", p, err)
			} else {
				d = p
			}
			if err := contentTemplates.ExecuteTemplate(resp, "slides.tmpl", d); err != nil {
				log.Print(err)
			}
		} else {
			contentFileServer.ServeHTTP(resp, req)
		}
	})
}
