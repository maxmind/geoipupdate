package download

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGetOutdatedEditions checks the metadata fetching functionality.
func TestGetOutdatedEditions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	edition := "edition-1"
	dbFile := filepath.Join(tempDir, edition+Extension)
	// equivalent MD5: 618dd27a10de24809ec160d6807f363f
	err = os.WriteFile(dbFile, []byte("edition-1 content"), os.ModePerm)
	require.NoError(t, err)

	edition = "edition-2"
	dbFile = filepath.Join(tempDir, edition+Extension)
	err = os.WriteFile(dbFile, []byte("edition-2 content"), os.ModePerm)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonData := `
{
    "databases": [
        {
            "edition_id": "edition-1",
            "md5": "618dd27a10de24809ec160d6807f363f",
            "date": "2024-02-23"
        },
        {
            "edition_id": "edition-2",
            "md5": "abc123",
            "date": "2024-02-23"
        },
        {
            "edition_id": "edition-3",
            "md5": "def456",
            "date": "2024-02-02"
        }
    ]
}
`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonData))
	}))
	defer server.Close()

	ctx := context.Background()
	d, err := New(
		0,  // accountID is not relevant for this test.
		"", // licenseKey is not relevant for this test.
		server.URL,
		nil, // proxy is not relevant for this test.
		tempDir,
		false, // preserveFileTimes is not relevant for this test.
		[]string{"edition-1", "edition-2", "edition-3"},
		false, // verbose is not relevant for this test.
	)
	require.NoError(t, err)

	// edition-1 md5 hasn't changed
	expectedOutdatedEditions := []Metadata{
		{
			EditionID: "edition-2",
			MD5:       "abc123",
			Date:      "2024-02-23",
		},
		{
			EditionID: "edition-3",
			MD5:       "def456",
			Date:      "2024-02-02",
		},
	}

	outdatedEditions, err := d.GetOutdatedEditions(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, expectedOutdatedEditions, outdatedEditions)

	expectedDatabases := append(
		expectedOutdatedEditions,
		Metadata{
			EditionID: "edition-1",
			MD5:       "618dd27a10de24809ec160d6807f363f",
			Date:      "2024-02-23",
		},
	)
	require.ElementsMatch(t, expectedDatabases, d.metadata)
}
