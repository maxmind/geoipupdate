// Package database provides an abstraction over getting and writing a
// database file.
package database

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/maxmind/geoipupdate/v6/internal/vars"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/internal"
)

const (
	metadataEndpoint = "%s/geoip/updates/metadata?"
	downloadEndpoint = "%s/geoip/databases/%s/download?"
)

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
	verbose bool,
) *HTTPReader {
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
		verbose:    verbose,
	}
}

// Read attempts to fetch database updates for a specific editionID.
// It takes an editionID and it's previously downloaded hash if available
// as arguments and returns a ReadResult struct as a response.
// It's the responsibility of the Writer to close the io.ReadCloser
// included in the response after consumption.
func (r *HTTPReader) Read(ctx context.Context, editionID, hash string) (*ReadResult, error) {
	result, err := r.get(ctx, editionID, hash)
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
	edition, err := r.getMetadata(ctx, editionID)
	if err != nil {
		return nil, err
	}

	if edition.MD5 == hash {
		if r.verbose {
			log.Printf("No new updates available for %s", editionID)
		}
		return &ReadResult{EditionID: editionID, OldHash: hash, NewHash: hash}, nil
	}

	date := strings.ReplaceAll(edition.Date, "-", "")

	params := url.Values{}
	params.Add("date", date)
	params.Add("suffix", "tar.gz")

	escapedEdition := url.PathEscape(edition.EditionID)
	requestURL := fmt.Sprintf(downloadEndpoint, r.path, escapedEdition) + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating download request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(r.accountID), r.licenseKey)

	response, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing download request: %w", err)
	}
	// It is safe to close the response body reader as it wouldn't be
	// consumed in case this function returns an error.
	defer func() {
		if err != nil {
			response.Body.Close()
		}
	}()

	if response.StatusCode != http.StatusOK {
		//nolint:errcheck // we are already returning an error.
		buf, _ := io.ReadAll(io.LimitReader(response.Body, 256))
		httpErr := internal.HTTPError{
			Body:       string(buf),
			StatusCode: response.StatusCode,
		}
		return nil, fmt.Errorf("unexpected HTTP status code: %w", httpErr)
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, fmt.Errorf("encountered an error creating GZIP reader: %w", err)
	}
	defer func() {
		if err != nil {
			gzReader.Close()
		}
	}()

	tarReader := tar.NewReader(gzReader)

	// iterate through the tar archive to extract the mmdb file
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil, errors.New("tar archive does not contain an mmdb file")
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar archive: %w", err)
		}

		if strings.HasSuffix(header.Name, ".mmdb") {
			break
		}
	}

	modifiedAt, err := parseTime(response.Header.Get("Last-Modified"))
	if err != nil {
		return nil, fmt.Errorf("reading Last-Modified header: %w", err)
	}

	if r.verbose {
		log.Printf("Updates available for %s", editionID)
	}

	return &ReadResult{
		reader: editionReader{
			Reader:         tarReader,
			gzCloser:       gzReader,
			responseCloser: response.Body,
		},
		EditionID:  editionID,
		OldHash:    hash,
		NewHash:    edition.MD5,
		ModifiedAt: modifiedAt,
	}, nil
}

// metadata represents the metadata content for a certain database returned by the
// metadata endpoint.
type metadata struct {
	Date      string `json:"date"`
	EditionID string `json:"edition_id"`
	MD5       string `json:"md5"`
}

func (r *HTTPReader) getMetadata(ctx context.Context, editionID string) (*metadata, error) {
	params := url.Values{}
	params.Add("edition_id", editionID)

	metadataRequestURL := fmt.Sprintf(metadataEndpoint, r.path) + params.Encode()

	if r.verbose {
		log.Printf("Requesting metadata for %s: %s", editionID, metadataRequestURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataRequestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating metadata request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(r.accountID), r.licenseKey)

	response, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing metadata request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading metadata response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		httpErr := internal.HTTPError{
			Body:       string(responseBody),
			StatusCode: response.StatusCode,
		}
		return nil, fmt.Errorf("unexpected HTTP status code: %w", httpErr)
	}

	var metadataResponse struct {
		Databases []metadata `json:"databases"`
	}

	if err := json.Unmarshal(responseBody, &metadataResponse); err != nil {
		return nil, fmt.Errorf("parsing metadata body: %w", err)
	}

	if len(metadataResponse.Databases) != 1 {
		return nil, fmt.Errorf("response does not contain edition %s", editionID)
	}

	edition := metadataResponse.Databases[0]

	return &edition, nil
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

// editionReader embeds a tar.Reader and holds references to other readers to close.
type editionReader struct {
	*tar.Reader
	gzCloser       io.Closer
	responseCloser io.Closer
}

// Close closes the additional referenced readers.
func (e editionReader) Close() error {
	var err error
	if e.gzCloser != nil {
		gzErr := e.gzCloser.Close()
		if gzErr != nil {
			err = errors.Join(err, gzErr)
		}
	}

	if e.responseCloser != nil {
		responseErr := e.responseCloser.Close()
		if responseErr != nil {
			err = errors.Join(err, responseErr)
		}
	}
	return err
}
