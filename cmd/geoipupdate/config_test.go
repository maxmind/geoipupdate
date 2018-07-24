package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		Description string
		Input       string
		Output      *Config
		Err         string
	}{
		{
			Description: "Default config",
			Input: `# Please see https://dev.maxmind.com/geoip/geoipupdate/ for instructions
# on setting up geoipupdate, including information on how to download a
# pre-filled GeoIP.conf file.

# Enter your account ID and license key below. These are available from
# https://www.maxmind.com/en/my_license_key. If you are only using free
# GeoLite databases, you may leave the 0 values.
AccountID 0
LicenseKey 000000000000

# Enter the edition IDs of the databases you would like to update.
# Multiple edition IDs are separated by spaces.
EditionIDs GeoLite2-Country GeoLite2-City

# The remaining settings are OPTIONAL.

# The directory to store the database files. Defaults to DATADIR
# DatabaseDirectory DATADIR

# The server to use. Defaults to "updates.maxmind.com".
# Host updates.maxmind.com

# The proxy host name or IP address. You may optionally specify a
# port number, e.g., 127.0.0.1:8888. If no port number is specified, 1080
# will be used.
# Proxy 127.0.0.1:8888

# The user name and password to use with your proxy server.
# ProxyUserPassword username:password

# Whether to preserve modification times of files downloaded from the server.
# Defaults to "0".
# PreserveFileTimes 0

# The lock file to use. This ensures only one geoipupdate process can run at a
# time.
# Note: Once created, this lockfile is not removed from the filesystem.
# Defaults to ".geoipupdate.lock" under the DatabaseDirectory.
# LockFile DATADIR/.geoipupdate.lock
`,
			Output: &Config{
				DatabaseDirectory: "/tmp",
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LicenseKey:        "000000000000",
				LockFile:          "/tmp/.geoipupdate.lock",
				URL:               "https://updates.maxmind.com",
			},
		},
		{
			Description: "Default config, old names",
			Input: `# Please see https://dev.maxmind.com/geoip/geoipupdate/ for instructions
# on setting up geoipupdate, including information on how to download a
# pre-filled GeoIP.conf file.

# Enter your account ID and license key below. These are available from
# https://www.maxmind.com/en/my_license_key. If you are only using free
# GeoLite databases, you may leave the 0 values.
UserId 0
LicenseKey 000000000000

# Enter the edition IDs of the databases you would like to update.
# Multiple edition IDs are separated by spaces.
ProductIds GeoLite2-Country GeoLite2-City

# The remaining settings are OPTIONAL.

# The directory to store the database files. Defaults to DATADIR
# DatabaseDirectory DATADIR

# The server to use. Defaults to "updates.maxmind.com".
# Host updates.maxmind.com

# The proxy host name or IP address. You may optionally specify a
# port number, e.g., 127.0.0.1:8888. If no port number is specified, 1080
# will be used.
# Proxy 127.0.0.1:8888

# The user name and password to use with your proxy server.
# ProxyUserPassword username:password

# Whether to preserve modification times of files downloaded from the server.
# Defaults to "0".
# PreserveFileTimes 0

# The lock file to use. This ensures only one geoipupdate process can run at a
# time.
# Note: Once created, this lockfile is not removed from the filesystem.
# Defaults to ".geoipupdate.lock" under the DatabaseDirectory.
# LockFile DATADIR/.geoipupdate.lock
`,
			Output: &Config{
				DatabaseDirectory: "/tmp",
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LicenseKey:        "000000000000",
				LockFile:          "/tmp/.geoipupdate.lock",
				URL:               "https://updates.maxmind.com",
			},
		},
		{
			Description: "Everything populated",
			Input: `# Please see https://dev.maxmind.com/geoip/geoipupdate/ for instructions
# on setting up geoipupdate, including information on how to download a
# pre-filled GeoIP.conf file.

# Enter your account ID and license key below. These are available from
# https://www.maxmind.com/en/my_license_key. If you are only using free
# GeoLite databases, you may leave the 0 values.
AccountID 1234
LicenseKey abcdefghi

# Enter the edition IDs of the databases you would like to update.
# Multiple edition IDs are separated by spaces.
EditionIDs GeoLite2-Country GeoLite2-City GeoIP2-City

# The remaining settings are OPTIONAL.

# The directory to store the database files. Defaults to DATADIR
DatabaseDirectory /home

# The server to use. Defaults to "updates.maxmind.com".
Host updates.example.com

# The proxy host name or IP address. You may optionally specify a
# port number, e.g., 127.0.0.1:8888. If no port number is specified, 1080
# will be used.
Proxy 127.0.0.1:8888

# The user name and password to use with your proxy server.
ProxyUserPassword username:password

# Whether to preserve modification times of files downloaded from the server.
# Defaults to "0".
PreserveFileTimes 1

# The lock file to use. This ensures only one geoipupdate process can run at a
# time.
# Note: Once created, this lockfile is not removed from the filesystem.
# Defaults to ".geoipupdate.lock" under the DatabaseDirectory.
LockFile /usr/lock
`,
			Output: &Config{
				AccountID:         1234,
				DatabaseDirectory: "/tmp", // Argument takes precedence
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City", "GeoIP2-City"},
				LicenseKey:        "abcdefghi",
				LockFile:          "/usr/lock",
				Proxy:             "127.0.0.1:8888",
				ProxyUserPassword: "username:password",
				PreserveFileTimes: true,
				URL:               "https://updates.example.com",
			},
		},
		{
			Description: "Invalid line",
			Input: `AccountID 123
LicenseKey
# Host updates.maxmind.com
`,
			Err: "invalid format on line 2",
		},
		{
			Description: "Option is there multiple times",
			Input: `AccountID 123
AccountID 456
`,
			Err: "`AccountID' is in the config multiple times",
		},
		{
			Description: "Option is there multiple times with different names",
			Input: `AccountID 123
UserId 456
`,
			Err: "`UserId' is in the config multiple times",
		},
		{
			Description: "Invalid account ID",
			Input: `AccountID 1a
`,
			Err: `invalid account ID format: strconv.Atoi: parsing "1a": invalid syntax`,
		},
		{
			Description: "Invalid PreserveFileTimes",
			Input: `PreserveFileTimes true
`,
			Err: "`PreserveFileTimes' must be 0 or 1",
		},
		{
			Description: "Unknown option",
			Input: `AccountID 123
EditionID GeoIP2-City
`,
			Err: "unknown option on line 2",
		},
		{
			Description: "Missing required key",
			Input: `Host updates.maxmind.com
`,
			Err: "the `AccountID' option is required",
		},
		{
			Description: "Deprecated options",
			Input: `AccountID 0
LicenseKey abcd
EditionIDs GeoIP2-City
Protocol http
SkipHostnameVerification 1
SkipPeerVerification 1
`,
			Output: &Config{
				DatabaseDirectory: "/tmp",
				EditionIDs:        []string{"GeoIP2-City"},
				LicenseKey:        "abcd",
				LockFile:          "/tmp/.geoipupdate.lock",
				URL:               "https://updates.maxmind.com",
			},
		},
	}

	tempFh, err := ioutil.TempFile("", "conf-test")
	require.NoError(t, err)
	tempName := tempFh.Name()
	require.NoError(t, tempFh.Close())
	defer func() {
		_ = os.Remove(tempName)
	}()

	for _, test := range tests {
		require.NoError(t, ioutil.WriteFile(tempName, []byte(test.Input), 0600))
		config, err := NewConfig(tempName, "/tmp")
		if test.Err == "" {
			assert.NoError(t, err, test.Description)
		} else {
			assert.EqualError(t, err, test.Err, test.Description)
		}
		assert.Equal(t, test.Output, config, test.Description)
	}
}
