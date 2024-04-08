package geoipupdate

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/maxmind/geoipupdate/v7/internal/vars"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		Description string
		Input       string
		Env         map[string]string
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
				proxyURL:          "",
				proxyUserInfo:     "",
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
			Err: `invalid account ID format`,
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
			Err:         `invalid account ID format`,
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
		{
			Description: "Config flags override env vars override config file",
			Input:       "AccountID\t\t123\nLicenseKey\t\t456\nParallelism\t\t1\n",
			Env: map[string]string{
				"GEOIPUPDATE_DB_DIR":              "/tmp/db",
				"GEOIPUPDATE_EDITION_IDS":         "GeoLite2-Country GeoLite2-City",
				"GEOIPUPDATE_HOST":                "updates.maxmind.com",
				"GEOIPUPDATE_LICENSE_KEY":         "000000000001",
				"GEOIPUPDATE_LOCK_FILE":           "/tmp/lock",
				"GEOIPUPDATE_PARALLELISM":         "2",
				"GEOIPUPDATE_PRESERVE_FILE_TIMES": "1",
				"GEOIPUPDATE_PROXY":               "127.0.0.1:8888",
				"GEOIPUPDATE_PROXY_USER_PASSWORD": "username:password",
				"GEOIPUPDATE_RETRY_FOR":           "1m",
				"GEOIPUPDATE_VERBOSE":             "1",
			},
			Flags: []Option{WithParallelism(3)},
			Output: &Config{
				AccountID:         123,
				DatabaseDirectory: "/tmp/db",
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LicenseKey:        "000000000001",
				LockFile:          "/tmp/lock",
				Parallelism:       3,
				PreserveFileTimes: true,
				Proxy: &url.URL{
					Scheme: "http",
					User:   url.UserPassword("username", "password"),
					Host:   "127.0.0.1:8888",
				},
				RetryFor: 1 * time.Minute,
				URL:      "https://updates.maxmind.com",
				Verbose:  true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			withEnvVars(t, test.Env, func() {
				tempName := filepath.Join(t.TempDir(), "/GeoIP-test.conf")
				require.NoError(t, os.WriteFile(tempName, []byte(test.Input), 0o600))
				testFlags := append([]Option{WithConfigFile(tempName)}, test.Flags...)
				config, err := NewConfig(testFlags...)
				if test.Err == "" {
					require.NoError(t, err, test.Description)
				} else {
					require.EqualError(t, err, test.Err, test.Description)
				}
				assert.Equal(t, test.Output, config, test.Description)
			})
		})
	}
}

func TestSetConfigFromFile(t *testing.T) {
	tests := []struct {
		Description string
		Input       string
		Expected    Config
		Err         string
	}{
		{
			Description: "All config file related variables",
			Input: `AccountID 1
			DatabaseDirectory /tmp/db
			EditionIDs GeoLite2-Country GeoLite2-City
			Host updates.maxmind.com
			LicenseKey 000000000001
			LockFile /tmp/lock
			Parallelism 2
			PreserveFileTimes 1
			Proxy 127.0.0.1:8888
			ProxyUserPassword username:password
			RetryFor 1m
	`,
			Expected: Config{
				AccountID:         1,
				DatabaseDirectory: filepath.Clean("/tmp/db"),
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LicenseKey:        "000000000001",
				LockFile:          filepath.Clean("/tmp/lock"),
				Parallelism:       2,
				PreserveFileTimes: true,
				proxyURL:          "127.0.0.1:8888",
				proxyUserInfo:     "username:password",
				RetryFor:          1 * time.Minute,
				URL:               "https://updates.maxmind.com",
			},
		},
		{
			Description: "Empty config",
			Input:       "",
			Expected:    Config{},
		},
		{
			Description: "Invalid account ID",
			Input:       "AccountID 1a",
			Err:         `invalid account ID format`,
		},
		{
			Description: "Invalid PreserveFileTimes",
			Input:       "PreserveFileTimes 1a",
			Err:         "`PreserveFileTimes' must be 0 or 1",
		},
		{
			Description: "RetryFor needs a unit",
			Input:       "RetryFor 5",
			Err:         "'5' is not a valid duration",
		},
		{
			Description: "RetryFor needs to be non-negative",
			Input:       "RetryFor -5m",
			Err:         "'-5m' is not a valid duration",
		},
		{
			Description: "Parallelism should be a number",
			Input:       "Parallelism a",
			Err:         "'a' is not a valid parallelism value: strconv.Atoi: parsing \"a\": invalid syntax",
		},
		{
			Description: "Parallelism should be a positive number",
			Input:       "Parallelism 0",
			Err:         "parallelism should be greater than 0, got '0'",
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			tempName := filepath.Join(t.TempDir(), "/GeoIP-test.conf")
			require.NoError(t, os.WriteFile(tempName, []byte(test.Input), 0o600))

			var config Config

			err := setConfigFromFile(&config, tempName)
			if test.Err == "" {
				require.NoError(t, err, test.Description)
			} else {
				require.EqualError(t, err, test.Err, test.Description)
			}
			assert.Equal(t, test.Expected, config, test.Description)
		})
	}
}

