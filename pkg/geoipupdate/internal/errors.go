// Package internal provides internal structures.
package internal

import (
	"errors"
	"fmt"
)

// HTTPError is an error from performing an HTTP request.
type HTTPError struct {
	Body       string
	StatusCode int
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("received HTTP status code: %d: %s", h.StatusCode, h.Body)
}

// IsPermanentError returns true if the error is non-retriable.
func IsPermanentError(err error) bool {
	var httpErr HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 {
		return true
	}

	return false
}
