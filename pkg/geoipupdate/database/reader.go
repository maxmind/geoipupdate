package database

//Reader provides an interface for copying data from a MaxMind databases to a target Writer
type Reader interface {
	Get(destination Writer, editionID string) error
}
