// +build !windows

package geoipupdate

var (
	// These match what you'd get building the C geoipupdate from source.
	DefaultConfigFile        = "/usr/local/etc/GeoIP.conf"
	DefaultDatabaseDirectory = "/usr/local/share/GeoIP"
)
