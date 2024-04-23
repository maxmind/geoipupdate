package geoipupdate

import (
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/maxmind/geoipupdate/v7/internal/vars"
)

const schemeHTTPS = "https"

// Config is a parsed configuration file.
type Config struct {
	// AccountID is the account ID.
	AccountID int
	// confFile is the path to any configuration file used when
	// potentially populating Config fields.
	configFile string
	// DatabaseDirectory is where database files are going to be
	// stored.
	DatabaseDirectory string
	// EditionIDs are the database editions to be updated.
	EditionIDs []string
	// LicenseKey is the license attached to the account.
	LicenseKey string
	// LockFile is the path of a lock file that ensures that only one
	// geoipupdate process can run at a time.
	LockFile string
	// PreserveFileTimes sets whether database modification times
	// are preserved across downloads.
	PreserveFileTimes bool
	// Parallelism defines the number of concurrent downloads that
	// can be triggered at the same time. It defaults to 1, which
	// wouldn't change the existing behavior of downloading files
	// sequentially.
	Parallelism int
	// Proxy is host name or IP address of a proxy server.
	Proxy *url.URL
	// proxyURL is the host value of Proxy
	proxyURL string
	// proxyUserInfo is the userinfo value of Proxy
	proxyUserInfo string
	// RetryFor is the retry timeout for HTTP requests. It defaults
	// to 5 minutes.
	RetryFor time.Duration
	// URL points to maxmind servers.
	URL string
	// Verbose turns on debug statements.
	Verbose bool
	// Output turns on sending the download/update result to stdout as JSON.
	Output bool
}

// Option is a function type that modifies a configuration object.
// It is used to define functions that override a config with
// values set as command line arguments.
type Option func(f *Config) error

// WithParallelism returns an Option that sets the Parallelism
// value of a config.
func WithParallelism(i int) Option {
	return func(c *Config) error {
		if i < 0 {
			return fmt.Errorf("parallelism can't be negative, got '%d'", i)
		}
		if i > 0 {
			c.Parallelism = i
		}
		return nil
	}
}

// WithDatabaseDirectory returns an Option that sets the DatabaseDirectory
// value of a config.
func WithDatabaseDirectory(dir string) Option {
	return func(c *Config) error {
		if dir != "" {
			c.DatabaseDirectory = filepath.Clean(dir)
		}
		return nil
	}
}

// WithVerbose enable verbose output for the config.
func WithVerbose(c *Config) error {
	c.Verbose = true
	return nil
}

// WithOutput enables JSON output for the config.
func WithOutput(c *Config) error {
	c.Output = true
	return nil
}

// WithConfigFile returns an Option that sets the configuration
// file to be used.
func WithConfigFile(file string) Option {
	return func(c *Config) error {
		if file != "" {
			c.configFile = filepath.Clean(file)
		}
		return nil
	}
}

// NewConfig creates a new configuration and populates it based on an optional
// config file pointed to by an option set with WithConfigFile, then by various
// environment variables, and then finally by flag overrides provided by
// flagOptions. Values from the later override the former.
func NewConfig(
	flagOptions ...Option,
) (*Config, error) {
	// config defaults
	config := &Config{
		URL:               "https://updates.maxmind.com",
		DatabaseDirectory: filepath.Clean(vars.DefaultDatabaseDirectory),
		RetryFor:          5 * time.Minute,
		Parallelism:       1,
	}

	// Potentially populate config.configFilePath. We will rerun this function
	// again later to ensure the flag values override env variables.
	err := setConfigFromFlags(config, flagOptions...)
	if err != nil {
		return nil, err
	}

	// Override config with values from the config file.
	if confFile := config.configFile; confFile != "" {
		err = setConfigFromFile(config, confFile)
		if err != nil {
			return nil, err
		}
	}

	// Override config with values from environment variables.
	err = setConfigFromEnv(config)
	if err != nil {
		return nil, err
	}

	// Override config with values from option flags.
	err = setConfigFromFlags(config, flagOptions...)
	if err != nil {
		return nil, err
	}

	// Set config values that depend on other config values. For instance
	// proxyURL may have been set by the default config, and proxyUserInfo
	// by config file. Both of these values need to be combined to create
	// the public Proxy field that is a *url.URL.

	config.Proxy, err = parseProxy(config.proxyURL, config.proxyUserInfo)
	if err != nil {
		return nil, err
	}

	if config.LockFile == "" {
		config.LockFile = filepath.Join(config.DatabaseDirectory, ".geoipupdate.lock")
	}

	// Validate config values now that all config sources have been considered and
	// any value that may need to be created from other values has been set.

	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	// Reset values that were only needed to communicate information between
	// config overrides.

	config.configFile = ""
	config.proxyURL = ""
	config.proxyUserInfo = ""

	return config, nil
}

