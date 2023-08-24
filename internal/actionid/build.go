package actionid

import (
	"fmt"
	"runtime"
)

type BuildConfig struct {
	GOOS string
	GOARCH string
	CleanGOEXPERIMENT string
	RuntimeVersion string

	ToolID map[string]string
	ForcedGCFlags []string
}

func resolveBuildConfig() (BuildConfig, error) {
	return BuildConfig{
		GOOS: runtime.GOOS,
		GOARCH: runtime.GOARCH,
	}, nil
}

func (b BuildConfig) GetArchEnv() (string, string) {
	switch b.GOARCH {
	case "arm":
		return "GOARM", "7" // TODO: support other GOARM values
	case "amd64":
		return "GOAMD64", "v1" // TODO: support other GOAMD64 values
		// TODO: support other GOARCH values
	default:
		return "", ""
	}
}

func (b BuildConfig) GetRuntimeVersion() string {
	if b.RuntimeVersion != "" {
		return b.RuntimeVersion
	}

	return runtime.Version()
}

func (b BuildConfig) GetToolID(name string) (string, error) {
	if b.ToolID == nil {
		return "", fmt.Errorf("tool id %q not found", name)
	}

	v, exists := b.ToolID[name]
	if !exists {
		return "", fmt.Errorf("tool id %q not found", name)
	}

	return v, nil
}