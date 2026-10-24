package handler

import (
	"html/template"
	"io/fs"
	"path"
	"strings"
	"time"
)

type Renderer struct {
	pages map[string]*template.Template
	loc   *time.Location
}

// NewRenderer builds one template set per page: layout + all partials + that page.
func NewRenderer(fsys fs.FS, loc *time.Location) (*Renderer, error) {
	// 1. pageFiles, err := fs.Glob(fsys, "templates/pages/*.html"). Return the error if any.
	pageFiles, err := fs.Glob(fsys, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}
	// 2. partials, err := fs.Glob(fsys, "templates/partials/*.html"). Return the error if any.
	partials, err := fs.Glob(fsys, "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	// 3. pages := map[string]*template.Template{}.
	pages := map[string]*template.Template{}
	// 4. For every pageFile:
	for _, pageFile := range pageFiles {
		//      files := append([]string{"templates/layout.html"}, partials...); files = append(files, pageFile).
		files := append([]string{"templates/layout.html"}, partials...)
		files = append(files, pageFile)
		//      name := strings.TrimSuffix(path.Base(pageFile), ".html").
		name := strings.TrimSuffix(path.Base(pageFile), ".html")
		//      t, err := template.New(name).Funcs(templateFuncs(loc)).ParseFS(fsys, files...). Return nil and the error if it fails.
		t, err := template.New(name).Funcs(templateFuncs(loc)).ParseFS(fsys, files...)
		if err != nil {
			return nil, err
		}
		//      pages[name] = t.
		pages[name] = t
	}
	// 5. Return &Renderer{pages: pages, loc: loc}, nil.
	return &Renderer{pages: pages, loc: loc}, nil
}
