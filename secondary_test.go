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
	failOrExit := func(ctx context.Context) (error, func()) {
		select {
		case <-time.Tick(time.Second):
			return nil, t.FailNow
		case <-ctx.Done():
			return ctx.Err(), t.FailNow
		}
	}

	t.Run("primary is faster than timeout", func(t *testing.T) {
		var out string
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) (error, func()) {
				return nil, func() {
					out = "primary"
				}
			},
			func(context.Context) (error, func()) {
				return nil, t.FailNow
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
			func(context.Context) (error, func()) {
				return assert.AnError, t.FailNow
			},
			func(context.Context) (error, func()) {
				return nil, func() {
					out = "secondary"
				}
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "secondary", out)
	})

	t.Run("secondary failed after primary failed", func(t *testing.T) {
		err := Secondary(
			context.Background(),
			time.Second/2,
			func(context.Context) (error, func()) {
				return errors.New("primary: mortal combat"), t.FailNow
			},
			func(context.Context) (error, func()) {
				return errors.New("secondary: mortal combat"), t.FailNow
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
			func(context.Context) (error, func()) {
				return assert.AnError, t.FailNow
			},
			func(context.Context) (error, func()) {
				atomic.AddInt32(&counter, 1)
				time.Sleep(time.Second)
				return nil, func() {
					out = "secondary"
				}
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
			func(context.Context) (error, func()) {
				time.Sleep(time.Second / 2)
				return nil, func() {
					out = "primary"
				}
			},
			func(context.Context) (error, func()) {
				return nil, t.FailNow
			},
		)
		require.Nil(t, err)
		assert.Equal(t, "primary", out)
	})

	t.Run("shifted secondary should proceed if global context was deadlined", func(t *testing.T) {
		var out string
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()
		err := Secondary(
			ctx,
			time.Millisecond,
			failOrExit,
			func(context.Context) (error, func()) {
				return nil, func() {
					out = "secondary"
				}
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
}
