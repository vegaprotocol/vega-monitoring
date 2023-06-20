package version

import (
	"runtime/debug"
)

var (
	cliVersionHash = ""
	cliVersion     = "v0.1.0"
)

func init() {
	info, _ := debug.ReadBuildInfo()

	if info == nil {
		cliVersionHash = "unknown"
		return
	}

	modified := false

	for _, v := range info.Settings {
		if v.Key == "vcs.revision" {
			cliVersionHash = v.Value
		}
		if v.Key == "vcs.modified" && v.Value == "true" {
			modified = true
		}
	}
	if modified {
		cliVersionHash += "-modified"
	}
}
