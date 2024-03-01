// Package writer is responsible for writing databases editions to various
// destinations.
package writer

import (
	"io"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/api"
)

const (
	// zeroMD5 is the default value provided as an MD5 hash for a non-existent
	// database.
	zeroMD5 = "00000000000000000000000000000000"
)

// Writer presents common methods required to write mmdb databases.
type Writer interface {
	GetHash(editionID string) (string, error)
	Write(metadata api.Metadata, reader io.Reader) error
}
