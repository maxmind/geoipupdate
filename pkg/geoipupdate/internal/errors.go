// Package internal provides internal structures.
package internal

import (
	"errors"
	"fmt"
)

// ResponseError represents an error response returned by the geoip servers.
type ResponseError struct {
	StatusCode int
	Code       string `json:"code"`
	Message    string `json:"error"`
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("received: HTTP status code '%d' - Error code '%s' - Message '%s'", e.StatusCode, e.Code, e.Message)
}

// IsPermanentError returns true if the error is non-retriable.
func IsPermanentError(err error) bool {
	var r ResponseError
	if errors.As(err, &r) && r.StatusCode >= 400 && r.StatusCode < 500 {
		return true
	}

	return false
}
