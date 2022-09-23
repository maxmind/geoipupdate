// Package geoipupdate provides a library for using MaxMind's GeoIP Update
// service.
package geoipupdate

import (
	"net/http"
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

// GetFilename returns the filename for the given edition ID.
func GetFilename(
	_ *Config,
	editionID string,
	_ *http.Client,
) (string, error) { //nolint:unparam // the error return value was kept for API compatibility
	return editionID + ".mmdb", nil
}
