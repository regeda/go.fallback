package failover

import (
	"context"
	"sync"
	"time"
)

// MasterSlave returns successful master response otherwise a slave result will be acquired.
//
// A slave can shift early before a master complete a job.
// But if a master was lucky then a slave result will be omitted.
func MasterSlave(master, slave Requester, shiftTimeout time.Duration) Requester {
	return RequesterFunc(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		shift := time.NewTimer(shiftTimeout)
		defer shift.Stop()
		done := make(chan struct{})
		var (
			once sync.Once
			sout interface{}
			serr error
		)
		slaveRunner := func() {
			sout, serr = slave.Request(ctx, in)
		}
		go func() {
			defer close(done)
			out, err = master.Request(ctx, in)
		}()
		for {
			select {
			case <-shift.C:
				go once.Do(slaveRunner)
			case <-done:
				if err == nil {
					return
				}
				once.Do(slaveRunner)
				return sout, serr
			}
		}
	})
}
