package download

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestGetHash tests functionality of the LocalFileWriter.GetHash method.
func TestGetHash(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "db")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	d := Download{
		databaseDir: tempDir,
	}

	// returns a zero hash for a non existing edition.
	hash, err := d.getHash("NewEdition")
	require.NoError(t, err)
	require.Equal(t, zeroMD5, hash)

	// returns the correct md5 for an existing edition.
	edition := "edition-1"
	dbFile := filepath.Join(tempDir, edition+Extension)

	err = os.WriteFile(dbFile, []byte("edition-1 content"), os.ModePerm)
	require.NoError(t, err)

	hash, err = d.getHash(edition)
	require.NoError(t, err)
	require.Equal(t, "618dd27a10de24809ec160d6807f363f", hash)
}
