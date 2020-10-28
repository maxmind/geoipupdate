// Package internal is none of your business
package internal

import (
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// MaybeRetryRequest is an internal implementation detail of this module. It
// shouldn't be used by users of the geoipupdate Go library. You can use the
// RetryFor field of geoipupdate.Config if you'd like to retry failed requests
// when using the library directly.
func MaybeRetryRequest(c *http.Client, retryFor time.Duration, req *http.Request) (*http.Response, error) {
	if retryFor < 0 {
		return nil, errors.New("negative retry duration")
	}
	if req.Body != nil {
		return nil, errors.New("can't retry requests with bodies")
	}
	var resp *http.Response
	var err error

	start := time.Now()
	for i := uint(0); ; i++ {
		resp, err = c.Do(req)
		if err == nil && resp.StatusCode < 500 {
			break
		}

		currentDuration := time.Since(start)

		waitDuration := 200 * time.Millisecond * (1 << i)
		if currentDuration+waitDuration > retryFor {
			break
		}
		if err == nil {
			_ = resp.Body.Close()
		}
		time.Sleep(waitDuration)
	}
	if err != nil {
		return nil, errors.Wrap(err, "error performing http request")
	}
	return resp, nil
}
