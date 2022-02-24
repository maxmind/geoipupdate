// Package internal is none of your business
package internal

import (
	"time"

	"github.com/pkg/errors"
)

// RetryWithBackoff calls the provided function repeatedly until it succeeds or
// until the retry duration is up.
func RetryWithBackoff(
	fn func() error,
	retryFor time.Duration,
) error {
	start := time.Now()

	for i := uint(0); ; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		underlyingErr := errors.Cause(err)
		var httpErr HTTPError
		//nolint:revive // suggests checking status code first, which is not valid.
		if errors.As(underlyingErr, &httpErr) &&
			httpErr.StatusCode >= 400 &&
			httpErr.StatusCode < 500 {
			return err
		}

		currentDuration := time.Since(start)

		waitDuration := 200 * time.Millisecond * (1 << i)

		if currentDuration+waitDuration > retryFor {
			return err
		}

		time.Sleep(waitDuration)
	}
}
