package internal

import (
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestRetryWithBackoff(t *testing.T) {
	t.Run("never succeeds", func(t *testing.T) {
		var n int
		err := RetryWithBackoff(
			func() error {
				n++
				return errors.New("foo")
			},
			6*time.Second,
		)
		assert.Equal(t, 5, n)
		assert.Error(t, err)
	})

	t.Run("succeeds after failures", func(t *testing.T) {
		var n int
		err := RetryWithBackoff(
			func() error {
				n++
				if n < 3 {
					return errors.New("foo")
				}
				return nil
			},
			6*time.Second,
		)
		assert.Equal(t, 3, n)
		assert.NoError(t, err)
	})
}
