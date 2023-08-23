// Copyright 2017-2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the go/LICENSE file.
package actionid

import (
	"crypto/sha256"
	"hash"
	"runtime"
	"strings"
)

// hashSalt is a salt string added to the beginning of every hash
// created by NewHash. Using the Go version makes sure that different
// versions of the go command (or even different Git commits during
// work on the development branch) do not address the same cache
// entries, so that a bug in one version does not affect the execution
// of other versions. This salt will result in additional ActionID files
// in the cache, but not additional copies of the large output files,
// which are still addressed by unsalted SHA256.
//
// We strip any GOEXPERIMENTs the go tool was built with from this
// version string on the assumption that they shouldn't affect go tool
// execution. This allows bootstrapping to converge faster: dist builds
// go_bootstrap without any experiments, so by stripping experiments
// go_bootstrap and the final go binary will use the same salt.
var hashSalt = []byte(stripExperiment(runtime.Version()))

// stripExperiment strips any GOEXPERIMENT configuration from the Go
// version string.
func stripExperiment(version string) string {
	if i := strings.Index(version, " X:"); i >= 0 {
		return version[:i]
	}
	return version
}

func newHash(name string) (hash.Hash) {
	h := sha256.New()
	h.Write(hashSalt)
	return h
}