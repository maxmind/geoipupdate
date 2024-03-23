package client

import (
	"net/http"
)

// HTTPReader is a Reader that uses an HTTP client to retrieve
// databases.
type HTTPReader struct {
	// client is an http client responsible of fetching database updates.
	client *http.Client
	// path is the request path.
	path string
	// accountID is used for request auth.
	accountID int
	// licenseKey is used for request auth.
	licenseKey string
	// verbose turns on/off debug logs.
	verbose bool
}

// NewHTTPReader creates a Reader that downloads database updates via
// HTTP.
func NewHTTPReader(
	path string,
	accountID int,
	licenseKey string,
	verbose bool,
	httpClient *http.Client,
) *HTTPReader {
	return &HTTPReader{
		client:     httpClient,
		path:       path,
		accountID:  accountID,
		licenseKey: licenseKey,
		verbose:    verbose,
	}
}
