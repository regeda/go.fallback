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
			func(ctx context.Context) error {
				time.Sleep(time.Millisecond)
				Resolve(ctx, func() {
					out = "first"
				})
				return nil
			},
			func(ctx context.Context) error {
				time.Sleep(time.Second / 2)
				Resolve(ctx, func() {
					out = "second"
				})
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "first", out)
	})

	t.Run("execution was failed if all primaries were failed", func(t *testing.T) {
		err := Primary(
			context.Background(),
			func(context.Context) error {
				time.Sleep(time.Millisecond)
				return errors.New("first failed")
			},
			func(context.Context) error {
				time.Sleep(time.Second / 2)
				return errors.New("second failed")
			},
		)
		require.NotNil(t, err)
		assert.EqualError(t, err, "first failed")
	})

	t.Run("lucky primary should be better", func(t *testing.T) {
		var out string
		err := Primary(
			context.Background(),
			func(context.Context) error {
				return assert.AnError
			},
			func(ctx context.Context) error {
				time.Sleep(time.Second / 2)
				Resolve(ctx, func() {
					out = "second"
				})
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "second", out)
	})
}
