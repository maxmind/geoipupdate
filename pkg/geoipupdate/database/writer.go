package database

import (
	"io"
	"time"
)

// Writer provides an interface for writing the database to a target location.
type Writer interface {
	io.WriteCloser
	ValidHash(expectedHash string) error
	GetHash() (string, error)
	SetFileModificationTime(lastModified time.Time) error
	Commit() error
}
