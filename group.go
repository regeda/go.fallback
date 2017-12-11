// Package fallback provides execution, failover synchronization
// and Context cancellation for a group of goroutines.
package fallback

import (
	"context"
	"sync"
)

// Func perfoms a task and returns an error or "done" function.
//
// "Done" function will be executed in thread-safe mode. There is you
// can do assignments in shared memory without locks or semaphores.
// Basically, "done" function is performed once.
type Func func() (func(), error)

// Group executes functions in goroutines and wait for a result.
type Group interface {
	Go(Func)
	Wait() bool
}

// Primary resolves the first non-error result.
type Primary struct {
	cancel context.CancelFunc

	wg sync.WaitGroup

	doneFn   func()
	doneOnce sync.Once
}

// NewPrimary creates Primary group
func NewPrimary() *Primary {
	return &Primary{}
}

// NewPrimaryWithContext creates Primary group with a context.
// If a successful result was obtained, other functions will be canceled immediately.
func NewPrimaryWithContext(ctx context.Context) (*Primary, context.Context) {
	p := NewPrimary()
	ctx, p.cancel = context.WithCancel(ctx)
	return p, ctx
}

// Go executes a function in a goroutine.
func (p *Primary) Go(f Func) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if doneFn, err := f(); err == nil {
			p.doneOnce.Do(func() {
				p.doneFn = doneFn
				if p.cancel != nil {
					p.cancel()
				}
			})
		}
	}()
}

// Wait blocks until all goroutines are completed.
// Wait fails if all functions returned an error.
func (p *Primary) Wait() bool {
	p.wg.Wait()
	if p.doneFn == nil {
		return false
	}
	p.doneFn()
	return true
}

const (
	open = iota
	shift
	cancel
)

// Secondary executes functions if a primary failed or Shift command was performed.
type Secondary struct {
	primary Group

	self Primary

	cond  *sync.Cond
	state int
}

// NewSecondary creates Secondary group on primary dependency.
func NewSecondary(primary Group) *Secondary {
	var m sync.Mutex
	return &Secondary{
		primary: primary,
		cond:    sync.NewCond(&m),
	}
}

// NewSecondaryWithContext creates Secondary group with a context.
// If a successful result was obtained, other functions will be canceled immediately.
func NewSecondaryWithContext(ctx context.Context, primary Group) (*Secondary, context.Context) {
	s := NewSecondary(primary)
	ctx, s.self.cancel = context.WithCancel(ctx)
	return s, ctx
}

// Go executes a function in a goroutine.
func (s *Secondary) Go(f Func) {
	s.self.Go(func() (func(), error) {
		s.cond.L.Lock()
		for s.state == open {
			s.cond.Wait()
		}
		if s.state&cancel == cancel {
			s.cond.L.Unlock()
			return nil, context.Canceled
		}
		s.cond.L.Unlock()
		return f()
	})
}

// Wait blocks until the primary is completed.
// If a primary failed, secondary functions will be performed.
// Otherwise secondary function will be canceled.
func (s *Secondary) Wait() bool {
	if s.primary.Wait() {
		if s.self.cancel != nil {
			s.self.cancel()
		}
		s.broadcast(cancel)
		return true
	}

	s.Shift()

	return s.self.Wait()
}

// Shift run secondary functions without primary result waiting.
func (s *Secondary) Shift() {
	s.broadcast(shift)
}

func (s *Secondary) broadcast(state int) {
	s.cond.L.Lock()
	s.state |= state
	s.cond.L.Unlock()
	s.cond.Broadcast()
}
