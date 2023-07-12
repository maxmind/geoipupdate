//go:build go1.12
// +build go1.12

package main

import "runtime/debug"

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Getting the build info failed, e.g., it was disabled on build.
		return
	}
	if version == unknownVersion {
		// This will set the version on go install ...
		version = info.Main.Version
	}

	var rev, time, arch, os string
	dirty := false
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			rev = kv.Value
		case "vcs.time":
			time = kv.Value
		case "vcs.modified":
			dirty = kv.Value == "true"
		case "GOARCH":
			arch = kv.Value
		case "GOOS":
			os = kv.Value
		}
	}

	bi := ""

	if len(rev) >= 8 {
		bi += rev[:8]
		if dirty {
			bi += "-modified"
		}
		bi += ", "
	}
	if time != "" {
		bi += time + ", "
	}
	bi += os + "-" + arch
	version += " (" + bi + ")"
}
