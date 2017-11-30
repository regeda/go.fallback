package fallback

import (
	"context"
	"sync"
)

// Group executes functions and resolves a result.
type Group interface {
	Go(func() error)
	Resolve() bool
}

// Primary executes functions and resolves the first non-error result.
type Primary struct {
	cancel context.CancelFunc

	wg sync.WaitGroup

	ok       bool
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
func (p *Primary) Go(f func() error) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		if err := f(); err == nil {
			p.doneOnce.Do(func() {
				p.ok = true
				if p.cancel != nil {
					p.cancel()
				}
			})
		}
	}()
}

// Resolve waits for functions result.
// Resolve fails if all functions returned an error.
func (p *Primary) Resolve() bool {
	p.wg.Wait()
	return p.ok
}

const (
	secondaryOpen = iota
	secondaryShift
	secondaryCancel
)

// Secondary executes functions if a primary failed or Shift command was performed.
type Secondary struct {
	primary Group

	cancel context.CancelFunc

	cond  *sync.Cond
	state int

	wg sync.WaitGroup

	ok       bool
	doneOnce sync.Once
}

// NewSecondary creates Secondary group with primary dependency.
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
	ctx, s.cancel = context.WithCancel(ctx)
	return s, ctx
}

// Go executes a function in a goroutine.
func (s *Secondary) Go(f func() error) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.cond.L.Lock()
		for s.state == secondaryOpen {
			s.cond.Wait()
		}
		if s.state&secondaryCancel == secondaryCancel {
			s.cond.L.Unlock()
			return
		}
		s.cond.L.Unlock()
		if err := f(); err == nil {
			s.doneOnce.Do(func() {
				s.ok = true
				if s.cancel != nil {
					s.cancel()
				}
			})
		}
	}()
}

// Resolve waits for primary result.
// If a primary failed, secondary functions will be performed.
// Otherwise secondary function will be canceled.
func (s *Secondary) Resolve() bool {
	if s.primary.Resolve() {
		go s.broadcast(secondaryCancel)
		return true
	}

	s.Shift()

	s.wg.Wait()

	return s.ok
}

// Shift run secondary functions without primary result waiting.
func (s *Secondary) Shift() {
	s.broadcast(secondaryShift)
}

func (s *Secondary) broadcast(state int) {
	s.cond.L.Lock()
	s.state |= state
	s.cond.L.Unlock()
	s.cond.Broadcast()
}