// setConfigFromFile sets Config fields based on the configuration file.
func setConfigFromFile(config *Config, path string) error {
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}

	defer fh.Close()

	scanner := bufio.NewScanner(fh)
	lineNumber := 0
	keysSeen := map[string]struct{}{}
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return fmt.Errorf("invalid format on line %d", lineNumber)
		}
		key := fields[0]
		value := strings.Join(fields[1:], " ")

		if _, ok := keysSeen[key]; ok {
			return fmt.Errorf("`%s' is in the config multiple times", key)
		}
		keysSeen[key] = struct{}{}

		switch key {
		case "AccountID", "UserId":
			accountID, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("invalid account ID format")
			}
			config.AccountID = accountID
			keysSeen["AccountID"] = struct{}{}
			keysSeen["UserId"] = struct{}{}
		case "DatabaseDirectory":
			config.DatabaseDirectory = filepath.Clean(value)
		case "EditionIDs", "ProductIds":
			config.EditionIDs = strings.Fields(value)
			keysSeen["EditionIDs"] = struct{}{}
			keysSeen["ProductIds"] = struct{}{}
		case "Host":
			u, err := url.Parse(value)
			if err != nil {
				return fmt.Errorf("failed to parse Host: %w", err)
			}
			if u.Scheme == "" {
				u.Scheme = schemeHTTPS
			}
			config.URL = u.String()
		case "LicenseKey":
			config.LicenseKey = value
		case "LockFile":
			config.LockFile = filepath.Clean(value)
		case "PreserveFileTimes":
			if value != "0" && value != "1" {
				return errors.New("`PreserveFileTimes' must be 0 or 1")
			}
			config.PreserveFileTimes = value == "1"
		case "Proxy":
			config.proxyURL = value
		case "ProxyUserPassword":
			config.proxyUserInfo = value
		case "Protocol", "SkipHostnameVerification", "SkipPeerVerification":
			// Deprecated.
		case "RetryFor":
			dur, err := time.ParseDuration(value)
			if err != nil || dur < 0 {
				return fmt.Errorf("'%s' is not a valid duration", value)
			}
			config.RetryFor = dur
		case "Parallelism":
			parallelism, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("'%s' is not a valid parallelism value: %w", value, err)
			}
			if parallelism <= 0 {
				return fmt.Errorf("parallelism should be greater than 0, got '%d'", parallelism)
			}
			config.Parallelism = parallelism
		default:
			return fmt.Errorf("unknown option on line %d", lineNumber)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	return nil
}

