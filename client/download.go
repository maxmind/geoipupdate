package client

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/maxmind/geoipupdate/v7/internal"
	"github.com/maxmind/geoipupdate/v7/internal/vars"
)

// DownloadResponse describes the result of a Download call.
type DownloadResponse struct {
	// LastModified is the date that the database was last modified. It will
	// only be set if UpdateAvailable is true.
	LastModified time.Time

	// MD5 is the string representation of the new database. It will only be set
	// if UpdateAvailable is true.
	MD5 string

	// Reader can be read to access the database itself. It will only contain a
	// database if UpdateAvailable is true.
	//
	// If the Download call does not return an error, Reader will always be
	// non-nil.
	//
	// If UpdateAvailable is true, the caller must read Reader to completion and
	// close it.
	Reader io.ReadCloser

	// UpdateAvailable is true if there is an update available for download. It
	// will be false if the MD5 used in the Download call matches what the server
	// currently has.
	UpdateAvailable bool
}

// Download attempts to download the edition.
//
// The editionID parameter is a valid database edition ID, such as
// "GeoIP2-City".
//
// The MD5 parameter is a string representation of the MD5 sum of the database
// MMDB file you have previously downloaded. If you don't yet have one
// downloaded, this can be "". This is used to know if an update is available
// and avoid consuming resources if there is not.
//
// If the current MD5 checksum matches what the server currently has, no
// download is performed.
func (c Client) Download(
	ctx context.Context,
	editionID,
	md5 string,
) (DownloadResponse, error) {
	metadata, err := c.getMetadata(ctx, editionID)
	if err != nil {
		return DownloadResponse{}, err
	}

	if metadata.MD5 == md5 {
		return DownloadResponse{
			Reader:          io.NopCloser(strings.NewReader("")),
			UpdateAvailable: false,
		}, nil
	}

	reader, modifiedTime, err := c.download(ctx, editionID, metadata.Date)
	if err != nil {
		return DownloadResponse{}, err
	}

	return DownloadResponse{
		LastModified:    modifiedTime,
		MD5:             metadata.MD5,
		Reader:          reader,
		UpdateAvailable: true,
	}, nil
}

const downloadEndpoint = "%s/geoip/databases/%s/download?"

func (c *Client) download(
	ctx context.Context,
	editionID,
	date string,
) (io.ReadCloser, time.Time, error) {
	date = strings.ReplaceAll(date, "-", "")

	params := url.Values{}
	params.Add("date", date)
	params.Add("suffix", "tar.gz")

	escapedEdition := url.PathEscape(editionID)
	requestURL := fmt.Sprintf(downloadEndpoint, c.endpoint, escapedEdition) + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("creating download request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(c.accountID), c.licenseKey)

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("performing download request: %w", err)
	}
	// It is safe to close the response body reader as it wouldn't be
	// consumed in case this function returns an error.
	defer func() {
		if err != nil {
			// TODO(horgh): Should we fully consume the body?
			response.Body.Close()
		}
	}()

	if response.StatusCode != http.StatusOK {
		// TODO(horgh): Should we fully consume the body?
		//nolint:errcheck // we are already returning an error.
		buf, _ := io.ReadAll(io.LimitReader(response.Body, 256))
		httpErr := internal.HTTPError{
			Body:       string(buf),
			StatusCode: response.StatusCode,
		}
		return nil, time.Time{}, fmt.Errorf("unexpected HTTP status code: %w", httpErr)
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("encountered an error creating GZIP reader: %w", err)
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
			return nil, time.Time{}, errors.New("tar archive does not contain an mmdb file")
		}
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("reading tar archive: %w", err)
		}

		if strings.HasSuffix(header.Name, ".mmdb") {
			break
		}
	}

	lastModified, err := parseTime(response.Header.Get("Last-Modified"))
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("reading Last-Modified header: %w", err)
	}

	return editionReader{
			Reader:         tarReader,
			gzCloser:       gzReader,
			responseCloser: response.Body,
		},
		lastModified,
		nil
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
