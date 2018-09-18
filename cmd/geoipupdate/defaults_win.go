// +build windows

package main

// These can be changed using compile time flags. e.g.,
//
// go build -ldflags "-X main.defaultConfigFile=/usr/local/etc/GeoIP.conf
// -X main.defaultDatabaseDirectory=/usr/local/share/GeoIP"
var (
	// I'm not sure these make sense. However they can be overridden at runtime
	// and in the configuration, so we have some flexibility.
	defaultConfigFile        = `C:\ProgramData\GeoIP.conf`
	defaultDatabaseDirectory = `C:\ProgramData\GeoIP`
)
