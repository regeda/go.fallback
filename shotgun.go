package failover

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

func Shotgun(ctx context.Context, hh ...Handler) (err error) {
	var wg errgroup.Group
	doneCh := make(chan struct{})
	go func() {
		err = wg.Wait()
		close(doneCh)
	}()
	var outOnce sync.Once
	for _, h := range hh {
		h := h
		wg.Go(func() error {
			herr, hdone := h(ctx)
			if herr == nil {
				outOnce.Do(func() {
					hdone()
					doneCh <- struct{}{}
				})
			}
			return herr
		})
	}
	<-doneCh
	return
}
