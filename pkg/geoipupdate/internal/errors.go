// Package internal provides internal structures.
package internal

import (
	"errors"
	"fmt"

	"golang.org/x/net/http2"
)

// HTTPError is an error from performing an HTTP request.
type HTTPError struct {
	Body       string
	StatusCode int
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("received HTTP status code: %d: %s", h.StatusCode, h.Body)
}

// IsTemporaryError returns true if the error is temporary.
func IsTemporaryError(err error) bool {
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		isPermanent := httpErr.StatusCode >= 400 && httpErr.StatusCode < 500
		return !isPermanent
	}

	var streamErr http2.StreamError
	if errors.As(err, &streamErr) && streamErr.Code == http2.ErrCodeInternal {
		return true
	}

	return false
}
