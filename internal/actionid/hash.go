// Copyright 2017-2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the go/LICENSE file.
package actionid

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
	"sync"
)

func debugHash() bool {
	return os.Getenv("DEBUG_HASH") != ""
}

type Hash struct {
	h     hash.Hash
	name  string // for debugging
	debug bool   // for debugging
}

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
func (b BuildConfig) GetHashSalt() []byte {
	return []byte(stripExperiment(b.GetRuntimeVersion()))
}

// stripExperiment strips any GOEXPERIMENT configuration from the Go
// version string.
func stripExperiment(version string) string {
	if i := strings.Index(version, " X:"); i >= 0 {
		return version[:i]
	}
	return version
}

func newHash(buildConfig BuildConfig, name string) *Hash {
	h := &Hash{h: sha256.New(), name: name, debug: debugHash()}
	if h.debug {
		fmt.Fprintf(os.Stderr, "HASH[%s]\n", h.name)
	}

	h.Write(buildConfig.GetHashSalt())
	return h
}

func (h *Hash) Write(b []byte) (int, error) {
	if h.debug {
		fmt.Fprintf(os.Stderr, "HASH[%s] %q\n", h.name, b)
	}
	return h.h.Write(b)
}

func (h *Hash) Sum(b []byte) [HashSize]byte {
	var out [HashSize]byte
	h.h.Sum(out[:0])
	if h.debug {
		fmt.Fprintf(os.Stderr, "HASH[%s] %x\n", h.name, out)
	}
	return out
}

// HashToString converts the hash h to a string to be recorded
// in package archives and binaries as part of the build ID.
// We use the first 120 bits of the hash (5 chunks of 24 bits each) and encode
// it in base64, resulting in a 20-byte string. Because this is only used for
// detecting the need to rebuild installed files (not for lookups
// in the object file cache), 120 bits are sufficient to drive the
// probability of a false "do not need to rebuild" decision to effectively zero.
// We embed two different hashes in archives and four in binaries,
// so cutting to 20 bytes is a significant savings when build IDs are displayed.
// (20*4+3 = 83 bytes compared to 64*4+3 = 259 bytes for the
// more straightforward option of printing the entire h in base64).
func HashToString(h [32]byte) string {
	const b64 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	const chunks = 5
	var dst [chunks * 4]byte
	for i := 0; i < chunks; i++ {
		v := uint32(h[3*i])<<16 | uint32(h[3*i+1])<<8 | uint32(h[3*i+2])
		dst[4*i+0] = b64[(v>>18)&0x3F]
		dst[4*i+1] = b64[(v>>12)&0x3F]
		dst[4*i+2] = b64[(v>>6)&0x3F]
		dst[4*i+3] = b64[v&0x3F]
	}
	return string(dst[:])
}

var hashFileCache struct {
	sync.Mutex
	m map[string][HashSize]byte
}

// FileHash returns the hash of the named file.
// It caches repeated lookups for a given file,
// and the cache entry for a file can be initialized
// using SetFileHash.
// The hash used by FileHash is not the same as
// the hash used by NewHash.
func FileHash(file string) ([HashSize]byte, error) {
	hashFileCache.Lock()
	out, ok := hashFileCache.m[file]
	hashFileCache.Unlock()

	if ok {
		return out, nil
	}

	h := sha256.New()
	f, err := os.Open(file)
	if err != nil {
		return [HashSize]byte{}, err
	}
	_, err = io.Copy(h, f)
	f.Close()
	if err != nil {
		return [HashSize]byte{}, err
	}
	h.Sum(out[:0])

	SetFileHash(file, out)
	return out, nil
}

// SetFileHash sets the hash returned by FileHash for file.
func SetFileHash(file string, sum [HashSize]byte) {
	hashFileCache.Lock()
	if hashFileCache.m == nil {
		hashFileCache.m = make(map[string][HashSize]byte)
	}
	hashFileCache.m[file] = sum
	hashFileCache.Unlock()
}