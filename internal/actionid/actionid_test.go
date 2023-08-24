package actionid

import (
	"fmt"
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

	const expected = "557fd19213c8599b373bb1105f6927e96be69563106a74284b5e829737a27908"
	if v := fmt.Sprintf("%x", actionID); v != expected {
		t.Errorf("GetActionID() = %v, expected = %q", v, expected)
	}
}