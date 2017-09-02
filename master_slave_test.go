package failover

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMasterSlave(t *testing.T) {
	mock := func(s string, err error, count int) RequesterFunc {
		return func(context.Context, interface{}) (interface{}, error) {
			if count == 0 {
				t.Fatalf("mock %s is zombie", s)
			}
			count--
			return s, err
		}
	}

	mocksleep := func(s string, err error, sleep time.Duration) RequesterFunc {
		var fulfilled bool
		return func(context.Context, interface{}) (interface{}, error) {
			if fulfilled {
				t.Fatalf("mock %s is zombie", s)
			}
			fulfilled = true
			time.Sleep(sleep)
			return s, err
		}
	}

	t.Run("master is faster than timeout", func(t *testing.T) {
		service := MasterSlave(
			mock("master", nil, 1),
			mock("slave", nil, 0),
			time.Second,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "master", out)
	})

	t.Run("slave should return if master failed", func(t *testing.T) {
		service := MasterSlave(
			mock("", errors.New("some shit"), 1),
			mock("slave", nil, 1),
			time.Second,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("slave failed after master failed", func(t *testing.T) {
		service := MasterSlave(
			mock("", errors.New("master: some shit"), 1),
			mock("", errors.New("slave: some shit"), 1),
			time.Second,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "slave: some shit")
		require.Empty(t, out)
	})

	t.Run("slave should run once if master failed", func(t *testing.T) {
		service := MasterSlave(
			mock("", errors.New("master: some shit"), 1),
			mocksleep("slave", nil, time.Second),
			time.Second,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("master should return with some delay", func(t *testing.T) {
		service := MasterSlave(
			mocksleep("master", nil, time.Second),
			mock("slave", nil, 1),
			time.Millisecond,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "master", out)
	})

	t.Run("slave should return with some delay if master failed", func(t *testing.T) {
		service := MasterSlave(
			mock("", errors.New("some shit"), 1),
			mocksleep("slave", nil, time.Second),
			time.Second/2,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("slave should return if master failed after long delay", func(t *testing.T) {
		service := MasterSlave(
			mocksleep("", errors.New("some shit"), time.Second),
			mock("slave", nil, 1),
			time.Second/2,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("slave should return with delay if master failed after long delay", func(t *testing.T) {
		service := MasterSlave(
			mocksleep("", errors.New("some shit"), time.Second),
			mocksleep("slave", nil, 2*time.Second),
			time.Millisecond,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("error if global context failed", func(t *testing.T) {
		service := MasterSlave(
			mocksleep("master", nil, time.Second),
			mock("slave", nil, 0),
			time.Second,
		)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()
		out, err := service.Request(ctx, struct{}{})
		assert.NotNil(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.Empty(t, out)
	})

	t.Run("master should return regardless slave slow exec", func(t *testing.T) {
		service := MasterSlave(
			mocksleep("master", nil, time.Second),
			mocksleep("slave", nil, 2*time.Second),
			time.Second/2,
		)
		out, err := service.Request(context.Background(), struct{}{})
		require.Nil(t, err)
		assert.Equal(t, "master", out)
	})
}
