package main

import (
	"os"
)

// These can be changed using compile time flags. e.g.,
//
// go build -ldflags "-X main.defaultConfigFile=/usr/local/etc/GeoIP.conf
// -X main.defaultDatabaseDirectory=/usr/local/share/GeoIP"
var (
	// I'm not sure these make sense. However they can be overridden at runtime
	// and in the configuration, so we have some flexibility.
	defaultConfigFile        = os.Getenv("SYSTEMDRIVE") + `\ProgramData\GeoIP.conf`
	defaultDatabaseDirectory = os.Getenv("SYSTEMDRIVE") + `\ProgramData\MaxMind\GeoIP`
)
