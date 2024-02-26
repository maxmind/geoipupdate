package internal

import (
	"net/http"
	"testing"

	"golang.org/x/net/http2"
)

func TestIsPermanentError(t *testing.T) {
	tt := map[string]struct {
		err  error
		want bool
	}{
		"HTTP2 INTERNAL_ERROR": {
			err: http2.StreamError{
				Code: http2.ErrCodeInternal,
			},
			want: false,
		},
		"bad gateway": {
			err: ResponseError{
				StatusCode: http.StatusBadGateway,
			},
			want: false,
		},
		"bad request": {
			err: ResponseError{
				StatusCode: http.StatusBadRequest,
			},
			want: true,
		},
		"nil": {
			err:  nil,
			want: false,
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			got := IsPermanentError(tc.err)
			if tc.want != got {
				t.Errorf("expected %v got %v", tc.want, got)
			}
		})
	}
}
