// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/maxmind/geoipupdate/v4/pkg/geoipupdate/internal"
	"github.com/pkg/errors"
)

// NewClient creates an *http.Client for use in updating.
func NewClient(
	config *Config,
) *http.Client {
	transport := http.DefaultTransport
	if config.Proxy != nil {
		proxy := http.ProxyURL(config.Proxy)
		transport.(*http.Transport).Proxy = proxy
	}
	return &http.Client{Transport: transport}
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

	var buf []byte
	err := internal.RetryWithBackoff(
		func() error {
			if config.Verbose {
				log.Printf("Performing get filename request to %s", maxMindURL)
			}

			//nolint: noctx // as it would require a breaking API change
			req, err := http.NewRequest(http.MethodGet, maxMindURL, nil)
			if err != nil {
				return errors.Wrap(err, "error creating HTTP request")
			}
			req.Header.Add("User-Agent", "geoipupdate/"+Version)

			res, err := client.Do(req)
			if err != nil {
				return errors.Wrap(err, "error performing HTTP request")
			}

			buf, err = ioutil.ReadAll(io.LimitReader(res.Body, 256))
			if err != nil {
				_ = res.Body.Close()
				return errors.Wrap(err, "error reading response body")
			}

			if err := res.Body.Close(); err != nil {
				return errors.Wrap(err, "closing body")
			}

			if res.StatusCode != http.StatusOK {
				err := internal.HTTPError{
					Body:       string(buf),
					StatusCode: res.StatusCode,
				}
				return errors.Wrap(err, "unexpected HTTP status code")
			}

			if len(buf) == 0 {
				return errors.New("response body is empty")
			}

			if bytes.Count(buf, []byte("\n")) > 0 ||
				bytes.Count(buf, []byte("\x00")) > 0 {
				return errors.New("invalid characters in filename")
			}

			return nil
		},
		config.RetryFor,
	)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}
