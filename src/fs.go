package main

import (
	frontend "github.com/vidar-team/Cardinal_frontend/dist"
	"net/http"
)

// FrontendFS is the file system struct of challengers' frontend.
type FrontendFS struct {
	FS http.FileSystem
}

func frontendFS() *FrontendFS {
	return &FrontendFS{FS: frontend.New()}
}

// Open: open file.
func (f *FrontendFS) Open(name string) (http.File, error) {
	return f.FS.Open(name)
}

// Exists: check if the file exist.
func (f *FrontendFS) Exists(prefix string, filepath string) bool {
	if _, err := f.FS.Open(filepath); err != nil {
		return false
	}
	return true
}
