package failover

import (
	"context"
	"sync"
	"time"
)

// MasterSlave emits successful master response otherwise a slave result will be acquired.
//
// A slave can shift early before a master complete a job.
// But if a master was lucky then a slave result will be declined.
func MasterSlave(ctx context.Context, shiftTimeout time.Duration, master, slave Handler) error {
	shift := time.NewTimer(shiftTimeout)
	defer shift.Stop()
	doneCh := make(chan struct{})
	var (
		slaveOnce             sync.Once
		masterErr, slaveErr   error
		masterDone, slaveDone func()
	)
	slaveRunner := func() {
		slaveErr, slaveDone = slave(ctx)
	}
	go func() {
		defer close(doneCh)
		masterErr, masterDone = master(ctx)
	}()
	for {
		select {
		case <-shift.C:
			go slaveOnce.Do(slaveRunner)
		case <-doneCh:
			if masterErr == nil {
				masterDone()
				return nil
			}
			slaveOnce.Do(slaveRunner)
			if slaveErr == nil {
				slaveDone()
			}
			return slaveErr
		}
	}
}
