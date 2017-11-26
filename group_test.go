package fallback

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	delay = time.Second / 2
)

func TestSuccessfulPrimary(t *testing.T) {
	var n int

	p := NewPrimary()
	p.Go(func() error {
		n = 1
		return nil
	})

	require.True(t, p.Resolve())
	assert.Equal(t, 1, n)
}

func TestFailedPrimary(t *testing.T) {
	p := NewPrimary()
	p.Go(func() error {
		return assert.AnError
	})

	assert.False(t, p.Resolve())
}

func TestFailedPrimarySuccessfulSecondary(t *testing.T) {
	var n int

	p := NewPrimary()
	p.Go(func() error {
		return assert.AnError
	})

	s := NewSecondary(p)
	s.Go(func() error {
		n = 1
		return nil
	})

	require.True(t, s.Resolve())
	assert.Equal(t, 1, n)
}

func TestSuccessfulPrimaryIsBetter(t *testing.T) {
	var n int

	p := NewPrimary()
	p.Go(func() error {
		return assert.AnError
	})
	p.Go(func() error {
		n = 1
		return nil
	})

	require.True(t, p.Resolve())
	assert.Equal(t, 1, n)
}

func TestFailedPrimaryFailedSecondary(t *testing.T) {
	p := NewPrimary()
	p.Go(func() error {
		return assert.AnError
	})

	s := NewSecondary(p)
	s.Go(func() error {
		return assert.AnError
	})

	assert.False(t, s.Resolve())
}

func TestSecondaryNeverRunIfPrimaryCompleteSuccessfully(t *testing.T) {
	var n int

	p := NewPrimary()
	p.Go(func() error {
		n = 1
		return nil
	})

	s := NewSecondary(p)
	s.Go(func() error {
		t.FailNow()
		return nil
	})

	require.True(t, s.Resolve())
	assert.Equal(t, 1, n)
}

func TestPrimaryCompleteSuccessfullyNeverthelessShiftedSecondaryWellDone(t *testing.T) {
	var pn, sn int32

	p := NewPrimary()
	p.Go(func() error {
		time.Sleep(delay)
		atomic.AddInt32(&pn, 1)
		return nil
	})

	s := NewSecondary(p)
	s.Go(func() error {
		atomic.AddInt32(&sn, 1)
		return nil
	})

	s.Shift()

	require.True(t, s.Resolve())
	assert.EqualValues(t, 1, atomic.LoadInt32(&pn))
	assert.EqualValues(t, 1, atomic.LoadInt32(&sn))
}

func TestPrimaryCancelOther(t *testing.T) {
	var n int32

	p, ctx := NewPrimaryWithContext(context.Background())
	p.Go(func() error {
		time.Sleep(delay)
		assert.Equal(t, context.Canceled, ctx.Err())
		atomic.AddInt32(&n, 1)
		return ctx.Err()
	})
	p.Go(func() error {
		atomic.AddInt32(&n, 1)
		return nil
	})

	require.True(t, p.Resolve())
	assert.EqualValues(t, 2, atomic.LoadInt32(&n))
}

func TestSecondaryCancelOtherIfPrimaryFailed(t *testing.T) {
	var n int32

	p := NewPrimary()
	p.Go(func() error {
		atomic.AddInt32(&n, 1)
		return assert.AnError
	})

	s, ctx := NewSecondaryWithContext(context.Background(), p)
	s.Go(func() error {
		time.Sleep(delay)
		assert.Equal(t, context.Canceled, ctx.Err())
		atomic.AddInt32(&n, 1)
		return ctx.Err()
	})
	s.Go(func() error {
		atomic.AddInt32(&n, 1)
		return nil
	})

	require.True(t, s.Resolve())
	assert.EqualValues(t, 3, atomic.LoadInt32(&n))
}

func TestShiftedSecondaryCancelOtherIfPrimaryFailed(t *testing.T) {
	var n int32

	p := NewPrimary()
	p.Go(func() error {
		atomic.AddInt32(&n, 1)
		return assert.AnError
	})

	s, ctx := NewSecondaryWithContext(context.Background(), p)
	s.Go(func() error {
		time.Sleep(2 * delay)
		assert.Equal(t, context.Canceled, ctx.Err())
		atomic.AddInt32(&n, 1)
		return ctx.Err()
	})
	s.Go(func() error {
		time.Sleep(delay)
		atomic.AddInt32(&n, 1)
		return nil
	})

	s.Shift()

	require.True(t, s.Resolve())
	assert.EqualValues(t, 3, atomic.LoadInt32(&n))
}

func BenchmarkSuccessfulPrimary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewPrimary()
		p.Go(func() error {
			return nil
		})

		if !p.Resolve() {
			b.FailNow()
		}
	}
}

func BenchmarkFailedPrimary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewPrimary()
		p.Go(func() error {
			return assert.AnError
		})

		s := NewSecondary(p)
		s.Go(func() error {
			return nil
		})

		if !s.Resolve() {
			b.FailNow()
		}
	}
}
