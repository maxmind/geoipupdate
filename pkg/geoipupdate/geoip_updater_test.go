package geoipupdate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFileName(t *testing.T) {
	filename, err := GetFilename(nil, "GeoIP2-City", nil)
	require.NoError(t, err)
	assert.Equal(
		t,
		"GeoIP2-City.mmdb",
		filename,
	)
}
