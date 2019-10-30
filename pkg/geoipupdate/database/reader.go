package database

// Reader provides an interface for retrieving a database update and copying it
// into place.
type Reader interface {
	Get(destination Writer, editionID string) error
}
