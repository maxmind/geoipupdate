package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Config is a parsed configuration file.
type Config struct {
	AccountID         int
	DatabaseDirectory string
	EditionIDs        []string
	LicenseKey        string
	LockFile          string
	PreserveFileTimes bool
	Proxy             string
	ProxyUserPassword string
	URL               string
}

// NewConfig parses the configuration file.
func NewConfig(
	file,
	databaseDirectory string,
) (*Config, error) {
	fh, err := os.Open(file)
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
	var host string
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
			config.DatabaseDirectory = value
		case "EditionIDs", "ProductIds":
			config.EditionIDs = strings.Fields(value)
			keysSeen["EditionIDs"] = struct{}{}
			keysSeen["ProductIds"] = struct{}{}
		case "Host":
			host = value
		case "LicenseKey":
			config.LicenseKey = value
		case "LockFile":
			config.LockFile = value
		case "PreserveFileTimes":
			if value != "0" && value != "1" {
				return nil, errors.New("`PreserveFileTimes' must be 0 or 1")
			}
			if value == "1" {
				config.PreserveFileTimes = true
			}
		case "Proxy":
			config.Proxy = value
		case "ProxyUserPassword":
			config.ProxyUserPassword = value
		case "Protocol", "SkipHostnameVerification", "SkipPeerVerification":
			// Deprecated.
		default:
			return nil, errors.Errorf("unknown option on line %d", lineNumber)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "error reading file")
	}

	requiredKeys := []string{"AccountID", "LicenseKey", "EditionIDs"}
	for _, k := range requiredKeys {
		if _, ok := keysSeen[k]; !ok {
			return nil, errors.Errorf("the `%s' option is required", k)
		}
	}

	// Set defaults.

	// Argument takes precedence.
	if databaseDirectory != "" {
		config.DatabaseDirectory = databaseDirectory
	}

	if config.DatabaseDirectory == "" {
		return nil, errors.New("no database directory specified, please set one in the config or provide it with -d")
	}

	if host == "" {
		host = "updates.maxmind.com"
	}

	if config.LockFile == "" {
		config.LockFile = filepath.Join(config.DatabaseDirectory, ".geoipupdate.lock")
	}

	config.URL = "https://" + host

	return config, nil
}
