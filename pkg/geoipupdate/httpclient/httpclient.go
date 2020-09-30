package httpclient

import (
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pkg/errors"
)

// HTTPClient is an HTTP client that retries with exponential backoff on errors
type HTTPClient struct {
	client   *http.Client
	retryFor time.Duration
}

// NewHTTPClient makes a new client that'll retry for the given duration for each request
func NewHTTPClient(c *http.Client, retryFor time.Duration) *HTTPClient {
	return &HTTPClient{c, retryFor}
}

// Do performs the given request and yields a successful HTTP response or the error that occurred during the last retry
func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = c.retryFor
	var resp *http.Response
	err := backoff.Retry(
		func() error {
			var err error
			resp, err = c.client.Do(req)
			return errors.Wrap(err, "error performing http request")
		},
		exp,
	)
	return resp, err
}
