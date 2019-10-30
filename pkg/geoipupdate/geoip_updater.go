package geoipupdate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// NewClient creates an *http.Client for use in updating.
func NewClient(
	config *Config,
) *http.Client {
	var client *http.Client
	transport := http.DefaultTransport
	if config.Proxy != nil {
		proxy := http.ProxyURL(config.Proxy)
		transport.(*http.Transport).Proxy = proxy
	}
	client = &http.Client{Transport: transport}
	return client
}

// GetFilename looks up the filename for the given edition ID.
func GetFilename(
	config *Config,
	editionID string,
	client *http.Client,
) (string, error) {
	maxMindURL := fmt.Sprintf(
		"%s/app/update_getfilename?product_id=%s",
		config.URL,
		url.QueryEscape(editionID),
	)

	if config.Verbose {
		log.Printf("Performing get filename request to %s", maxMindURL)
	}
	res, err := client.Get(maxMindURL)
	if err != nil {
		return "", errors.Wrap(err, "error performing HTTP request")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Fatalf("error closing response body: %+v", errors.Wrap(err, "closing body"))
		}
	}()

	buf, err := ioutil.ReadAll(io.LimitReader(res.Body, 256))
	if err != nil {
		return "", errors.Wrap(err, "error reading response body")
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.Errorf("unexpected HTTP status code: %s: %s", res.Status, buf)
	}

	if len(buf) == 0 {
		return "", errors.New("response body is empty")
	}

	if bytes.Count(buf, []byte("\n")) > 0 ||
		bytes.Count(buf, []byte("\x00")) > 0 {
		return "", errors.New("invalid characters in filename")
	}

	return string(buf), nil
}
