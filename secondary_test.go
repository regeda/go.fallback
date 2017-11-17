package fallback

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecondary(t *testing.T) {
	failOrExit := func(ctx context.Context) error {
		select {
		case <-time.Tick(time.Second):
			t.FailNow()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	t.Run("primary is faster than timeout", func(t *testing.T) {
		var out string
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) error {
				out = "primary"
				return nil
			},
			func(context.Context) error {
				t.FailNow()
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "primary", out)
	})

	t.Run("secondary should proceed if primary failed", func(t *testing.T) {
		var out string
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) error {
				return assert.AnError
			},
			func(context.Context) error {
				out = "secondary"
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "secondary", out)
	})

	t.Run("secondary failed after primary failed", func(t *testing.T) {
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) error {
				return errors.New("primary: mortal combat")
			},
			func(context.Context) error {
				return errors.New("secondary: mortal combat")
			},
		)
		require.NotNil(t, err)
		assert.EqualError(t, err, "secondary: mortal combat")
	})

	t.Run("secondary should run once if primary failed", func(t *testing.T) {
		var (
			out     string
			counter int32
		)
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) error {
				return assert.AnError
			},
			func(context.Context) error {
				atomic.AddInt32(&counter, 1)
				time.Sleep(time.Second)
				out = "secondary"
				return nil
			},
		)
		require.Nil(t, err)
		require.Equal(t, int32(1), atomic.LoadInt32(&counter))
		assert.Equal(t, "secondary", out)
	})

	t.Run("primary should return successful result after some delay", func(t *testing.T) {
		var out string
		err := Secondary(
			context.Background(),
			time.Millisecond,
			func(ctx context.Context) error {
				time.Sleep(time.Second / 2)
				Resolve(ctx, func() {
					out = "primary"
				})
				return nil
			},
			func(ctx context.Context) error {
				Resolve(ctx, t.FailNow)
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "primary", out)
	})

	t.Run("primary should proceed regardless secondary result", func(t *testing.T) {
		var out string
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(ctx context.Context) error {
				Resolve(ctx, func() {
					out = "primary"
				})
				time.Sleep(time.Second)
				return nil
			},
			func(ctx context.Context) error {
				Resolve(ctx, t.FailNow)
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "primary", out)
	})

	t.Run("secondary should proceed if primary failed after long delay", func(t *testing.T) {
		var out string
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) error {
				time.Sleep(time.Second)
				return assert.AnError
			},
			func(context.Context) error {
				out = "secondary"
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "secondary", out)
	})

	t.Run("secondary should proceed with delay if primary failed after long delay", func(t *testing.T) {
		var out string
		err := Secondary(
			context.Background(),
			time.Millisecond,
			func(context.Context) error {
				time.Sleep(time.Second)
				return assert.AnError
			},
			func(context.Context) error {
				time.Sleep(2 * time.Second)
				out = "secondary"
				return nil
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "secondary", out)
	})

	t.Run("shifted secondary should proceed if global context was deadlined", func(t *testing.T) {
		var out string
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()
		err := Secondary(
			ctx,
			time.Millisecond,
			failOrExit,
			func(context.Context) error {
				out = "secondary"
				return nil
			},
		)
		assert.Nil(t, err)
		assert.Equal(t, "secondary", out)
	})

	t.Run("error if global context deadlined", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()
		err := Secondary(
			ctx,
			time.Millisecond,
			failOrExit,
			failOrExit,
		)
		assert.NotNil(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("primary should proceed regardless secondary slow exec", func(t *testing.T) {
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) error {
				time.Sleep(time.Second)
				return nil
			},
			func(ctx context.Context) error {
				time.Sleep(2 * time.Second)
				Resolve(ctx, t.FailNow)
				return nil
			},
		)
		require.Nil(t, err)
	})
}
