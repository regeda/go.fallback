// Package forecaster allows to test Fallback approach.
//
// The main idea is to have two different providers with a balance between an accurate data and quick data retrieving.
// Usually, quick load gives dirty data. Dirty data allows to avoid unreliable response.
//
// It will work if a system was ready to take a low rate of dirty data. It's a sponsor requirement.
package forecaster

import (
	"context"
	"time"
)

var (
	// accurate works slowly
	accurate = fake{
		name: "accurate",
		d:    delay{0.2, time.Second},
	}
	// quick returns dirty data
	quick = fake{
		name: "quick",
		d:    delay{0.01, 200 * time.Millisecond},
	}
)

type fake struct {
	name string
	d    delay
}

func (s *fake) Forecast(ctx context.Context) (*Response, error) {
	respCh := make(chan Response)
	go func() {
		resp := Response{s.name, doNothing(s.d.delay())}
		select {
		case <-ctx.Done():
		case respCh <- resp:
		}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		return &resp, nil
	}
}

func doNothing(delay time.Duration) time.Time {
	if delay > 0 {
		time.Sleep(delay)
	}
	return time.Now()
}
