package internal

import (
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
