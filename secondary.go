package failover

import (
	"context"
	"sync"
	"time"
)

type primaryContext struct {
	context.Context
	done chan struct{}
	err  error
}

func Secondary(ctx context.Context, shiftTimeout time.Duration, primary, secondary func(context.Context) error) error {
	shift := time.NewTimer(shiftTimeout)
	defer shift.Stop()
	primaryCtx := primaryContext{
		Context: ctx,
		done:    make(chan struct{}),
	}
	var (
		secondaryErr error
		once         sync.Once
	)
	go func() {
		defer close(primaryCtx.done)
		primaryCtx.err = primary(ctx)
	}()
	secondaryRunner := func() {
		secondaryErr = secondary(&primaryCtx)
	}
	for {
		select {
		case <-shift.C:
			go once.Do(secondaryRunner)
		case <-primaryCtx.done:
			if primaryCtx.err == nil {
				return nil
			}
			once.Do(secondaryRunner)
			return secondaryErr
		}
	}
}

func Sync(ctx context.Context, resolve func()) {
	if primaryCtx, ok := ctx.(*primaryContext); ok {
		<-primaryCtx.done
		if primaryCtx.err == nil {
			return
		}
	}
	resolve()
}
