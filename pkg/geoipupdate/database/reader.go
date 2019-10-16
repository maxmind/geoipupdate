package database

import "time"

type Reader interface {
	Get(destination Writer, editionID string) error
	LastModified() (time.Time, error)
}
