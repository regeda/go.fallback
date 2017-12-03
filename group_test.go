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
	p.Go(func() (error, func()) {
		return nil, func() {
			n = 1
		}
	})

	require.True(t, p.Wait())
	assert.Equal(t, 1, n)
}

func TestFailedPrimary(t *testing.T) {
	p := NewPrimary()
	p.Go(func() (error, func()) {
		return assert.AnError, t.FailNow
	})

	assert.False(t, p.Wait())
}

func TestSuccessfulPrimaryIsBetter(t *testing.T) {
	var n int

	p := NewPrimary()
	p.Go(func() (error, func()) {
		return assert.AnError, t.FailNow
	})
	p.Go(func() (error, func()) {
		return nil, func() {
			n = 1
		}
	})

	require.True(t, p.Wait())
	assert.Equal(t, 1, n)
}

func TestFailedPrimarySuccessfulSecondary(t *testing.T) {
	var n int

	p := NewPrimary()
	p.Go(func() (error, func()) {
		return assert.AnError, t.FailNow
	})

	s := NewSecondary(p)
	s.Go(func() (error, func()) {
		return nil, func() {
			n = 1
		}
	})

	require.True(t, s.Wait())
	assert.Equal(t, 1, n)
}

func TestFailedPrimaryFailedSecondary(t *testing.T) {
	p := NewPrimary()
	p.Go(func() (error, func()) {
		return assert.AnError, t.FailNow
	})

	s := NewSecondary(p)
	s.Go(func() (error, func()) {
		return assert.AnError, t.FailNow
	})

	assert.False(t, s.Wait())
}

func TestSecondaryNeverRunIfPrimaryCompleteSuccessfully(t *testing.T) {
	var n int

	p := NewPrimary()
	p.Go(func() (error, func()) {
		return nil, func() {
			n = 1
		}
	})

	s := NewSecondary(p)
	s.Go(func() (error, func()) {
		t.FailNow()
		return nil, NoopFunc
	})

	require.True(t, s.Wait())
	assert.Equal(t, 1, n)
}

func TestPrimaryCompleteSuccessfullyNeverthelessShiftedSecondaryWellDone(t *testing.T) {
	var (
		pn, sn int32
		result string
	)

	p := NewPrimary()
	p.Go(func() (error, func()) {
		time.Sleep(delay)
		atomic.AddInt32(&pn, 1)
		return nil, func() {
			result = "primary"
		}
	})

	s := NewSecondary(p)
	s.Go(func() (error, func()) {
		atomic.AddInt32(&sn, 1)
		return nil, func() {
			result = "secondary"
		}
	})

	s.Shift()

	require.True(t, s.Wait())
	assert.EqualValues(t, 1, atomic.LoadInt32(&pn))
	assert.EqualValues(t, 1, atomic.LoadInt32(&sn))
	assert.Equal(t, "primary", result)
}

func TestPrimaryCancelOther(t *testing.T) {
	var n int32

	p, ctx := NewPrimaryWithContext(context.Background())
	p.Go(func() (error, func()) {
		time.Sleep(delay)
		assert.Equal(t, context.Canceled, ctx.Err())
		atomic.AddInt32(&n, 1)
		return ctx.Err(), t.FailNow
	})
	p.Go(func() (error, func()) {
		atomic.AddInt32(&n, 1)
		return nil, NoopFunc
	})

	require.True(t, p.Wait())
	assert.EqualValues(t, 2, atomic.LoadInt32(&n))
}

func TestSecondaryCancelOtherIfPrimaryFailed(t *testing.T) {
	var n int32

	p := NewPrimary()
	p.Go(func() (error, func()) {
		atomic.AddInt32(&n, 1)
		return assert.AnError, t.FailNow
	})

	s, ctx := NewSecondaryWithContext(context.Background(), p)
	s.Go(func() (error, func()) {
		time.Sleep(delay)
		assert.Equal(t, context.Canceled, ctx.Err())
		atomic.AddInt32(&n, 1)
		return ctx.Err(), t.FailNow
	})
	s.Go(func() (error, func()) {
		atomic.AddInt32(&n, 1)
		return nil, NoopFunc
	})

	require.True(t, s.Wait())
	assert.EqualValues(t, 3, atomic.LoadInt32(&n))
}

func TestShiftedSecondaryCancelOtherIfPrimaryFailed(t *testing.T) {
	var n int32

	p := NewPrimary()
	p.Go(func() (error, func()) {
		atomic.AddInt32(&n, 1)
		return assert.AnError, t.FailNow
	})

	s, ctx := NewSecondaryWithContext(context.Background(), p)
	s.Go(func() (error, func()) {
		time.Sleep(2 * delay)
		assert.Equal(t, context.Canceled, ctx.Err())
		atomic.AddInt32(&n, 1)
		return ctx.Err(), t.FailNow
	})
	s.Go(func() (error, func()) {
		time.Sleep(delay)
		atomic.AddInt32(&n, 1)
		return nil, NoopFunc
	})

	s.Shift()

	require.True(t, s.Wait())
	assert.EqualValues(t, 3, atomic.LoadInt32(&n))
}

func BenchmarkPrimary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewPrimary()
		p.Go(func() (error, func()) {
			return nil, NoopFunc
		})

		if !p.Wait() {
			b.FailNow()
		}
	}
}

func BenchmarkPrimaryWithCanceledSecondary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewPrimary()
		p.Go(func() (error, func()) {
			return nil, NoopFunc
		})

		s := NewSecondary(p)
		s.Go(func() (error, func()) {
			b.FailNow()
			return nil, NoopFunc
		})

		if !s.Wait() {
			b.FailNow()
		}
	}
}

func BenchmarkSecondaryWithFailedPrimary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewPrimary()
		p.Go(func() (error, func()) {
			return assert.AnError, b.FailNow
		})

		s := NewSecondary(p)
		s.Go(func() (error, func()) {
			return nil, NoopFunc
		})

		if !s.Wait() {
			b.FailNow()
		}
	}
}
