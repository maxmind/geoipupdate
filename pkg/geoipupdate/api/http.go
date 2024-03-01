package api

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/internal"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/vars"
)

const (
	metadataEndpoint = "%s/geoip/updates/metadata?%s"
	downloadEndpoint = "%s/geoip/databases/%s/download?date=%s&suffix=tar.gz"
)

// httpDownloader is a http implementation of the DownloadAPI interface.
type httpDownloader struct {
	// accountID is the requester's account ID.
	accountID int
	// client is an http client responsible of fetching database updates.
	client *http.Client
	// host points to maxmind servers.
	host string
	// licenseKey is the requester's license key.
	licenseKey string
}

// NewHTTPDownloader initializes a new httpDownloader struct.
//
//nolint:revive // unexported type fileLock is not meant to be used as a standalone type.
func NewHTTPDownloader(
	accountID int,
	licenseKey string,
	client *http.Client,
	host string,
) *httpDownloader {
	return &httpDownloader{
		accountID:  accountID,
		client:     client,
		host:       host,
		licenseKey: licenseKey,
	}
}

// GetMetadata makes an http request to retrieve metadata about the provided database editions.
func (h *httpDownloader) GetMetadata(ctx context.Context, editions []string) ([]Metadata, error) {
	var editionsQuery []string
	for _, e := range editions {
		editionsQuery = append(editionsQuery, "edition_id="+url.QueryEscape(e))
	}

	requestURL := fmt.Sprintf(metadataEndpoint, h.host, strings.Join(editionsQuery, "&"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(h.accountID), h.licenseKey)

	response, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing HTTP request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		errResponse := internal.ResponseError{
			StatusCode: response.StatusCode,
		}

		if err := json.Unmarshal(responseBody, &errResponse); err != nil {
			errResponse.Message = err.Error()
		}

		return nil, fmt.Errorf("requesting metadata: %w", errResponse)
	}

	var metadataResponse struct {
		Databases []Metadata `json:"databases"`
	}

	if err := json.Unmarshal(responseBody, &metadataResponse); err != nil {
		return nil, fmt.Errorf("parsing body: %w", err)
	}

	return metadataResponse.Databases, nil
}

// GetEdition makes an http request to download the requested database edition.
// It returns an io.Reader that points to the content of the database file.
func (h *httpDownloader) GetEdition(
	ctx context.Context,
	edition Metadata,
) (reader io.Reader, cleanupCallback func(), err error) {
	date := strings.ReplaceAll(edition.Date, "-", "")
	requestURL := fmt.Sprintf(downloadEndpoint, h.host, edition.EditionID, date)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(h.accountID), h.licenseKey)

	response, err := h.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("performing HTTP request: %w", err)
	}
	// It is safe to close the response body reader as it wouldn't be
	// consumed in case this function returns an error.
	defer func() {
		if err != nil {
			response.Body.Close()
		}
	}()

	if response.StatusCode != http.StatusOK {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, nil, fmt.Errorf("reading error response body: %w", err)
		}

		errResponse := internal.ResponseError{
			StatusCode: response.StatusCode,
		}

		if err := json.Unmarshal(responseBody, &errResponse); err != nil {
			errResponse.Message = err.Error()
		}

		return nil, nil, fmt.Errorf("requesting edition: %w", errResponse)
	}

	gzReader, err := gzip.NewReader(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("encountered an error creating GZIP reader: %w", err)
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
			return nil, nil, errors.New("tar archive does not contain an mmdb file")
		}
		if err != nil {
			return nil, nil, fmt.Errorf("reading tar archive: %w", err)
		}

		if strings.HasSuffix(header.Name, ".mmdb") {
			break
		}
	}

	cleanupCallback = func() {
		gzReader.Close()
		response.Body.Close()
	}

	return tarReader, cleanupCallback, nil
}
