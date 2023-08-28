package actionid

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"testing"
)

var fixturesBuildConfigs = map[string]BuildConfig{
	"darwin/arm64": {
		GOOS: "darwin",
		GOARCH: "arm64",
		RuntimeVersion: "go1.21.0",
		ToolID: map[string]string{
			"compile": "compile version go1.21.0",
		},
		ForcedGCFlags: []string{"-shared"}, // darwin + arm64 + PIE can use -shared
	},
}

var fixturesPackages = map[string]*Package{
	"internal/goarch": {
		ImportPath: "internal/goarch",
		Standard: true,
		Dir: "go/src/internal/goarch",
		GoFiles: []string{
			"goarch.go",
			"goarch_arm64.go",
			"zgoarch_arm64.go",
		},
	},
}


func Test_GetActionID(t *testing.T) {
	t.Setenv("DEBUG_HASH", "true")

	buildConfig := fixturesBuildConfigs["darwin/arm64"]
	a := Action{
		Package: fixturesPackages["internal/goarch"],
		Deps: []*Action{},
	}

	actionID, err := GetActionID(buildConfig, a)
	if err != nil {
		t.Errorf("GetActionID() error = %v", err)
	}

	fmt.Println(HashToString(actionID))

	b, err := hex.DecodeString("69b9199d34d1dfd415af8a400c2cae6d25d91b7ba1b9ec2e78acc19d4e2e7b2a")
	if err != nil {
		panic(err)
	}
	fmt.Println(HashToString(ActionID(b)))


	const expected = "557fd19213c8599b373bb1105f6927e96be69563106a74284b5e829737a27908"
	if v := fmt.Sprintf("%x", actionID); v == expected {
		t.Errorf("GetActionID() = %v, expected = %q", v, expected)
	}
}

// FindAndHash reads all of r and returns the offsets of occurrences of id.
// While reading, findAndHash also computes and returns
// a hash of the content of r, but with occurrences of id replaced by zeros.
// FindAndHash reads bufSize bytes from r at a time.
// If bufSize == 0, FindAndHash uses a reasonable default.
func FindAndHash(r io.Reader, id string, bufSize int) (matches []int64, hash [32]byte, err error) {
	if bufSize == 0 {
		bufSize = 31 * 1024 // bufSize+little will likely fit in 32 kB
	}
	if len(id) == 0 {
		return nil, [32]byte{}, fmt.Errorf("buildid.FindAndHash: no id specified")
	}
	if len(id) > bufSize {
		return nil, [32]byte{}, fmt.Errorf("buildid.FindAndHash: buffer too small")
	}
	zeros := make([]byte, len(id))
	idBytes := []byte(id)

	// For Mach-O files, we want to exclude the code signature.
	// The code signature contains hashes of the whole file (except the signature
	// itself), including the buildid. So the buildid cannot contain the signature.
	// r = excludeMachoCodeSignature(r)

	// The strategy is to read the file through buf, looking for id,
	// but we need to worry about what happens if id is broken up
	// and returned in parts by two different reads.
	// We allocate a tiny buffer (at least len(id)) and a big buffer (bufSize bytes)
	// next to each other in memory and then copy the tail of
	// one read into the tiny buffer before reading new data into the big buffer.
	// The search for id is over the entire tiny+big buffer.
	tiny := (len(id) + 127) &^ 127 // round up to 128-aligned
	buf := make([]byte, tiny+bufSize)
	h := sha256.New()
	start := tiny
	for offset := int64(0); ; {
		// The file offset maintained by the loop corresponds to &buf[tiny].
		// buf[start:tiny] is left over from previous iteration.
		// After reading n bytes into buf[tiny:], we process buf[start:tiny+n].
		n, err := io.ReadFull(r, buf[tiny:])
		if err != io.ErrUnexpectedEOF && err != io.EOF && err != nil {
			return nil, [32]byte{}, err
		}

		// Process any matches.
		for {
			i := bytes.Index(buf[start:tiny+n], idBytes)
			if i < 0 {
				break
			}
			matches = append(matches, offset+int64(start+i-tiny))
			h.Write(buf[start : start+i])
			h.Write(zeros)
			start += i + len(id)
		}
		if n < bufSize {
			// Did not fill buffer, must be at end of file.
			h.Write(buf[start : tiny+n])
			break
		}

		// Process all but final tiny bytes of buf (bufSize = len(buf)-tiny).
		// Note that start > len(buf)-tiny is possible, if the search above
		// found an id ending in the final tiny fringe. That's OK.
		if start < len(buf)-tiny {
			h.Write(buf[start : len(buf)-tiny])
			start = len(buf) - tiny
		}

		// Slide ending tiny-sized fringe to beginning of buffer.
		copy(buf[0:], buf[bufSize:])
		start -= bufSize
		offset += int64(bufSize)
	}
	h.Sum(hash[:0])
	return matches, hash, nil
}