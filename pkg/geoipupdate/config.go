package geoipupdate

import (
	"bufio"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
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
}

// NewConfig parses the configuration file.
func NewConfig( // nolint: gocyclo
	file,
	defaultDatabaseDirectory,
	databaseDirectory string,
	verbose bool,
) (*Config, error) {
	fh, err := os.Open(filepath.Clean(file))
	if err != nil {
		return nil, errors.Wrap(err, "error opening file")
	}
	defer func() {
		if err := fh.Close(); err != nil {
			log.Fatalf("Error closing config file: %+v", errors.Wrap(err, "closing file"))
		}
	}()

	config := &Config{}
	scanner := bufio.NewScanner(fh)
	lineNumber := 0
	keysSeen := map[string]struct{}{}
	var host, proxy, proxyUserPassword string
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			return nil, errors.Errorf("invalid format on line %d", lineNumber)
		}
		key := fields[0]
		value := strings.Join(fields[1:], " ")

		if _, ok := keysSeen[key]; ok {
			return nil, errors.Errorf("`%s' is in the config multiple times", key)
		}
		keysSeen[key] = struct{}{}

		switch key {
		case "AccountID", "UserId":
			accountID, err := strconv.Atoi(value)
			if err != nil {
				return nil, errors.Wrap(err, "invalid account ID format")
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
			host = value
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
		default:
			return nil, errors.Errorf("unknown option on line %d", lineNumber)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "error reading file")
	}

	if _, ok := keysSeen["EditionIDs"]; !ok {
		return nil, errors.Errorf("the `EditionIDs` option is required")
	}

	{
		_, LicenseKeySeen := keysSeen["LicenseKey"]
		_, AccountIDSeen := keysSeen["AccountID"]

		if LicenseKeySeen && !AccountIDSeen {
			return nil, errors.Errorf("the `AccountID` option is required if the `LicenseKey` option is set")
		}

		if AccountIDSeen && !LicenseKeySeen {
			return nil, errors.Errorf("the `LicenseKey` option is required if the `AccountID` option is set")
		}

		if AccountIDSeen && config.AccountID == 0 && LicenseKeySeen && config.LicenseKey != "000000000000" {
			return nil, errors.New("setting an `AccountID` option of 0 with a `LicenseKey` option other than 000000000000 is disallowed")
		}
	}

	// Set defaults & post-process.

	// Argument takes precedence.
	if databaseDirectory != "" {
		config.DatabaseDirectory = filepath.Clean(databaseDirectory)
	}

	if config.DatabaseDirectory == "" {
		config.DatabaseDirectory = filepath.Clean(defaultDatabaseDirectory)
	}

	config.Verbose = verbose

	if host == "" {
		host = "updates.maxmind.com"
	}

	if config.LockFile == "" {
		config.LockFile = filepath.Join(config.DatabaseDirectory, ".geoipupdate.lock")
	}

	config.URL = "https://" + host

	config.Proxy, err = parseProxy(proxy, proxyUserPassword)
	if err != nil {
		return nil, err
	}

	// We used to recommend using 999999 / 000000000000 for free downloads and
	// many people still use this combination. We need to check for the
	// 000000000000 license key to ensure that a real AccountID of 999999 will
	// work in the future.
	if (config.AccountID == 0 || config.AccountID == 999999) && config.LicenseKey == "000000000000" {
		config.AccountID = 0
		config.LicenseKey = ""
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

	// If no scheme is provided, use http.
	matches := schemeRE.FindStringSubmatch(proxy)
	if matches == nil {
		proxy = "http://" + proxy
	} else {
		scheme := strings.ToLower(matches[1])
		// The http package only supports http and socks5.
		if scheme != "http" && scheme != "socks5" {
			return nil, errors.Errorf("unsupported proxy type: %s", scheme)
		}
	}

	// Now that we have a scheme, we should be able to parse.
	u, err := url.Parse(proxy)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing proxy URL")
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
