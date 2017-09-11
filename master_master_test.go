package failover

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMasterMaster(t *testing.T) {
	t.Run("faster master should be better", func(t *testing.T) {
		var out string
		err := MasterMaster(
			context.Background(),
			mocker(t).Done("first", &out).Sleep(time.Millisecond),
			mocker(t).Done("second", &out).Sleep(time.Second/2),
		)
		require.Nil(t, err)
		assert.Equal(t, "first", out)
	})

	t.Run("execution was failed if all masters were failed", func(t *testing.T) {
		err := MasterMaster(
			context.Background(),
			mocker(t).WithError("first failed").Sleep(time.Millisecond),
			mocker(t).WithError("second failed").Sleep(time.Second/2),
		)
		require.NotNil(t, err)
		assert.EqualError(t, err, "first failed")
	})

	t.Run("lucky master should be better", func(t *testing.T) {
		var out string
		err := MasterMaster(
			context.Background(),
			mocker(t).WithError("first failed").Count(1),
			mocker(t).Done("second", &out).Sleep(time.Second/2),
		)
		require.Nil(t, err)
		assert.Equal(t, "second", out)
	})
}
