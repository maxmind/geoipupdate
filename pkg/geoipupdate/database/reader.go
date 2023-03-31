package database

// Reader provides an interface for retrieving a database update and copying it
// into place.
type Reader interface {
	Queue(destination Writer, editionID string)
	Get(destination Writer, editionID string) error
	Wait() error
	Stop() error
}
