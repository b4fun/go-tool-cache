// Copyright 2011-2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the go/LICENSE file.

package actionid

import (
	"fmt"
	"runtime/debug"
)

const HashSize = 32

// An ActionID is a cache action key, the hash of a complete description of a
// repeatable computation (command line, environment variables,
// input file contents, executable contents).
type ActionID [HashSize]byte

// Module describes a Go module.
type Module struct {
	Path      string `json:",omitempty"` // module path
	Version   string `json:",omitempty"` // module version
	GoVersion string `json:",omitempty"` // go version used in module
}

type PackageInternal struct {
	BuildInfo *debug.BuildInfo
	PGOProfile   string   // path to PGO profile
}

// Package describes a single package found in a directory.
type Package struct {
	Dir        string  `json:",omitempty"` // directory containing package sources
	ImportPath string  `json:",omitempty"` // import path of package in dir
	Module     *Module `json:",omitempty"` // info about package's module, if any
	Goroot     bool    `json:",omitempty"` // is this package found in the Go root?
	Standard   bool    `json:",omitempty"` // is this package part of the standard Go library?

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

	// Dependency information
	Deps []string `json:",omitempty"` // all (recursively) imported dependencies

	Internal PackageInternal `json:"-"` // internal use only
}

const (
	// trim path is expected to be true as we need this to calculate persistent action id
	// for the same package in different machines using the same toolchain.
	trimPath = true

	// this flag seems to be false in normal binary building path.
	// TODO: investigate the usage in unit test binary building path.
	omitDebug = true

	// forceLibrary is not supported yet
	forceLibrary = false

	// coverMode is not supported yet
	coverMode = ""

	// fuzzInstrument is not supported yet
	fuzzInstrument = false

	// only pure go build is supported
	buildToolchainName = "gc"
)

type Action struct {
	Package *Package
	Deps []*Action
}

func GetActionID(buildConfig BuildConfig, a Action) (ActionID, error) {
	p := a.Package

	h := newHash(buildConfig, "build " + p.ImportPath)

	// Configuration independent of compiler toolchain.
	// Note: buildmode has already been accounted for in buildGcflags
	// and should not be inserted explicitly. Most buildmodes use the
	// same compiler settings and can reuse each other's results.
	// If not, the reason is already recorded in buildGcflags.
	fmt.Fprintf(h, "compile\n")

	if trimPath {
		// When -trimpath is used with a package built from the module cache,
		// its debug information refers to the module path and version
		// instead of the directory.
		if p.Module != nil {
			fmt.Fprintf(h, "module %s@%s\n", p.Module.Path, p.Module.Version)
		}
	}

	if p.Module != nil {
		fmt.Fprintf(h, "go %s\n", p.Module.GoVersion)
	}
	fmt.Fprintf(h, "goos %s goarch %s\n", buildConfig.GOOS, buildConfig.GOARCH)
	fmt.Fprintf(h, "import %q\n", p.ImportPath)
	// TODO: investigate the usage of the local flags
	const local = false
	const localPrefix = ""
	fmt.Fprintf(h, "omitdebug %v standard %v local %v prefix %q\n", omitDebug, p.Standard, local, localPrefix)
	if trimPath {
		fmt.Fprintln(h, "trimpath")
	}
	if forceLibrary {
		fmt.Fprintln(h, "forcelibrary")
	}
	if len(p.CgoFiles)+len(p.SwigFiles)+len(p.SwigCXXFiles) > 0 {
		return ActionID{}, fmt.Errorf("cgo / SWIG builds not supported yet")
	}
	if coverMode != "" {
		return ActionID{}, fmt.Errorf("cover build is not supported yet")
	}
	if fuzzInstrument{ 
		return ActionID{}, fmt.Errorf("fuzz build is not supported yet")
	}
	if p.Internal.BuildInfo != nil {
		fmt.Fprintf(h, "modinfo %s\n", p.Internal.BuildInfo.String())
	}

	if buildToolchainName == "gc" {
		compileToolID, err := buildConfig.GetToolID("compile")
		if err != nil {
			return ActionID{}, err
		}
		var internalGcFlags []string // TODO
		fmt.Fprintf(h, "compile %s %q %q\n", compileToolID, buildConfig.ForcedGCFlags, internalGcFlags)
		if len(p.SFiles) > 0 {
			return ActionID{}, fmt.Errorf("assembly files not supported yet")
		}

		// GOARM, GOMIPS, etc.
		key, val := buildConfig.GetArchEnv()
		fmt.Fprintf(h, "%s=%s\n", key, val)

		if buildConfig.CleanGOEXPERIMENT != "" {
			fmt.Fprintf(h, "GOEXPERIMENT=%s\n", buildConfig.CleanGOEXPERIMENT)
		}

		// NOTE: no support for magic environments
	}

	// Input files.
	inputFiles := stringsList(
		p.GoFiles,
		p.CgoFiles,
		p.CFiles,
		p.CXXFiles,
		p.FFiles,
		p.MFiles,
		p.HFiles,
		p.SFiles,
		p.SysoFiles,
		p.SwigFiles,
		p.SwigCXXFiles,
		p.EmbedFiles,
	)
	for _, file := range inputFiles {
		fileHash := "TODO"
		fmt.Fprintf(h, "file %s %s\n", file, fileHash)
	}
	if p.Internal.PGOProfile != "" {
		fileHash := "TODO"
		fmt.Fprintf(h, "pgofile %s\n", fileHash)
	}
	for _, a1 := range p.Deps {
		fmt.Fprintf(h, "dep %s\n", a1)
	}
	for _, a1 := range a.Deps {
		p1 := a1.Package
		if p1 != nil {
			contentID := "TODO"
			fmt.Fprintf(h, "import %s %s\n", p1.ImportPath, contentID)
		}
	}

	return ActionID(h.Sum(nil)), nil
}

func stringsList(xs ...[]string) []string {
	var n int
	for _, x := range xs {
		n += len(x)
	}
	out := make([]string, 0, n)
	for _, x := range xs {
		out = append(out, x...)
	}
	return out
}