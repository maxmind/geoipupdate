package database

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

// Reader provides an interface for retrieving a database update and copying it
// into place.
type Reader interface {
	Read(context.Context, string, string) (*ReadResult, error)
}

// ReadResult is the struct returned by a Reader's Get method.
type ReadResult struct {
	reader     io.ReadCloser
	EditionID  string    `json:"edition_id"`
	OldHash    string    `json:"old_hash"`
	NewHash    string    `json:"new_hash"`
	ModifiedAt time.Time `json:"modified_at"`
	CheckedAt  time.Time `json:"checked_at"`
}

// MarshalJSON is a custom json marshaler that strips out zero time fields.
func (r ReadResult) MarshalJSON() ([]byte, error) {
	type partialResult ReadResult
	s := &struct {
		partialResult
		ModifiedAt interface{} `json:"modified_at,omitempty"`
	}{
		partialResult: partialResult(r),
		ModifiedAt:    nil,
	}

	if !r.ModifiedAt.IsZero() {
		s.ModifiedAt = r.ModifiedAt
	}

	return json.Marshal(s)
}
