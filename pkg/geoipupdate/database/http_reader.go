// Package database provides an abstraction over getting and writing a
// database file.
package database

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate"
	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/retry"
	"github.com/pkg/errors"
)

// HTTPDatabaseReader is a Reader that uses an HTTP client to retrieve
// databases.
type HTTPDatabaseReader struct {
	client            *http.Client
	retryFor          time.Duration
	url               string
	licenseKey        string
	accountID         int
	preserveFileTimes bool
	verbose           bool
}

// NewHTTPDatabaseReader creates a Reader that downloads database updates via
// HTTP.
func NewHTTPDatabaseReader(client *http.Client, config *geoipupdate.Config) Reader {
	return &HTTPDatabaseReader{
		client:            client,
		retryFor:          config.RetryFor,
		url:               config.URL,
		licenseKey:        config.LicenseKey,
		accountID:         config.AccountID,
		preserveFileTimes: config.PreserveFileTimes,
		verbose:           config.Verbose,
	}
}

// Get retrieves the given edition ID using an HTTP client, writes it to the
// Writer, and validates the hash before committing.
func (reader *HTTPDatabaseReader) Get(destination Writer, editionID string) error {
	defer func() {
		if err := destination.Close(); err != nil {
			log.Println(err)
		}
	}()

	maxMindURL := fmt.Sprintf(
		"%s/geoip/databases/%s/update?db_md5=%s",
		reader.url,
		url.PathEscape(editionID),
		url.QueryEscape(destination.GetHash()),
	)

	req, err := http.NewRequest(http.MethodGet, maxMindURL, nil) // nolint: noctx
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}
	req.SetBasicAuth(fmt.Sprintf("%d", reader.accountID), reader.licenseKey)

	if reader.verbose {
		log.Printf("Performing update request to %s", maxMindURL)
	}
	response, err := retry.MaybeRetryRequest(reader.client, reader.retryFor, req)
	if err != nil {
		return errors.Wrap(err, "error performing HTTP request")
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Fatalf("Error closing response body: %+v", errors.Wrap(err, "closing body"))
		}
	}()

	if response.StatusCode == http.StatusNotModified {
		if reader.verbose {
			log.Printf("No new updates available for %s", editionID)
		}
		return nil
	}

	if response.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(io.LimitReader(response.Body, 256))
		if err == nil {
			return errors.Errorf("unexpected HTTP status code: %s: %s", response.Status, buf)
		}
		return errors.Errorf("unexpected HTTP status code: %s", response.Status)
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return errors.Wrap(err, "encountered an error creating GZIP reader")
	}
	defer func() {
		if err := gzReader.Close(); err != nil {
			log.Printf("error closing gzip reader: %s", err)
		}
	}()

	if _, err = io.Copy(destination, gzReader); err != nil { //nolint:gosec
		return errors.Wrap(err, "error writing response")
	}

	newMD5 := response.Header.Get("X-Database-MD5")
	if newMD5 == "" {
		return errors.New("no X-Database-MD5 header found")
	}
	if err := destination.ValidHash(newMD5); err != nil {
		return err
	}

	if err := destination.Commit(); err != nil {
		return errors.Wrap(err, "encountered an issue committing database update")
	}

	if reader.preserveFileTimes {
		modificationTime, err := lastModified(response.Header.Get("Last-Modified"))
		if err != nil {
			return errors.Wrap(err, "unable to get last modified time")
		}
		err = destination.SetFileModificationTime(modificationTime)
		if err != nil {
			return errors.Wrap(err, "unable to set modification time")
		}
	}

	return nil
}

// LastModified retrieves the date that the MaxMind database was last modified.
func lastModified(lastModified string) (time.Time, error) {
	if lastModified == "" {
		return time.Time{}, errors.New("no Last-Modified header found")
	}

	t, err := time.ParseInLocation(time.RFC1123, lastModified, time.UTC)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "error parsing time")
	}

	return t, nil
}
