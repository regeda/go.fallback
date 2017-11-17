package fallback

import (
	"context"
	"sync"
)

type broadcastContext struct {
	context.Context
	once sync.Once
	done chan bool
}

func (ctx *broadcastContext) resolve(f func()) {
	ctx.once.Do(func() {
		f()
		ctx.done <- true
	})
}

func Primary(ctx context.Context, fn ...func(context.Context) error) (err error) {
	var (
		errOnce sync.Once
		wg      sync.WaitGroup
	)
	wg.Add(len(fn))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	broadcastCtx := broadcastContext{
		Context: ctx,
		done:    make(chan bool),
	}
	for _, f := range fn {
		f := f
		go func() {
			defer wg.Done()
			if ferr := f(&broadcastCtx); ferr != nil {
				errOnce.Do(func() {
					err = ferr
				})
			}
		}()
	}
	go func() {
		wg.Wait()
		close(broadcastCtx.done)
	}()
	if <-broadcastCtx.done {
		return nil
	}
	return
}
