package geoipupdate

import (
	"testing"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/api"
	"github.com/stretchr/testify/require"
)

func TestOutputFormat(t *testing.T) {
	oldHashes := map[string]string{
		"edition-1": "1",
		"edition-2": "10",
	}

	metadata := []api.Metadata{
		{
			EditionID: "edition-1",
			MD5:       "1",
			Date:      "2024-01-01",
		},
		{
			EditionID: "edition-2",
			MD5:       "11",
			Date:      "2024-02-01",
		},
	}

	//nolint:lll
	expectedOutput := `[{"edition_id":"edition\-1","old_hash":"1","new_hash":"1","modified_at":1704067200,"checked_at":\d+},{"edition_id":"edition\-2","old_hash":"10","new_hash":"11","modified_at":1706745600,"checked_at":\d+}]`

	output, err := makeOutput(metadata, oldHashes)
	require.NoError(t, err)
	require.Regexp(t, expectedOutput, string(output))
}
