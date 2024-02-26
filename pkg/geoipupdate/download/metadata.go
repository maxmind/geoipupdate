package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/internal"
	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/vars"
)

const (
	metadataEndpoint = "%s/geoip/updates/metadata?%s"
)

// metadataResponse represents a successful response returned by the metadata endpoint.
type metadataResponse struct {
	Databases []Metadata `json:"databases"`
}

// Metadata represents the metadata content for a certain database returned by the
// metadata endpoint.
type Metadata struct {
	Date      string `json:"date"`
	EditionID string `json:"edition_id"`
	MD5       string `json:"md5"`
}

// GetOutdatedEditions returns the list of outdated database editions.
func (d *Download) GetOutdatedEditions(ctx context.Context) ([]Metadata, error) {
	var editionsQuery []string
	for _, e := range d.editionIDs {
		editionsQuery = append(editionsQuery, "edition_id="+url.QueryEscape(e))
	}

	requestURL := fmt.Sprintf(metadataEndpoint, d.url, strings.Join(editionsQuery, "&"))
	if d.verbose {
		log.Printf("Requesting edition metadata: %s", requestURL)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(d.accountID), d.licenseKey)

	response, err := d.client.Do(req)
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

	var metadata metadataResponse
	if err := json.Unmarshal(responseBody, &metadata); err != nil {
		return nil, fmt.Errorf("parsing body: %w", err)
	}

	var outdatedEditions []Metadata
	for _, m := range metadata.Databases {
		oldMD5 := d.oldEditionsHash[m.EditionID]
		if oldMD5 != m.MD5 {
			outdatedEditions = append(outdatedEditions, m)
			continue
		}

		if d.verbose {
			log.Printf("Database %s up to date", m.EditionID)
		}
	}

	d.metadata = metadata.Databases

	return outdatedEditions, nil
}
