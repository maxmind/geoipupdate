// Package download provides a library for checking/downloading/updating mmdb files.
package download

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const (
	// Extension is the typical extension used for database files.
	Extension = ".mmdb"

	// zeroMD5 is the default value provided as an MD5 hash for a non-existent
	// database.
	zeroMD5 = "00000000000000000000000000000000"
)

// Downloader represents common methods required to implement the functionality
// required to download/update mmdb files.
type Downloader interface {
	GetOutdatedEditions(ctx context.Context) ([]Metadata, error)
	DownloadEdition(ctx context.Context, edition Metadata) error
	MakeOutput() ([]byte, error)
}

// Download exposes methods needed to check for and perform update to a set of mmdb editions.
type Download struct {
	// accountID is the requester's account ID.
	accountID int
	// client is an http client responsible of fetching database updates.
	client *http.Client
	// databaseDir is the database download path.
	databaseDir string
	// editionIDs is the list of editions to be updated.
	editionIDs []string
	// licenseKey is the requester's license key.
	licenseKey string
	// oldEditionsHash holds the hashes of the previously downloaded mmdb editions.
	oldEditionsHash map[string]string
	// metadata holds the metadata pulled for each edition.
	metadata []Metadata
	// preserveFileTimes sets whether database modification times are preserved across downloads.
	preserveFileTimes bool
	// url points to maxmind servers.
	url string

	now     func() time.Time
	verbose bool
}

// New initializes a new Downloader struct.
func New(
	accountID int,
	licenseKey string,
	serverURL string,
	proxy *url.URL,
	databaseDir string,
	preserveFileTimes bool,
	editionIDs []string,
	verbose bool,
) (*Download, error) {
	transport := http.DefaultTransport
	if proxy != nil {
		proxyFunc := http.ProxyURL(proxy)
		transport.(*http.Transport).Proxy = proxyFunc
	}

	d := Download{
		accountID:         accountID,
		client:            &http.Client{Transport: transport},
		databaseDir:       databaseDir,
		editionIDs:        editionIDs,
		licenseKey:        licenseKey,
		oldEditionsHash:   map[string]string{},
		preserveFileTimes: preserveFileTimes,
		url:               serverURL,
		now:               time.Now,
		verbose:           verbose,
	}

	for _, e := range editionIDs {
		hash, err := d.getHash(e)
		if err != nil {
			return nil, fmt.Errorf("getting existing %q database hash: %w", e, err)
		}
		d.oldEditionsHash[e] = hash
	}

	return &d, nil
}

// getHash returns the hash of a certain database file.
func (d *Download) getHash(editionID string) (string, error) {
	databaseFilePath := d.getFilePath(editionID)
	//nolint:gosec // we really need to read this file.
	database, err := os.Open(databaseFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if d.verbose {
				log.Print("Database does not exist, returning zeroed hash")
			}
			return zeroMD5, nil
		}
		return "", fmt.Errorf("opening database: %w", err)
	}

	defer func() {
		if err := database.Close(); err != nil {
			log.Println(fmt.Errorf("closing database: %w", err))
		}
	}()

	md5Hash := md5.New()
	if _, err := io.Copy(md5Hash, database); err != nil {
		return "", fmt.Errorf("calculating database hash: %w", err)
	}

	result := byteToString(md5Hash.Sum(nil))
	if d.verbose {
		log.Printf("Calculated MD5 sum for %s: %s", databaseFilePath, result)
	}
	return result, nil
}

// getFilePath construct the file path for a database edition.
func (d *Download) getFilePath(editionID string) string {
	return filepath.Join(d.databaseDir, editionID) + Extension
}

// byteToString returns the base16 representation of a byte array.
func byteToString(b []byte) string {
	return hex.EncodeToString(b)
}

// ParseTime parses the date returned in a metadata response to time.Time.
func ParseTime(dateString string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", dateString, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing edition date: %w", err)
	}
	return t, nil
}
