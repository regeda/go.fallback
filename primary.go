package fallback

import (
	"context"
	"sync"
)

// Primary calls functions in goroutines and acquires the first successful result.
// If a result is loaded, all remaining functions will be canceled.
func Primary(ctx context.Context, fn ...Func) (err error) {
	var (
		wg                sync.WaitGroup
		errOnce, doneOnce sync.Once
	)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan bool)
	wg.Add(len(fn))
	for _, f := range fn {
		f := f
		go func() {
			defer wg.Done()
			ferr, fdone := f(ctx)
			if ferr == nil {
				doneOnce.Do(func() {
					fdone()
					done <- true
				})
			} else {
				errOnce.Do(func() {
					err = ferr
				})
			}
		}()
	}
	go func() {
		wg.Wait()
		close(done)
	}()
	if <-done {
		return nil
	}
	return
}
