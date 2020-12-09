// Package internal is none of your business
package internal

import (
	"time"
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

		currentDuration := time.Since(start)

		waitDuration := 200 * time.Millisecond * (1 << i)

		if currentDuration+waitDuration > retryFor {
			return err
		}

		time.Sleep(waitDuration)
	}
}
