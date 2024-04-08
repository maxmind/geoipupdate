package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/maxmind/geoipupdate/v7/internal"
	"github.com/maxmind/geoipupdate/v7/internal/vars"
)

const metadataEndpoint = "%s/geoip/updates/metadata?"

// metadata represents the metadata content for a certain database returned by the
// metadata endpoint.
type metadata struct {
	Date      string `json:"date"`
	EditionID string `json:"edition_id"`
	MD5       string `json:"md5"`
}

func (c *Client) getMetadata(
	ctx context.Context,
	editionID string,
) (*metadata, error) {
	params := url.Values{}
	params.Add("edition_id", editionID)

	metadataRequestURL := fmt.Sprintf(metadataEndpoint, c.endpoint) + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, metadataRequestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating metadata request: %w", err)
	}
	req.Header.Add("User-Agent", "geoipupdate/"+vars.Version)
	req.SetBasicAuth(strconv.Itoa(c.accountID), c.licenseKey)

	response, err := c.httpClient.Do(req)
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
