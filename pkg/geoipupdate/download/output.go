package download

import (
	"encoding/json"
	"fmt"
)

// ReadResult is the struct returned by a Reader's Get method.
type output struct {
	EditionID  string `json:"edition_id"`
	OldHash    string `json:"old_hash"`
	NewHash    string `json:"new_hash"`
	ModifiedAt int64  `json:"modified_at"`
	CheckedAt  int64  `json:"checked_at"`
}

// MakeOutput returns a json formatted summary about the current download attempt.
func (d *Download) MakeOutput() ([]byte, error) {
	var out []output
	now := d.now()
	for _, m := range d.metadata {
		modifiedAt, err := ParseTime(m.Date)
		if err != nil {
			return nil, err
		}
		o := output{
			EditionID:  m.EditionID,
			OldHash:    d.oldEditionsHash[m.EditionID],
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
