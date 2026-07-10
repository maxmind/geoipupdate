package internal

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/http2"
)

func TestIsRetryableError(t *testing.T) {
	tt := map[string]struct {
		err  error
		want bool
	}{
		"HTTP2 INTERNAL_ERROR": {
			err: http2.StreamError{
				Code: http2.ErrCodeInternal,
			},
			want: true,
		},
		"bad gateway": {
			err: HTTPError{
				StatusCode: http.StatusBadGateway,
			},
			want: true,
		},
		"too many requests": {
			err: HTTPError{
				StatusCode: http.StatusTooManyRequests,
			},
			want: true,
		},
		"request timeout": {
			err: HTTPError{
				StatusCode: http.StatusRequestTimeout,
			},
			want: true,
		},
		"bad request": {
			err: HTTPError{
				StatusCode: http.StatusBadRequest,
			},
			want: false,
		},
		"not found": {
			err: HTTPError{
				StatusCode: http.StatusNotFound,
			},
			want: false,
		},
		"url error wrapping proxy CONNECT HTTPError forbidden": {
			err: &url.Error{
				Op:  "Get",
				URL: "https://updates.maxmind.com/geoip/updates/metadata?edition_id=GeoIP2-City",
				Err: HTTPError{StatusCode: http.StatusForbidden},
			},
			want: false,
		},
		"url error wrapping proxy CONNECT HTTPError bad gateway": {
			err: &url.Error{
				Op:  "Get",
				URL: "https://updates.maxmind.com/geoip/updates/metadata?edition_id=GeoIP2-City",
				Err: HTTPError{StatusCode: http.StatusBadGateway},
			},
			want: true,
		},
		"plain forbidden error": {
			err:  errors.New("Forbidden"),
			want: true,
		},
		"nil": {
			err:  nil,
			want: false,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := IsRetryableError(tc.err)
			if tc.want != got {
				t.Errorf("expected retryable %v got %v", tc.want, got)
			}
		})
	}
}
