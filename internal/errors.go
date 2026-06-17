// Package internal provides internal structures.
package internal

import (
	"errors"
	"fmt"
	"net/http"
)

// HTTPError is an error from performing an HTTP request.
type HTTPError struct {
	Body       string
	StatusCode int
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("received HTTP status code: %d: %s", h.StatusCode, h.Body)
}

// IsRetryableError returns true if the error should be retried.
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		return isRetryableHTTPStatusCode(httpErr.StatusCode)
	}

	// Keep unknown transport and filesystem errors retryable. This preserves
	// the existing retry behavior for non-HTTP failures while making HTTP
	// response semantics explicit.
	return true
}

func isRetryableHTTPStatusCode(statusCode int) bool {
	if statusCode >= 500 {
		return true
	}

	return statusCode == http.StatusRequestTimeout ||
		statusCode == http.StatusTooManyRequests
}
