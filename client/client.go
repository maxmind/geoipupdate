// Package client is a client for downloading GeoIP2 and GeoLite2 MMDB
// databases.
package client

import (
	"fmt"
	"net/http"
)

// Client downloads GeoIP2 and GeoLite2 MMDB databases.
//
// After creation, it is valid for concurrent use.
//
//nolint:recvcheck // changing this would be a breaking change.
type Client struct {
	accountID  int
	endpoint   string
	httpClient *http.Client
	licenseKey string
}

// Option is an option for configuring Client.
type Option func(*Client)

// WithEndpoint sets the base endpoint to use. By default we use
// https://updates.maxmind.com.
func WithEndpoint(endpoint string) Option {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithHTTPClient sets the HTTP client to use. By default we use
// http.DefaultClient.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// New creates a Client.
func New(
	accountID int,
	licenseKey string,
	options ...Option,
) (Client, error) {
	if accountID <= 0 {
		return Client{}, fmt.Errorf("invalid account ID: %d", accountID)
	}

	if licenseKey == "" {
		return Client{}, fmt.Errorf("invalid license key: %s", licenseKey)
	}

	c := Client{
		accountID:  accountID,
		endpoint:   "https://updates.maxmind.com",
		httpClient: http.DefaultClient,
		licenseKey: licenseKey,
	}

	for _, opt := range options {
		opt(&c)
	}

	return c, nil
}
