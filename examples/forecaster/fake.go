package forecaster

import (
	"context"
	"time"
)

var (
	accurate = fake{
		name: "accurate",
		d:    delay{0.2, time.Second},
	}
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
