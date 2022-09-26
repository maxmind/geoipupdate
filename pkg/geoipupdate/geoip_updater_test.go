package geoipupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFileName(t *testing.T) {
	filename, _ := GetFilename(nil, "GeoIP2-City", nil)
	assert.Equal(
		t,
		"GeoIP2-City.mmdb",
		filename,
	)
}
