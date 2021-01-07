package fallback_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	fallback "github.com/regeda/go.fallback"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	delay = time.Second / 2
)

func noopFn() {}

func TestSuccessfulPrimary(t *testing.T) {
	var n int

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		return func() {
			n = 1
		}, nil
	})

	require.True(t, p.Wait())
	assert.Equal(t, 1, n)
}

func TestFailedPrimary(t *testing.T) {
	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		return t.FailNow, assert.AnError
	})

	assert.False(t, p.Wait())
}

func TestSuccessfulPrimaryIsBetter(t *testing.T) {
	var n int

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		return t.FailNow, assert.AnError
	})
	p.Go(func() (func(), error) {
		return func() {
			n = 1
		}, nil
	})

	require.True(t, p.Wait())
	assert.Equal(t, 1, n)
}

func TestFailedPrimarySuccessfulSecondary(t *testing.T) {
	var n int

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		return t.FailNow, assert.AnError
	})

	s := fallback.NewSecondary(p)
	s.Go(func() (func(), error) {
		return func() {
			n = 1
		}, nil
	})

	require.True(t, s.Wait())
	assert.Equal(t, 1, n)
}

func TestFailedPrimaryFailedSecondary(t *testing.T) {
	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		return t.FailNow, assert.AnError
	})

	s := fallback.NewSecondary(p)
	s.Go(func() (func(), error) {
		return t.FailNow, assert.AnError
	})

	assert.False(t, s.Wait())
}

func TestSecondaryNeverRunIfPrimaryCompleteSuccessfully(t *testing.T) {
	var n int

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		return func() {
			n = 1
		}, nil
	})

	s := fallback.NewSecondary(p)
	s.Go(func() (func(), error) {
		t.FailNow()
		return noopFn, nil
	})

	require.True(t, s.Wait())
	assert.Equal(t, 1, n)
}

func TestPrimaryCompleteSuccessfullyNeverthelessShiftedSecondaryWellDone(t *testing.T) {
	var (
		pn, sn int32
		result string
	)

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		time.Sleep(delay)
		atomic.AddInt32(&pn, 1)
		return func() {
			result = "primary"
		}, nil
	})

	s := fallback.NewSecondary(p)
	s.Go(func() (func(), error) {
		atomic.AddInt32(&sn, 1)
		return func() {
			result = "secondary"
		}, nil
	})

	s.Shift()

	require.True(t, s.Wait())
	assert.EqualValues(t, 1, atomic.LoadInt32(&pn))
	assert.EqualValues(t, 1, atomic.LoadInt32(&sn))
	assert.Equal(t, "primary", result)
}

func TestPrimaryCancelOther(t *testing.T) {
	var n int32

	p, ctx := fallback.NewPrimaryWithContext(context.TODO())
	p.Go(func() (func(), error) {
		atomic.AddInt32(&n, 1)
		select {
		case <-time.After(delay):
			t.FailNow()
			return noopFn, nil
		case <-ctx.Done():
			return t.FailNow, ctx.Err()
		}
	})
	p.Go(func() (func(), error) {
		atomic.AddInt32(&n, 1)
		return noopFn, nil
	})

	require.True(t, p.Wait())
	assert.EqualValues(t, 2, atomic.LoadInt32(&n))
}

func TestSecondaryCancelOtherIfPrimaryFailed(t *testing.T) {
	var n int32

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		atomic.AddInt32(&n, 1)
		return t.FailNow, assert.AnError
	})

	s, ctx := fallback.NewSecondaryWithContext(context.TODO(), p)
	s.Go(func() (func(), error) {
		atomic.AddInt32(&n, 1)
		select {
		case <-time.After(delay):
			t.FailNow()
			return noopFn, nil
		case <-ctx.Done():
			return t.FailNow, ctx.Err()
		}
	})
	s.Go(func() (func(), error) {
		atomic.AddInt32(&n, 1)
		return noopFn, nil
	})

	require.True(t, s.Wait())
	assert.EqualValues(t, 3, atomic.LoadInt32(&n))
}

func TestShiftedSecondaryCancelOtherIfPrimaryFailed(t *testing.T) {
	var n int32

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		atomic.AddInt32(&n, 1)
		return t.FailNow, assert.AnError
	})

	s, ctx := fallback.NewSecondaryWithContext(context.TODO(), p)
	s.Go(func() (func(), error) {
		atomic.AddInt32(&n, 1)
		select {
		case <-time.After(2 * delay):
			t.FailNow()
			return noopFn, nil
		case <-ctx.Done():
			return t.FailNow, ctx.Err()
		}
	})
	s.Go(func() (func(), error) {
		time.Sleep(delay)
		atomic.AddInt32(&n, 1)
		return noopFn, nil
	})

	s.Shift()

	require.True(t, s.Wait())
	assert.EqualValues(t, 3, atomic.LoadInt32(&n))
}

func TestShiftedSecondaryShouldBeCanceledIfPrimarySuccessfullyCompleted(t *testing.T) {
	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		return noopFn, nil
	})

	s, ctx := fallback.NewSecondaryWithContext(context.TODO(), p)
	s.Go(func() (func(), error) {
		select {
		case <-time.After(delay):
			t.FailNow()
			return noopFn, nil
		case <-ctx.Done():
			return t.FailNow, ctx.Err()
		}
	})

	s.Shift()

	require.True(t, s.Wait())

	select {
	case <-time.After(5 * delay):
		t.Fatalf("secondary should be canceled")
	case <-ctx.Done():
	}
}

func BenchmarkPrimary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := fallback.NewPrimary()
		p.Go(func() (func(), error) {
			return noopFn, nil
		})

		if !p.Wait() {
			b.FailNow()
		}
	}
}

func BenchmarkPrimaryWithCanceledSecondary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := fallback.NewPrimary()
		p.Go(func() (func(), error) {
			return noopFn, nil
		})

		s := fallback.NewSecondary(p)
		s.Go(func() (func(), error) {
			b.FailNow()
			return noopFn, nil
		})

		if !s.Wait() {
			b.FailNow()
		}
	}
}

func BenchmarkSecondaryWithFailedPrimary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := fallback.NewPrimary()
		p.Go(func() (func(), error) {
			return b.FailNow, assert.AnError
		})

		s := fallback.NewSecondary(p)
		s.Go(func() (func(), error) {
			return noopFn, nil
		})

		if !s.Wait() {
			b.FailNow()
		}
	}
}
