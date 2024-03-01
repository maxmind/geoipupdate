// Package lock is responsible for synchronizing concurrent access to the client.
package lock

// Lock presents common methods required to prevent concurrent access to
// the download client.
type Lock interface {
	Acquire() error
	Release() error
}