func TestSetConfigFromEnv(t *testing.T) {
	tests := []struct {
		Description            string
		AccountIDFileContents  string
		LicenseKeyFileContents string
		Env                    map[string]string
		Expected               Config
		Err                    string
	}{
		{
			Description: "All config related environment variables",
			Env: map[string]string{
				"GEOIPUPDATE_ACCOUNT_ID":          "1",
				"GEOIPUPDATE_ACCOUNT_ID_FILE":     "",
				"GEOIPUPDATE_DB_DIR":              "/tmp/db",
				"GEOIPUPDATE_EDITION_IDS":         "GeoLite2-Country GeoLite2-City",
				"GEOIPUPDATE_HOST":                "updates.maxmind.com",
				"GEOIPUPDATE_LICENSE_KEY":         "000000000001",
				"GEOIPUPDATE_LICENSE_KEY_FILE":    "",
				"GEOIPUPDATE_LOCK_FILE":           "/tmp/lock",
				"GEOIPUPDATE_PARALLELISM":         "2",
				"GEOIPUPDATE_PRESERVE_FILE_TIMES": "1",
				"GEOIPUPDATE_PROXY":               "127.0.0.1:8888",
				"GEOIPUPDATE_PROXY_USER_PASSWORD": "username:password",
				"GEOIPUPDATE_RETRY_FOR":           "1m",
				"GEOIPUPDATE_VERBOSE":             "1",
			},
			Expected: Config{
				AccountID:         1,
				DatabaseDirectory: "/tmp/db",
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LicenseKey:        "000000000001",
				LockFile:          "/tmp/lock",
				Parallelism:       2,
				PreserveFileTimes: true,
				proxyURL:          "127.0.0.1:8888",
				proxyUserInfo:     "username:password",
				RetryFor:          1 * time.Minute,
				URL:               "https://updates.maxmind.com",
				Verbose:           true,
			},
		},
		{
			Description:            "ACCOUNT_ID_FILE and LICENSE_KEY_FILE override",
			AccountIDFileContents:  "2",
			LicenseKeyFileContents: "000000000002",
			Env: map[string]string{
				"GEOIPUPDATE_ACCOUNT_ID":          "1",
				"GEOIPUPDATE_ACCOUNT_ID_FILE":     filepath.Join(t.TempDir(), "accountIDFile"),
				"GEOIPUPDATE_DB_DIR":              "/tmp/db",
				"GEOIPUPDATE_EDITION_IDS":         "GeoLite2-Country GeoLite2-City",
				"GEOIPUPDATE_HOST":                "updates.maxmind.com",
				"GEOIPUPDATE_LICENSE_KEY":         "000000000001",
				"GEOIPUPDATE_LICENSE_KEY_FILE":    filepath.Join(t.TempDir(), "licenseKeyFile"),
				"GEOIPUPDATE_LOCK_FILE":           "/tmp/lock",
				"GEOIPUPDATE_PARALLELISM":         "2",
				"GEOIPUPDATE_PRESERVE_FILE_TIMES": "1",
				"GEOIPUPDATE_PROXY":               "127.0.0.1:8888",
				"GEOIPUPDATE_PROXY_USER_PASSWORD": "username:password",
				"GEOIPUPDATE_RETRY_FOR":           "1m",
				"GEOIPUPDATE_VERBOSE":             "1",
			},
			Expected: Config{
				AccountID:         2,
				DatabaseDirectory: "/tmp/db",
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LicenseKey:        "000000000002",
				LockFile:          "/tmp/lock",
				Parallelism:       2,
				PreserveFileTimes: true,
				proxyURL:          "127.0.0.1:8888",
				proxyUserInfo:     "username:password",
				RetryFor:          1 * time.Minute,
				URL:               "https://updates.maxmind.com",
				Verbose:           true,
			},
		},
		{
			Description:            "Clean up ACCOUNT_ID_FILE and LICENSE_KEY_FILE",
			AccountIDFileContents:  "\n\n2\t\n",
			LicenseKeyFileContents: "\n000000000002\t\n\n",
			Env: map[string]string{
				"GEOIPUPDATE_ACCOUNT_ID":          "1",
				"GEOIPUPDATE_ACCOUNT_ID_FILE":     filepath.Join(t.TempDir(), "accountIDFile"),
				"GEOIPUPDATE_DB_DIR":              "/tmp/db",
				"GEOIPUPDATE_EDITION_IDS":         "GeoLite2-Country GeoLite2-City",
				"GEOIPUPDATE_HOST":                "updates.maxmind.com",
				"GEOIPUPDATE_LICENSE_KEY":         "000000000001",
				"GEOIPUPDATE_LICENSE_KEY_FILE":    filepath.Join(t.TempDir(), "licenseKeyFile"),
				"GEOIPUPDATE_LOCK_FILE":           "/tmp/lock",
				"GEOIPUPDATE_PARALLELISM":         "2",
				"GEOIPUPDATE_PRESERVE_FILE_TIMES": "1",
				"GEOIPUPDATE_PROXY":               "127.0.0.1:8888",
				"GEOIPUPDATE_PROXY_USER_PASSWORD": "username:password",
				"GEOIPUPDATE_RETRY_FOR":           "1m",
				"GEOIPUPDATE_VERBOSE":             "1",
			},
			Expected: Config{
				AccountID:         2,
				DatabaseDirectory: "/tmp/db",
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LicenseKey:        "000000000002",
				LockFile:          "/tmp/lock",
				Parallelism:       2,
				PreserveFileTimes: true,
				proxyURL:          "127.0.0.1:8888",
				proxyUserInfo:     "username:password",
				RetryFor:          time.Minute,
				URL:               "https://updates.maxmind.com",
				Verbose:           true,
			},
		},
		{
			Description: "Empty config",
			Env:         map[string]string{},
			Expected:    Config{},
		},
		{
			Description: "Invalid account ID",
			Env: map[string]string{
				"GEOIPUPDATE_ACCOUNT_ID": "1a",
			},
			Err: `invalid account ID format`,
		},
		{
			Description: "Invalid PreserveFileTimes",
			Env: map[string]string{
				"GEOIPUPDATE_PRESERVE_FILE_TIMES": "1a",
			},
			Err: "`GEOIPUPDATE_PRESERVE_FILE_TIMES' must be 0 or 1",
		},
		{
			Description: "RetryFor needs a unit",
			Env: map[string]string{
				"GEOIPUPDATE_RETRY_FOR": "5",
			},
			Err: "'5' is not a valid duration",
		},
		{
			Description: "RetryFor needs to be non-negative",
			Env: map[string]string{
				"GEOIPUPDATE_RETRY_FOR": "-5m",
			},
			Err: "'-5m' is not a valid duration",
		},
		{
			Description: "Parallelism should be a number",
			Env: map[string]string{
				"GEOIPUPDATE_PARALLELISM": "a",
			},
			Err: "'a' is not a valid parallelism value: strconv.Atoi: parsing \"a\": invalid syntax",
		},
		{
			Description: "Parallelism should be a positive number",
			Env: map[string]string{
				"GEOIPUPDATE_PARALLELISM": "0",
			},
			Err: "parallelism should be greater than 0, got '0'",
		},
		{
			Description: "Invalid Verbose",
			Env: map[string]string{
				"GEOIPUPDATE_VERBOSE": "1a",
			},
			Err: "`GEOIPUPDATE_VERBOSE' must be 0 or 1",
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			accountIDFile := test.Env["GEOIPUPDATE_ACCOUNT_ID_FILE"]
			licenseKeyFile := test.Env["GEOIPUPDATE_LICENSE_KEY_FILE"]

			if test.AccountIDFileContents != "" {
				require.NoError(t, os.WriteFile(accountIDFile, []byte(test.AccountIDFileContents), 0o600))
			}

			if test.LicenseKeyFileContents != "" {
				require.NoError(t, os.WriteFile(licenseKeyFile, []byte(test.LicenseKeyFileContents), 0o600))
			}

			withEnvVars(t, test.Env, func() {
				var config Config

				err := setConfigFromEnv(&config)
				if test.Err == "" {
					require.NoError(t, err, test.Description)
				} else {
					require.EqualError(t, err, test.Err, test.Description)
				}
				assert.Equal(t, test.Expected, config, test.Description)
			})
		})
	}
}

