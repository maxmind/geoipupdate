package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	{
		n, resp, err := testRetry(t, func(n int, causeError func()) { causeError() })
		assert.Equal(t, 5, n)
		assert.Error(t, err)
		assert.Nil(t, resp)
	}

	{
		n, resp, err := testRetry(
			t,
			func(n int, causeError func()) {
				if n < 3 {
					causeError()
				}
			},
		)
		assert.Equal(t, 3, n)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	}
}

func testRetry(t *testing.T, cb func(int, func())) (int, *http.Response, error) {
	var server *httptest.Server
	requests := 0
	server = httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, r *http.Request) {
				requests++
				cb(requests, func() { server.CloseClientConnections() })
			},
		),
	)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/error", nil)
	require.NoError(t, err)
	resp, err := MaybeRetryRequest(server.Client(), 5*time.Second, req)
	server.Close()
	return requests, resp, err
}
