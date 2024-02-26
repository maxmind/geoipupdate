package download

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOutputFormat(t *testing.T) {
	now, err := ParseTime("2024-02-23")
	require.NoError(t, err)

	d := Download{
		oldEditionsHash: map[string]string{
			"edition-1": "1",
			"edition-2": "10",
		},
		metadata: []Metadata{
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
		},
		now: func() time.Time { return now },
	}

	expectedOutput := `[{"edition_id":"edition-1","old_hash":"1","new_hash":"1","modified_at":1704067200,"checked_at":1708646400},{"edition_id":"edition-2","old_hash":"10","new_hash":"11","modified_at":1706745600,"checked_at":1708646400}]`

	output, err := d.MakeOutput()
	require.NoError(t, err)
	require.Equal(t, expectedOutput, string(output))
}
