// Package database provides an abstraction over getting and writing a
// database file.
package database

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/internal"
	"github.com/maxmind/geoipupdate/v5/pkg/geoipupdate/vars"
)

const urlFormat = "%s/geoip/databases/%s/update?db_md5=%s"

// HTTPReader is a Reader that uses an HTTP client to retrieve
// databases.
type HTTPReader struct {
	// client is an http client responsible of fetching database updates.
	client *http.Client
	// path is the request path.
	path string
	// accountID is used for request auth.
	accountID int
	// licenseKey is used for request auth.
	licenseKey string
	// retryFor sets the timeout for when a request can no longuer be retried.
	retryFor time.Duration
	// verbose turns on/off debug logs.
	verbose bool
}

// NewHTTPReader creates a Reader that downloads database updates via
// HTTP.
func NewHTTPReader(
	proxy *url.URL,
	path string,
	accountID int,
	licenseKey string,
	retryFor time.Duration,
	verbose bool,
) Reader {
	transport := http.DefaultTransport
	if proxy != nil {
		proxyFunc := http.ProxyURL(proxy)
		transport.(*http.Transport).Proxy = proxyFunc
	}

	return &HTTPReader{
		client:     &http.Client{Transport: transport},
		path:       path,
		accountID:  accountID,
		licenseKey: licenseKey,
		retryFor:   retryFor,
		verbose:    verbose,
	}
}

// Read attempts to fetch database updates for a specific editionID.
// It takes an editionID and it's previously downloaded hash if available
// as arguments and returns a ReadResult struct as a response.
// It's the responsibility of the Writer to close the io.ReadCloser
// included in the response after consumption.
func (r *HTTPReader) Read(ctx context.Context, editionID, hash string) (*ReadResult, error) {
	var result *ReadResult
	var err error

	// RetryFor value of 0 means that no retries should be performed.
	// Max zero retries has to be set to achieve that
	// because the backoff never stops if MaxElapsedTime is zero.
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = r.retryFor
	b := backoff.BackOff(exp)
	if exp.MaxElapsedTime == 0 {
		b = backoff.WithMaxRetries(exp, 0)
	}
	err = backoff.RetryNotify(
		func() error {
			result, err = r.get(ctx, editionID, hash)
			if err == nil {
				return nil
			}

			var httpErr internal.HTTPError
			if errors.As(err, &httpErr) && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
				return backoff.Permanent(err)
			}

			return err
		},
		b,
		func(err error, d time.Duration) {
			if r.verbose {
				log.Printf("Couldn't download %s, retrying in %v: %v", editionID, d, err)
			}
		},
	)
	if err != nil {
		return nil, fmt.Errorf("getting update for %s: %w", editionID, err)
	}

	return result, nil
}

// get makes an http request to fetch updates for a specific editionID if any.
func (r *HTTPReader) get(
	ctx context.Context,
	editionID string,
	hash string,
) (result *ReadResult, err error) {
	requestURL := fmt.Sprintf(
		urlFormat,
		r.path,
		url.PathEscape(editionID),
		url.QueryEscape(hash),
	)

	if r.verbose {
		log.Printf("Requesting updates for %s: %s", editionID, requestURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(fmt.Sprintf("%d", r.accountID), r.licenseKey)

	response, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing HTTP request: %w", err)
	}
	// It is safe to close the response body reader as it wouldn't be
	// consumed in case this function returns an error.
	defer func() {
		if err != nil {
			response.Body.Close()
		}
	}()

	switch response.StatusCode {
	case http.StatusNotModified:
		if r.verbose {
			log.Printf("No new updates available for %s", editionID)
		}
		return &ReadResult{EditionID: editionID, OldHash: hash, NewHash: hash}, nil
	case http.StatusOK:
	default:
		//nolint:errcheck // we are already returning an error.
		buf, _ := io.ReadAll(io.LimitReader(response.Body, 256))
		httpErr := internal.HTTPError{
			Body:       string(buf),
			StatusCode: response.StatusCode,
		}
		return nil, fmt.Errorf("unexpected HTTP status code: %w", httpErr)
	}

	newHash := response.Header.Get("X-Database-MD5")
	if newHash == "" {
		return nil, errors.New("no X-Database-MD5 header found")
	}

	modifiedAt, err := parseTime(response.Header.Get("Last-Modified"))
	if err != nil {
		return nil, fmt.Errorf("reading Last-Modified header: %w", err)
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, fmt.Errorf("encountered an error creating GZIP reader: %w", err)
	}

	if r.verbose {
		log.Printf("Updates available for %s", editionID)
	}

	return &ReadResult{
		reader:     gzReader,
		EditionID:  editionID,
		OldHash:    hash,
		NewHash:    newHash,
		ModifiedAt: modifiedAt,
	}, nil
}

// parseTime parses a string representation of a time into time.Time according to the
// RFC1123 format.
func parseTime(s string) (time.Time, error) {
	t, err := time.ParseInLocation(time.RFC1123, s, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing time: %w", err)
	}

	return t, nil
}
