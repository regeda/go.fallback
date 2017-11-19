package fallback

import (
	"context"
	"sync"
	"time"
)

// Secondary calls functions with a priority.
// A secondary function will be called if a primary function
// was failed or shift timeout was exceeded. But the successful
// result given from a primary function will be acquired
// in spite of a secondary execution.
func Secondary(ctx context.Context, shiftTimeout time.Duration, primary, secondary Func) error {
	shift := time.NewTimer(shiftTimeout)
	defer shift.Stop()
	var (
		once                       sync.Once
		primaryErr, secondaryErr   error
		primaryDone, secondaryDone func()
	)
	done := make(chan struct{})
	go func() {
		defer close(done)
		primaryErr, primaryDone = primary(ctx)
	}()
	secondaryRunner := func() {
		secondaryErr, secondaryDone = secondary(ctx)
	}
	for {
		select {
		case <-shift.C:
			go once.Do(secondaryRunner)
		case <-done:
			if primaryErr == nil {
				primaryDone()
				return nil
			}
			once.Do(secondaryRunner)
			if secondaryErr == nil {
				secondaryDone()
			}
			return secondaryErr
		}
	}
}
