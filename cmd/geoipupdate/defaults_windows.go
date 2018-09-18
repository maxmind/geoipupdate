package main

import (
	"os"
)

var (
	// I'm not sure these make sense. However they can be overridden at runtime
	// and in the configuration, so we have some flexibility.
	defaultConfigFile        = os.Getenv("SYSTEMDRIVE") + `\ProgramData\GeoIP.conf`
	defaultDatabaseDirectory = os.Getenv("SYSTEMDRIVE") + `\ProgramData\MaxMind\GeoIP`
)
