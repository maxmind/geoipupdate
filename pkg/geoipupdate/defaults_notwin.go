// +build !windows

package geoipupdate

var (
	// These match what you'd get building the C geoipupdate from source.

	// DefaultConfigFile is the default location that GeoipUpdate will look for the *.conf file
	DefaultConfigFile = "/usr/local/etc/GeoIP.conf"
	// DefaultDatabaseDirectory is the default directory that will be used for saving to the local file system
	DefaultDatabaseDirectory = "/usr/local/share/GeoIP"
)
