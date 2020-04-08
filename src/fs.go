package main

import (
	frontend "github.com/vidar-team/Cardinal_frontend/dist"
	manager "github.com/vidar-team/Cardinal_manager_frontend/dist"
	"net/http"
	"strings"
)

// FrontendFS is the file system struct of challengers' frontend.
type FrontendFS struct {
	frontendFS http.FileSystem
	managerFS  http.FileSystem
}

func frontendFS() *FrontendFS {
	return &FrontendFS{
		frontendFS: frontend.New(),
		managerFS:  manager.New(),
	}
}

// Open: open file.
func (f *FrontendFS) Open(name string) (http.File, error) {
	if strings.HasPrefix(name, "/manager") {
		return f.managerFS.Open(name)
	}
	return f.frontendFS.Open(name)
}

// Exists: check if the file exist.
func (f *FrontendFS) Exists(prefix string, filePath string) bool {
	if strings.HasPrefix(filePath, "/manager") {
		if _, err := f.managerFS.Open(filePath); err != nil {
			return false
		}
		return true
	}
	if _, err := f.frontendFS.Open(filePath); err != nil {
		return false
	}
	return true
}
