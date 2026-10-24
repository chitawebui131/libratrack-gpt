package main

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"libratrack/web"
)

// noListFS hides directories so that http.FileServer cannot list them.
type noListFS struct {
	fsys http.FileSystem
}

func (n noListFS) Open(name string) (http.File, error) {
	// 1. f, err := n.fsys.Open(name). If err != nil, return nil, err.
	f, err := n.fsys.Open(name)
	if err != nil {
		return nil, err
	}

	// 2. info, err := f.Stat(). If err != nil: f.Close(); return nil, err.
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	// 3. If info.IsDir(): f.Close(); return nil, os.ErrNotExist.
	if info.IsDir() {
		f.Close()
		return nil, os.ErrNotExist
	}

	// 4. Return f, nil.
	return f, nil
}

// staticHandler serves GET /static/*filepath from the embedded files.
func staticHandler() (gin.HandlerFunc, error) {
	// 1. sub, err := fs.Sub(web.FS, "static"). Return nil, err on error.
	sub, err := fs.Sub(web.FS, "static")
	if err != nil {
		return nil, err
	}

	// 2. server := http.StripPrefix("/static/", http.FileServer(noListFS{fsys: http.FS(sub)})).
	server := http.StripPrefix("/static/", http.FileServer(noListFS{fsys: http.FS(sub)}))

	// 3. Return func(c *gin.Context) { server.ServeHTTP(c.Writer, c.Request) }, nil.
	return func(c *gin.Context) {
		server.ServeHTTP(c.Writer, c.Request)
	}, nil
}
