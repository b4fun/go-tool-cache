package actionid

import (
	"fmt"
	"testing"
)

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

	a := Action{
		Package: fixturesPackages["internal/goarch"],
		Deps: []*Action{},
	}

	actionID, err := GetActionID(a)
	if err != nil {
		t.Errorf("GetActionID() error = %v", err)
	}

	const expected = "foobar"
	if v := fmt.Sprintf("%x", actionID); v != expected {
		t.Errorf("GetActionID() = %v, expected = %q", v, expected)
	}
}