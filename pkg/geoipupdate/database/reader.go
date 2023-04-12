package database

import (
	"context"
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
	editionID  string
	oldHash    string
	newHash    string
	modifiedAt time.Time
}
