package geoipupdate

import (
	"os"
)

var (
	// I'm not sure these make sense. However they can be overridden at runtime
	// and in the configuration, so we have some flexibility.
	DefaultConfigFile        = os.Getenv("SYSTEMDRIVE") + `\ProgramData\MaxMind\GeoIPUpdate\GeoIP.conf`
	DefaultDatabaseDirectory = os.Getenv("SYSTEMDRIVE") + `\ProgramData\MaxMind\GeoIPUpdate\GeoIP`
)
