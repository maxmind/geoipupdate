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

		// It'd be preferable to use errors.As(), but that's only in Go 1.13+.
		underlyingErr := errors.Cause(err)
		httpError, ok := underlyingErr.(HTTPError)
		if ok &&
			httpError.StatusCode >= 400 &&
			httpError.StatusCode < 500 {
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
