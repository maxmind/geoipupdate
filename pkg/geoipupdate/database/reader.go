package database

import (
	"context"
	"encoding/json"
	"fmt"
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
	Reader     io.ReadCloser `json:"-"`
	EditionID  string        `json:"edition_id"`
	OldHash    string        `json:"old_hash"`
	NewHash    string        `json:"new_hash"`
	ModifiedAt time.Time     `json:"modified_at"`
	CheckedAt  time.Time     `json:"checked_at"`
}

// MarshalJSON is a custom json marshaler that strips out zero time fields.
func (r ReadResult) MarshalJSON() ([]byte, error) {
	type partialResult ReadResult
	s := &struct {
		partialResult
		ModifiedAt int64 `json:"modified_at,omitempty"`
		CheckedAt  int64 `json:"checked_at,omitempty"`
	}{
		partialResult: partialResult(r),
		ModifiedAt:    0,
		CheckedAt:     0,
	}

	if !r.ModifiedAt.IsZero() {
		s.ModifiedAt = r.ModifiedAt.Unix()
	}

	if !r.CheckedAt.IsZero() {
		s.CheckedAt = r.CheckedAt.Unix()
	}

	res, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("marshaling ReadResult: %w", err)
	}
	return res, nil
}

// UnmarshalJSON is a custom json unmarshaler that converts timestamps to go
// time fields.
func (r *ReadResult) UnmarshalJSON(data []byte) error {
	type partialResult ReadResult
	s := &struct {
		partialResult
		ModifiedAt int64 `json:"modified_at,omitempty"`
		CheckedAt  int64 `json:"checked_at,omitempty"`
	}{}

	err := json.Unmarshal(data, &s)
	if err != nil {
		return fmt.Errorf("unmarshaling json into ReadResult: %w", err)
	}

	result := ReadResult(s.partialResult)
	result.ModifiedAt = time.Unix(s.ModifiedAt, 0).In(time.UTC)
	result.CheckedAt = time.Unix(s.CheckedAt, 0).In(time.UTC)
	*r = result

	return nil
}
