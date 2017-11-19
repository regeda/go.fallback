package fallback

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrimary(t *testing.T) {
	t.Run("faster primary should be better", func(t *testing.T) {
		var out string
		err := Primary(
			context.Background(),
			func(context.Context) (error, func()) {
				return nil, func() {
					out = "first"
				}
			},
			func(context.Context) (error, func()) {
				time.Sleep(time.Second / 2)
				return nil, func() {
					out = "second"
				}
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "first", out)
	})

	t.Run("execution was failed if all primaries were failed", func(t *testing.T) {
		err := Primary(
			context.Background(),
			func(context.Context) (error, func()) {
				return errors.New("first failed"), t.FailNow
			},
			func(context.Context) (error, func()) {
				time.Sleep(time.Second / 2)
				return errors.New("second failed"), t.FailNow
			},
		)
		require.NotNil(t, err)
		assert.EqualError(t, err, "first failed")
	})

	t.Run("lucky primary should be better", func(t *testing.T) {
		var out string
		err := Primary(
			context.Background(),
			func(context.Context) (error, func()) {
				return assert.AnError, t.FailNow
			},
			func(context.Context) (error, func()) {
				time.Sleep(time.Second / 2)
				return nil, func() {
					out = "second"
				}
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "second", out)
	})
}
