package geoipupdate

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/maxmind/geoipupdate/v6/pkg/geoipupdate/api"
)

// output contains information collected about a certain database edition
// collected during a download attempt.
type output struct {
	EditionID  string `json:"edition_id"`
	OldHash    string `json:"old_hash"`
	NewHash    string `json:"new_hash"`
	ModifiedAt int64  `json:"modified_at"`
	CheckedAt  int64  `json:"checked_at"`
}

// makeOutput returns a json formatted summary about the current download attempt.
func makeOutput(allEditions []api.Metadata, oldHashes map[string]string) ([]byte, error) {
	var out []output
	now := time.Now()
	for _, m := range allEditions {
		modifiedAt, err := api.ParseTime(m.Date)
		if err != nil {
			return nil, err
		}
		o := output{
			EditionID:  m.EditionID,
			OldHash:    oldHashes[m.EditionID],
			NewHash:    m.MD5,
			ModifiedAt: modifiedAt.Unix(),
			CheckedAt:  now.Unix(),
		}
		out = append(out, o)
	}
	res, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshaling output: %w", err)
	}
	return res, nil
}
