package geoipupdate

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/vars"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		Description string
		Input       string
		Flags       []Option
		Output      *Config
		Err         string
	}{
		{
			Description: "Default config",
			Input: `# Please see https://dev.maxmind.com/geoip/updating-databases?lang=en for instructions
# on setting up geoipupdate, including information on how to download a
# pre-filled GeoIP.conf file.

# Enter your account ID and license key below. These are available from
# https://www.maxmind.com/en/my_license_key. If you are only using free
# GeoLite databases, you may leave the 0 values.
AccountID 42
LicenseKey 000000000001

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

# The amount of time to retry for when errors during HTTP transactions are
# encountered. It can be specified as a (possibly fractional) decimal number
# followed by a unit suffix. Valid time units are "ns", "us" (or "µs"), "ms",
# "s", "m", "h".
# Defaults to "5m" (5 minutes).
# RetryFor 5m

# The number of parallel database downloads.
# Defaults to "1".
# Parallelism 1
`,
			Output: &Config{
				AccountID:         42,
				LicenseKey:        "000000000001",
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
		},
		{
			Description: "Default config, old names",
			Input: `# Please see https://dev.maxmind.com/geoip/updating-databases?lang=en for instructions
# on setting up geoipupdate, including information on how to download a
# pre-filled GeoIP.conf file.

# Enter your account ID and license key below. These are available from
# https://www.maxmind.com/en/my_license_key. If you are only using free
# GeoLite databases, you may leave the 0 values.
UserId 42
LicenseKey 000000000001

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
				AccountID:         42,
				LicenseKey:        "000000000001",
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
		},
		{
			Description: "Everything populated",
			Input: `# Please see https://dev.maxmind.com/geoip/updating-databases?lang=en for instructions
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

RetryFor 10m

Parallelism 3
`,
			Output: &Config{
				AccountID:         1234,
				DatabaseDirectory: filepath.Clean("/home"),
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City", "GeoIP2-City"},
				LicenseKey:        "abcdefghi",
				LockFile:          filepath.Clean("/usr/lock"),
				Proxy: &url.URL{
					Scheme: "http",
					User:   url.UserPassword("username", "password"),
					Host:   "127.0.0.1:8888",
				},
				PreserveFileTimes: true,
				URL:               "https://updates.example.com",
				RetryFor:          10 * time.Minute,
				Parallelism:       3,
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
			Description: "Missing required key in options",
			Input:       ``,
			Err:         "the `EditionIDs` option is required",
		},
		{
			Description: "LicenseKey is found but AccountID is not",
			Input: `LicenseKey abcd
EditionIDs GeoIP2-City
`,
			Err: "the `AccountID` option is required",
		},
		{
			Description: "AccountID is found but LicenseKey is not",
			Input: `AccountID 123
EditionIDs GeoIP2-City`,
			Err: "the `LicenseKey` option is required",
		},
		{
			Description: "AccountID 0 with the LicenseKey 000000000000 is treated as no AccountID/LicenseKey",
			Input: `AccountID 0
LicenseKey 000000000000
EditionIDs GeoIP2-City`,
			Err: "geoipupdate requires a valid AccountID and LicenseKey combination",
		},
		{
			Description: "AccountID 999999 with the LicenseKey 000000000000 is treated as no AccountID/LicenseKey",
			Input: `AccountID 999999
LicenseKey 000000000000
EditionIDs GeoIP2-City`,
			Err: "geoipupdate requires a valid AccountID and LicenseKey combination",
		},
		{
			Description: "RetryFor needs a unit",
			Input: `AccountID 42
LicenseKey 000000000001
RetryFor 5`,
			Err: "'5' is not a valid duration",
		},
		{
			Description: "RetryFor needs to be non-negative",
			Input: `AccountID 42
LicenseKey 000000000001
RetryFor -5m`,
			Err: "'-5m' is not a valid duration",
		},
		{
			Description: "Parallelism should be a number",
			Input: `AccountID 42
LicenseKey 000000000001
Parallelism a`,
			Err: "'a' is not a valid parallelism value: strconv.Atoi: parsing \"a\": invalid syntax",
		},
		{
			Description: "Parallelism should be a positive number",
			Input: `AccountID 42
LicenseKey 000000000001
Parallelism 0`,
			Err: "parallelism should be greater than 0, got '0'",
		},
		{
			Description: "Parallelism overridden by flag",
			Input: `AccountID 999999
LicenseKey abcd
EditionIDs GeoIP2-City
Parallelism 2`,
			Flags: []Option{WithParallelism(4)},
			Output: &Config{
				AccountID:         999999,
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoIP2-City"},
				LicenseKey:        "abcd",
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       4,
			},
		},
		{
			Description: "DatabaseDirectory overridden by flag",
			Input: `AccountID 999999
LicenseKey abcd
EditionIDs GeoIP2-City`,
			Flags: []Option{WithDatabaseDirectory("/tmp")},
			Output: &Config{
				AccountID:         999999,
				DatabaseDirectory: filepath.Clean("/tmp"),
				EditionIDs:        []string{"GeoIP2-City"},
				LicenseKey:        "abcd",
				LockFile:          filepath.Clean("/tmp/.geoipupdate.lock"),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
		},
		{
			Description: "AccountID 999999 with a non-000000000000 LicenseKey is treated normally",
			Input: `AccountID 999999
LicenseKey abcd
EditionIDs GeoIP2-City`,
			Output: &Config{
				AccountID:         999999,
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoIP2-City"},
				LicenseKey:        "abcd",
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
		},
		{
			Description: "Deprecated options",
			Input: `AccountID 123
LicenseKey abcd
EditionIDs GeoIP2-City
Protocol http
SkipHostnameVerification 1
SkipPeerVerification 1
`,
			Output: &Config{
				AccountID:         123,
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoIP2-City"},
				LicenseKey:        "abcd",
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
		},
		{
			Description: "CRLF line ending works",
			Input:       "AccountID 123\r\nLicenseKey 123\r\nEditionIDs GeoIP2-City\r\n",
			Output: &Config{
				AccountID:         123,
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoIP2-City"},
				LicenseKey:        "123",
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
		},
		{
			Description: "CR line ending does not work",
			Input:       "AccountID 0\rLicenseKey 123\rEditionIDs GeoIP2-City\r",
			//nolint: lll
			Err: `invalid account ID format: strconv.Atoi: parsing "0 LicenseKey 123 EditionIDs GeoIP2-City": invalid syntax`,
		},
		{
			Description: "Multiple spaces between option and value works",
			Input: `AccountID  123
LicenseKey  456
EditionIDs    GeoLite2-City      GeoLite2-Country
`,
			Output: &Config{
				AccountID:         123,
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoLite2-City", "GeoLite2-Country"},
				LicenseKey:        "456",
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
		},
		{
			Description: "Tabs between options and values works",

			Input: "AccountID\t123\nLicenseKey\t\t456\nEditionIDs\t\t\tGeoLite2-City\t\t\t\tGeoLite2-Country\t\t\t\t\n",
			Output: &Config{
				AccountID:         123,
				DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
				EditionIDs:        []string{"GeoLite2-City", "GeoLite2-Country"},
				LicenseKey:        "456",
				LockFile:          filepath.Clean(filepath.Join(vars.DefaultDatabaseDirectory, ".geoipupdate.lock")),
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
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
		t.Run(test.Description, func(t *testing.T) {
			require.NoError(t, ioutil.WriteFile(tempName, []byte(test.Input), 0o600))
			config, err := NewConfig(tempName, test.Flags...)
			if test.Err == "" {
				assert.NoError(t, err, test.Description)
			} else {
				assert.EqualError(t, err, test.Err, test.Description)
			}
			assert.Equal(t, test.Output, config, test.Description)
		})
	}
}

func TestParseProxy(t *testing.T) {
	tests := []struct {
		Proxy        string
		UserPassword string
		Output       string
		Err          string
	}{
		{
			Proxy:  "127.0.0.1",
			Output: "http://127.0.0.1:1080",
		},
		{
			Proxy:  "127.0.0.1:8888",
			Output: "http://127.0.0.1:8888",
		},
		{
			Proxy:  "http://127.0.0.1:8888",
			Output: "http://127.0.0.1:8888",
		},
		{
			Proxy:  "socks5://127.0.0.1",
			Output: "socks5://127.0.0.1:1080",
		},
		{
			Proxy:  "socks5://127.0.0.1:8888",
			Output: "socks5://127.0.0.1:8888",
		},
		{
			Proxy:  "Garbage",
			Output: "http://Garbage:1080",
		},
		{
			Proxy: "ftp://127.0.0.1",
			Err:   "unsupported proxy type: ftp",
		},
		{
			Proxy: "ftp://127.0.0.1:8888",
			Err:   "unsupported proxy type: ftp",
		},
		{
			Proxy:  "login:password@127.0.0.1",
			Output: "http://login:password@127.0.0.1:1080",
		},
		{
			Proxy:        "login:password@127.0.0.1",
			UserPassword: "something:else",
			Output:       "http://login:password@127.0.0.1:1080",
		},
		{
			Proxy:        "127.0.0.1",
			UserPassword: "something:else",
			Output:       "http://something:else@127.0.0.1:1080",
		},
		{
			Proxy:        "127.0.0.1:8888",
			UserPassword: "something:else",
			Output:       "http://something:else@127.0.0.1:8888",
		},
		{
			Proxy:        "user:password@127.0.0.1:8888",
			UserPassword: "user2:password2",
			Output:       "http://user:password@127.0.0.1:8888",
		},
		{
			Proxy:        "http://user:password@127.0.0.1:8888",
			UserPassword: "user2:password2",
			Output:       "http://user:password@127.0.0.1:8888",
		},
	}

	for _, test := range tests {
		t.Run(
			fmt.Sprintf("%s - %s", test.Proxy, test.UserPassword),
			func(t *testing.T) {
				output, err := parseProxy(test.Proxy, test.UserPassword)
				if test.Err != "" {
					assert.EqualError(t, err, test.Err)
					assert.Nil(t, output)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, test.Output, output.String())
				}
			},
		)
	}
}
