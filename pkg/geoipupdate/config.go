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
)

// Config is a parsed configuration file.
type Config struct {
	AccountID         int
	DatabaseDirectory string
	LicenseKey        string
	LockFile          string
	URL               string
	EditionIDs        []string
	Proxy             *url.URL
	PreserveFileTimes bool
	Verbose           bool
	RetryFor          time.Duration
	Parallelism       int
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

// NewConfig parses the configuration file.
// flagOptions is provided to provide optional flag overrides to the config
// file. databaseDirectory and verbose should be moved to flagOptions
// in the next major release.
func NewConfig( //nolint: gocyclo // long but breaking it up may be worse
	file,
	defaultDatabaseDirectory,
	databaseDirectory string,
	verbose bool,
	flagOptions ...Option,
) (*Config, error) {
	fh, err := os.Open(filepath.Clean(file))
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer fh.Close()

	// config defaults
	config := &Config{
		URL:               "https://updates.maxmind.com",
		DatabaseDirectory: filepath.Clean(defaultDatabaseDirectory),
		RetryFor:          5 * time.Minute,
		Parallelism:       1,
	}

	scanner := bufio.NewScanner(fh)
	lineNumber := 0
	keysSeen := map[string]struct{}{}
	var proxy, proxyUserPassword string
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid format on line %d", lineNumber)
		}
		key := fields[0]
		value := strings.Join(fields[1:], " ")

		if _, ok := keysSeen[key]; ok {
			return nil, fmt.Errorf("`%s' is in the config multiple times", key)
		}
		keysSeen[key] = struct{}{}

		switch key {
		case "AccountID", "UserId":
			accountID, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid account ID format: %w", err)
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
			config.URL = "https://" + value
		case "LicenseKey":
			config.LicenseKey = value
		case "LockFile":
			config.LockFile = filepath.Clean(value)
		case "PreserveFileTimes":
			if value != "0" && value != "1" {
				return nil, errors.New("`PreserveFileTimes' must be 0 or 1")
			}
			if value == "1" {
				config.PreserveFileTimes = true
			}
		case "Proxy":
			proxy = value
		case "ProxyUserPassword":
			proxyUserPassword = value
		case "Protocol", "SkipHostnameVerification", "SkipPeerVerification":
			// Deprecated.
		case "RetryFor":
			dur, err := time.ParseDuration(value)
			if err != nil || dur < 0 {
				return nil, fmt.Errorf("'%s' is not a valid duration", value)
			}
			config.RetryFor = dur
		case "Parallelism":
			parallelism, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("'%s' is not a valid parallelism value: %w", value, err)
			}
			if parallelism <= 0 {
				return nil, fmt.Errorf("parallelism should be greater than 0, got '%d'", parallelism)
			}
			config.Parallelism = parallelism
		default:
			return nil, fmt.Errorf("unknown option on line %d", lineNumber)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	if _, ok := keysSeen["EditionIDs"]; !ok {
		return nil, fmt.Errorf("the `EditionIDs` option is required")
	}

	if _, ok := keysSeen["AccountID"]; !ok {
		return nil, fmt.Errorf("the `AccountID` option is required")
	}

	if _, ok := keysSeen["LicenseKey"]; !ok {
		return nil, fmt.Errorf("the `LicenseKey` option is required")
	}

	// Post-processing: Argument takes precedence.
	if databaseDirectory != "" {
		config.DatabaseDirectory = filepath.Clean(databaseDirectory)
	}

	for _, option := range flagOptions {
		if err := option(config); err != nil {
			return nil, fmt.Errorf("error applying flag to config: %w", err)
		}
	}

	config.Verbose = verbose

	if config.LockFile == "" {
		config.LockFile = filepath.Join(config.DatabaseDirectory, ".geoipupdate.lock")
	}

	config.Proxy, err = parseProxy(proxy, proxyUserPassword)
	if err != nil {
		return nil, err
	}

	// We used to recommend using 999999 / 000000000000 for free downloads
	// and many people still use this combination. With a real account id
	// and license key now being required, we want to give those people a
	// sensible error message.
	if (config.AccountID == 0 || config.AccountID == 999999) && config.LicenseKey == "000000000000" {
		return nil, errors.New("geoipupdate requires a valid AccountID and LicenseKey combination")
	}

	return config, nil
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
		if scheme != "http" && scheme != "https" && scheme != "socks5" {
			return nil, fmt.Errorf("unsupported proxy type: %s", scheme)
		}
	}

	// Now that we have a scheme, we should be able to parse.
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing proxy URL: %w", err)
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
