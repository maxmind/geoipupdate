// +build !windows

package main

var (
	// These match what you'd get building the C geoipupdate from source.
	defaultConfigFile        = "/usr/local/etc/GeoIP.conf"
	defaultDatabaseDirectory = "/usr/local/share/GeoIP"
)
