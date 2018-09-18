// +build !windows

package main

// These can be changed using compile time flags. e.g.,
//
// go build -ldflags "-X main.defaultConfigFile=/usr/local/etc/GeoIP.conf
// -X main.defaultDatabaseDirectory=/usr/local/share/GeoIP"
var (
	// These match what you'd get building the C geoipupdate from source.
	defaultConfigFile        = "/usr/local/etc/GeoIP.conf"
	defaultDatabaseDirectory = "/usr/local/share/GeoIP"
)