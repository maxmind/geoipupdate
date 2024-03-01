// Package api provides a wrapper around the maxmind api responsible for
// checking/downloading mmdb files.
package api

import (
	"context"
	"fmt"
	"io"
	"time"
)

// DownloadAPI presents common methods required to download mmdb files.
type DownloadAPI interface {
	GetMetadata(ctx context.Context, editions []string) ([]Metadata, error)
	GetEdition(ctx context.Context, edition Metadata) (io.Reader, func(), error)
}

// Metadata represents the metadata content for a certain database returned by the
// metadata endpoint.
type Metadata struct {
	Date      string `json:"date"`
	EditionID string `json:"edition_id"`
	MD5       string `json:"md5"`
}

// ParseTime parses the date returned in a metadata response to time.Time.
func ParseTime(dateString string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", dateString, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing edition date: %w", err)
	}
	return t, nil
}
