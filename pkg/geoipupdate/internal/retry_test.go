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
		n, resp, err := testRetry(
			t,
			func(n int, causeError func()) int {
				causeError()
				return http.StatusOK
			},
		) // nolint: bodyclose
		assert.Equal(t, 5, n)
		assert.Error(t, err)
		assert.Nil(t, resp)
	}

	{
		n, resp, err := testRetry(
			t,
			func(n int, causeError func()) int {
				if n < 3 {
					causeError()
				}
				return http.StatusOK
			},
		)
		assert.Equal(t, 3, n)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		require.NoError(t, resp.Body.Close())
	}

	{
		n, resp, err := testRetry(
			t,
			func(int, func()) int { return http.StatusInternalServerError },
		)
		assert.Equal(t, 5, n)
		assert.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	}
}

func testRetry(t *testing.T, cb func(int, func()) int) (int, *http.Response, error) {
	var server *httptest.Server
	requests := 0
	server = httptest.NewServer(
		http.HandlerFunc(
			func(rw http.ResponseWriter, r *http.Request) {
				requests++
				rw.WriteHeader(cb(requests, func() { server.CloseClientConnections() }))
			},
		),
	)

	req, err := http.NewRequest(http.MethodGet, server.URL, nil) // nolint: noctx
	require.NoError(t, err)
	resp, err := MaybeRetryRequest(server.Client(), 6*time.Second, req)
	server.Close()
	return requests, resp, err
}
