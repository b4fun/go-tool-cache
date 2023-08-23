// Copyright 2011-2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the go/LICENSE file.

package actionid

// Module describes a Go module.
type Module struct {
	Path      string `json:",omitempty"` // module path
	Version   string `json:",omitempty"` // module version
	GoVersion string `json:",omitempty"` // go version used in module
}

// Package describes a single package found in a directory.
type Package struct {
	Dir           string   `json:",omitempty"` // directory containing package sources
	ImportPath    string   `json:",omitempty"` // import path of package in dir
	Module        *Module  `json:",omitempty"` // info about package's module, if any
	Goroot        bool     `json:",omitempty"` // is this package found in the Go root?
	Standard      bool     `json:",omitempty"` // is this package part of the standard Go library?

	// Source files
	GoFiles      []string `json:",omitempty"` // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles     []string `json:",omitempty"` // .go source files that import "C"
	CFiles       []string `json:",omitempty"` // .c source files
	CXXFiles     []string `json:",omitempty"` // .cc, .cpp and .cxx source files
	FFiles       []string `json:",omitempty"` // .f, .F, .for and .f90 Fortran source files
	MFiles       []string `json:",omitempty"` // .m source files
	HFiles       []string `json:",omitempty"` // .h, .hh, .hpp and .hxx source files
	SFiles       []string `json:",omitempty"` // .s source files
	SysoFiles    []string `json:",omitempty"` // .syso system object files added to package
	SwigFiles    []string `json:",omitempty"` // .swig files
	SwigCXXFiles []string `json:",omitempty"` // .swigcxx files
	EmbedFiles   []string `json:",omitempty"` // files matched by EmbedPatterns
	PGOProfile        string               // path to PGO profile

	// Dependency information
	Deps      []string          `json:",omitempty"` // all (recursively) imported dependencies
}
