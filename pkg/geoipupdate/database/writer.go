package database

import (
	"io"
	"time"
)

type Writer interface {
	io.WriteCloser
	ValidHash(expectedHash string) error
	GetHash() (string, error)
	SetFileModificationTime(lastModified time.Time) error
	Commit() error
}