// setConfigFromEnv sets Config fields based on environment variables.
func setConfigFromEnv(config *Config) error {
	if value, ok := os.LookupEnv("GEOIPUPDATE_ACCOUNT_ID"); ok {
		var err error
		config.AccountID, err = strconv.Atoi(value)
		if err != nil {
			return errors.New("invalid account ID format")
		}
	}

	if value := os.Getenv("GEOIPUPDATE_ACCOUNT_ID_FILE"); value != "" {
		var err error

		accountID, err := os.ReadFile(filepath.Clean(value))
		if err != nil {
			return fmt.Errorf("failed to open GEOIPUPDATE_ACCOUNT_ID_FILE: %w", err)
		}

		config.AccountID, err = strconv.Atoi(strings.TrimSpace(string(accountID)))
		if err != nil {
			return errors.New("invalid account ID format")
		}
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_DB_DIR"); ok {
		config.DatabaseDirectory = value
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_EDITION_IDS"); ok {
		config.EditionIDs = strings.Fields(value)
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_HOST"); ok {
		u, err := url.Parse(value)
		if err != nil {
			return fmt.Errorf("failed to parse GEOIPUPDATE_HOST: %w", err)
		}
		if u.Scheme == "" {
			u.Scheme = schemeHTTPS
		}
		config.URL = u.String()
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_LICENSE_KEY"); ok {
		config.LicenseKey = value
	}

	if value := os.Getenv("GEOIPUPDATE_LICENSE_KEY_FILE"); value != "" {
		var err error

		licenseKey, err := os.ReadFile(filepath.Clean(value))
		if err != nil {
			return fmt.Errorf("failed to open GEOIPUPDATE_LICENSE_KEY_FILE: %w", err)
		}

		config.LicenseKey = strings.TrimSpace(string(licenseKey))
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_LOCK_FILE"); ok {
		config.LockFile = value
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_PARALLELISM"); ok {
		parallelism, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid parallelism value: %w", value, err)
		}
		if parallelism <= 0 {
			return fmt.Errorf("parallelism should be greater than 0, got '%d'", parallelism)
		}
		config.Parallelism = parallelism
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_PRESERVE_FILE_TIMES"); ok {
		if value != "0" && value != "1" {
			return errors.New("`GEOIPUPDATE_PRESERVE_FILE_TIMES' must be 0 or 1")
		}
		config.PreserveFileTimes = value == "1"
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_PROXY"); ok {
		config.proxyURL = value
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_PROXY_USER_PASSWORD"); ok {
		config.proxyUserInfo = value
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_RETRY_FOR"); ok {
		dur, err := time.ParseDuration(value)
		if err != nil || dur < 0 {
			return fmt.Errorf("'%s' is not a valid duration", value)
		}
		config.RetryFor = dur
	}

	if value, ok := os.LookupEnv("GEOIPUPDATE_VERBOSE"); ok {
		if value != "0" && value != "1" {
			return errors.New("`GEOIPUPDATE_VERBOSE' must be 0 or 1")
		}
		config.Verbose = value == "1"
	}

	return nil
}

// setConfigFromFlags sets Config fields based on option flags.
func setConfigFromFlags(config *Config, flagOptions ...Option) error {
	for _, option := range flagOptions {
		if err := option(config); err != nil {
			return fmt.Errorf("error applying flag to config: %w", err)
		}
	}
	return nil
}

func validateConfig(config *Config) error {
	// We used to recommend using 999999 / 000000000000 for free downloads
	// and many people still use this combination. With a real account id
	// and license key now being required, we want to give those people a
	// sensible error message.
	if (config.AccountID == 0 || config.AccountID == 999999) && config.LicenseKey == "000000000000" {
		return errors.New("geoipupdate requires a valid AccountID and LicenseKey combination")
	}

	if len(config.EditionIDs) == 0 {
		return errors.New("the `EditionIDs` option is required")
	}

	if config.AccountID == 0 {
		return errors.New("the `AccountID` option is required")
	}

	if config.LicenseKey == "" {
		return errors.New("the `LicenseKey` option is required")
	}

	return nil
}

var schemeRE = regexp.MustCompile(`(?i)\A([a-z][a-z0-9+\-.]*)://`)

func parseProxy(
	proxy,
	proxyUserPassword string,
) (*url.URL, error) {
	if proxy == "" {
		return nil, nil
	}
	proxyURL := proxy

	// If no scheme is provided, use http.
	matches := schemeRE.FindStringSubmatch(proxyURL)
	if matches == nil {
		proxyURL = "http://" + proxyURL
	} else {
		scheme := strings.ToLower(matches[1])
		// The http package only supports http, https, and socks5.
		if scheme != "http" && scheme != schemeHTTPS && scheme != "socks5" {
			return nil, fmt.Errorf("unsupported proxy type: %s", scheme)
		}
	}

	// Now that we have a scheme, we should be able to parse.
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parsing proxy URL: %w", err)
	}

	if !strings.Contains(u.Host, ":") {
		u.Host += ":1080" // The 1080 default historically came from cURL.
	}

	// Historically if the Proxy option had a username and password they would
	// override any specified in the ProxyUserPassword option. Continue that.
	if u.User != nil {
		return u, nil
	}

	if proxyUserPassword == "" {
		return u, nil
	}

	userPassword := strings.SplitN(proxyUserPassword, ":", 2)
	if len(userPassword) != 2 {
		return nil, errors.New("proxy user/password is malformed")
	}
	u.User = url.UserPassword(userPassword[0], userPassword[1])

	return u, nil
}
