package failover

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// MasterMaster executes all given handlers and emits a successful result from a faster master.
func MasterMaster(ctx context.Context, masters ...Handler) (err error) {
	doneCh := make(chan struct{})
	var (
		ok      bool
		outOnce sync.Once
		wg      errgroup.Group
	)
	for _, master := range masters {
		h := master
		wg.Go(func() error {
			herr, hdone := h(ctx)
			if herr == nil {
				outOnce.Do(func() {
					ok = true
					hdone()
					doneCh <- struct{}{}
				})
			}
			return herr
		})
	}
	go func() {
		werr := wg.Wait()
		if !ok {
			err = werr
		}
		close(doneCh)
	}()
	<-doneCh
	return
}