func TestSetConfigFromFlags(t *testing.T) {
	tests := []struct {
		Description string
		Flags       []Option
		Expected    Config
		Err         string
	}{
		{
			Description: "All option flag related config set",
			Flags: []Option{
				WithDatabaseDirectory("/tmp/db"),
				WithOutput,
				WithParallelism(2),
				WithVerbose,
			},
			Expected: Config{
				DatabaseDirectory: filepath.Clean("/tmp/db"),
				Output:            true,
				Parallelism:       2,
				Verbose:           true,
			},
		},
		{
			Description: "Empty config",
			Flags:       []Option{},
			Expected:    Config{},
		},
		{
			Description: "Parallelism should be a positive number",
			Flags:       []Option{WithParallelism(-1)},
			Err:         "error applying flag to config: parallelism can't be negative, got '-1'",
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			var config Config

			err := setConfigFromFlags(&config, test.Flags...)
			if test.Err == "" {
				require.NoError(t, err, test.Description)
			} else {
				require.EqualError(t, err, test.Err, test.Description)
			}
			assert.Equal(t, test.Expected, config, test.Description)
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		Description string
		Config      Config
		Err         string
	}{
		{
			Description: "Basic config",
			Config: Config{
				AccountID:         42,
				LicenseKey:        "000000000001",
				DatabaseDirectory: "/tmp/db",
				EditionIDs:        []string{"GeoLite2-Country", "GeoLite2-City"},
				LockFile:          "/tmp/lock",
				URL:               "https://updates.maxmind.com",
				RetryFor:          5 * time.Minute,
				Parallelism:       1,
			},
			Err: "",
		},
		{
			Description: "EditionIDs required",
			Config:      Config{},
			Err:         "the `EditionIDs` option is required",
		},
		{
			Description: "AccountID required",
			Config: Config{
				EditionIDs: []string{"GeoLite2-Country", "GeoLite2-City"},
			},
			Err: "the `AccountID` option is required",
		},
		{
			Description: "LicenseKey required",
			Config: Config{
				AccountID:  42,
				EditionIDs: []string{"GeoLite2-Country", "GeoLite2-City"},
			},
			Err: "the `LicenseKey` option is required",
		},
		{
			Description: "Valid AccountID + LicenseKey combination",
			Config: Config{
				AccountID:  999999,
				LicenseKey: "000000000000",
			},
			Err: "geoipupdate requires a valid AccountID and LicenseKey combination",
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			config := test.Config
			err := validateConfig(&config)
			if test.Err == "" {
				require.NoError(t, err, test.Description)
			} else {
				require.EqualError(t, err, test.Err, test.Description)
			}
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
					require.EqualError(t, err, test.Err)
					assert.Nil(t, output)
				} else {
					require.NoError(t, err)
					assert.Equal(t, test.Output, output.String())
				}
			},
		)
	}
}

func withEnvVars(t *testing.T, newEnvVars map[string]string, f func()) {
	origEnv := os.Environ()

	for key, val := range newEnvVars {
		err := os.Setenv(key, val)
		require.NoError(t, err)
	}

	// Execute the test
	f()

	// Clean the environment
	os.Clearenv()

	// Reset the original environment variables
	for _, pair := range origEnv {
		parts := strings.SplitN(pair, "=", 2)
		err := os.Setenv(parts[0], parts[1])
		require.NoError(t, err)
	}
}
