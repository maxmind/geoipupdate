package database

import (
	"compress/gzip"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

//HTTPDatabaseReader is a database.Reader that uses an HTTP Client to retrieve the database data
type HTTPDatabaseReader struct {
	*http.Client
	URL        string
	LicenseKey string
	AccountID  int
	Verbose    bool
	response   *http.Response
}

//Get retrieves the data for a given editionID using an HTTP client to MaxMind, writes it to database.Writer,
// and validates the associated hash before committing
func (reader *HTTPDatabaseReader) Get(destination Writer, editionID string) error {
	defer func() {
		if err := destination.Close(); err != nil {
			log.Println(err)
		}
	}()
	lastHash, err := destination.GetHash()
	if err != nil {
		return errors.Wrap(err, "Unable to get previous hash")
	}
	url := fmt.Sprintf(
		"%s/geoip/databases/%s/update?db_md5=%s",
		reader.URL,
		url.PathEscape(editionID),
		url.QueryEscape(lastHash),
	)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return errors.Wrap(err, "error creating request")
	}
	if reader.AccountID != 0 {
		req.SetBasicAuth(fmt.Sprintf("%d", reader.AccountID), reader.LicenseKey)
	}

	if reader.Verbose {
		log.Printf("Performing update request to %s", url)
	}
	reader.response, err = reader.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error performing HTTP request")
	}
	defer func() {
		if err := reader.response.Body.Close(); err != nil {
			log.Fatalf("Error closing response body: %+v", errors.Wrap(err, "closing body"))
		}
	}()

	if reader.response.StatusCode == http.StatusNotModified {
		if reader.Verbose {
			log.Printf("No new updates available for %s", editionID)
		}
		return nil
	}

	if reader.response.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(io.LimitReader(reader.response.Body, 256))
		if err == nil {
			return errors.Errorf("unexpected HTTP status code: %s: %s", reader.response.Status, buf)
		}
		return errors.Errorf("unexpected HTTP status code: %s", reader.response.Status)
	}

	newMD5 := reader.response.Header.Get("X-Database-MD5")
	if newMD5 == "" {
		return errors.New("no X-Database-MD5 header found")
	}

	gzReader, err := gzip.NewReader(reader.response.Body)
	if err != nil {
		return errors.Wrap(err, "Encounter an error created GZIP reader")
	}

	if _, err = io.Copy(destination, gzReader); err != nil {
		return errors.Wrap(err, "Encountered an error writing out MaxMind's response")
	}

	if err := destination.ValidHash(newMD5); err != nil {
		return err
	}

	if err := destination.Commit(); err != nil {
		return errors.Wrap(err, "Encountered an issue committing database update")
	}

	return nil
}

//LastModified retrieves the date that the MaxMind database was last modified
func (reader *HTTPDatabaseReader) LastModified() (time.Time, error) {
	if reader.response == nil {
		return time.Time{}, errors.New("Request hasn't been made yet for data")
	}
	lastModifiedStr := reader.response.Header.Get("Last-Modified")
	if lastModifiedStr == "" {
		return time.Time{}, errors.New("no Last-Modified header found")
	}

	t, err := time.ParseInLocation(time.RFC1123, lastModifiedStr, time.UTC)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "error parsing time")
	}

	return t, nil
}
