package failover

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mock struct {
	t    *testing.T
	done func()
	err  error
}

func mocker(t *testing.T) *mock {
	return &mock{t: t}
}

func (m *mock) Done(s string, out *string) *mock {
	return &mock{
		t: m.t,
		done: func() {
			*out = s
		},
	}
}

func (m *mock) WithError(s string) *mock {
	return &mock{
		t:    m.t,
		done: m.done,
		err:  errors.New(s),
	}
}

func (m *mock) Count(count int) Handler {
	return func(context.Context) (error, func()) {
		if count == 0 {
			return nil, m.t.FailNow
		}
		count--
		return m.err, m.done
	}
}

func (m *mock) Sleep(sleep time.Duration) Handler {
	var fulfilled bool
	return func(context.Context) (error, func()) {
		if fulfilled {
			return nil, m.t.FailNow
		}
		fulfilled = true
		time.Sleep(sleep)
		return m.err, m.done
	}
}

func (m *mock) Context(wait time.Duration) Handler {
	return func(ctx context.Context) (error, func()) {
		select {
		case <-time.Tick(wait):
			return m.err, m.done
		case <-ctx.Done():
			return ctx.Err(), m.t.FailNow
		}
	}
}
