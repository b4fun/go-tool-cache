package actionid

import "runtime"

type BuildConfig struct {
	GOOS string
	GOARCH string
	CleanGOEXPERIMENT string
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