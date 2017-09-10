package failover

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMasterSlave(t *testing.T) {
	t.Run("master is faster than timeout", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Second,
			mocker(t).Done("master", &out).Count(1),
			mocker(t).Done("slave", &out).Count(0),
		)
		require.Nil(t, err)
		assert.Equal(t, "master", out)
	})

	t.Run("slave should return if master failed", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Second,
			mocker(t).Done("master", &out).WithError("some shit").Count(1),
			mocker(t).Done("slave", &out).Count(1),
		)
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("slave failed after master failed", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Second,
			mocker(t).Done("master", &out).WithError("master: some shit").Count(1),
			mocker(t).Done("slave", &out).WithError("slave: some shit").Count(1),
		)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "slave: some shit")
		require.Empty(t, out)
	})

	t.Run("slave should run once if master failed", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Second,
			mocker(t).Done("master", &out).WithError("master: some shit").Count(1),
			mocker(t).Done("slave", &out).Sleep(time.Second),
		)
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("master should return with some delay", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Millisecond,
			mocker(t).Done("master", &out).Sleep(time.Second),
			mocker(t).Done("slave", &out).Count(1),
		)
		require.Nil(t, err)
		assert.Equal(t, "master", out)
	})

	t.Run("slave should return with some delay if master failed", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Second/2,
			mocker(t).Done("master", &out).WithError("some shit").Count(1),
			mocker(t).Done("slave", &out).Sleep(time.Second),
		)
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("slave should return if master failed after long delay", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Second/2,
			mocker(t).Done("master", &out).WithError("some shit").Sleep(time.Second),
			mocker(t).Done("slave", &out).Count(1),
		)
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("slave should return with delay if master failed after long delay", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Millisecond,
			mocker(t).Done("master", &out).WithError("some shit").Sleep(time.Second),
			mocker(t).Done("slave", &out).Sleep(2*time.Second),
		)
		require.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("shifted slave should return if global context deadlined", func(t *testing.T) {
		var out string
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()
		err := MasterSlave(
			ctx,
			time.Millisecond,
			mocker(t).Done("master", &out).Context(time.Second),
			mocker(t).Done("slave", &out).Count(1),
		)
		assert.Nil(t, err)
		assert.Equal(t, "slave", out)
	})

	t.Run("error if global context deadlined", func(t *testing.T) {
		var out string
		ctx, cancel := context.WithTimeout(context.Background(), time.Second/2)
		defer cancel()
		err := MasterSlave(
			ctx,
			time.Millisecond,
			mocker(t).Done("master", &out).Context(time.Second),
			mocker(t).Done("slave", &out).Context(time.Second),
		)
		assert.NotNil(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.Empty(t, out)
	})

	t.Run("master should return regardless slave slow exec", func(t *testing.T) {
		var out string
		err := MasterSlave(
			context.Background(),
			time.Second/2,
			mocker(t).Done("master", &out).Sleep(time.Second),
			mocker(t).Done("slave", &out).Sleep(2*time.Second),
		)
		require.Nil(t, err)
		assert.Equal(t, "master", out)
	})
}
