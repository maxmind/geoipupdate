package client

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/maxmind/geoipupdate/v6/internal"
	"github.com/maxmind/geoipupdate/v6/internal/vars"
)

// Read attempts to fetch database updates for a specific editionID.
// It takes an editionID and its previously downloaded hash if available
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

const downloadEndpoint = "%s/geoip/databases/%s/download?"

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
