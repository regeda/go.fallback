package forecaster

import (
	"context"
	"errors"
	"time"

	fallback "github.com/regeda/go.fallback"
)

var Secondary Forecaster = forecasterFn(secondary)

func secondary(ctx context.Context) (*Response, error) {
	var out *Response

	p := fallback.NewPrimary()
	p.Go(func() (func(), error) {
		resp, err := accurate.Forecast(ctx)
		return func() {
			out = resp
		}, err
	})

	s := fallback.NewSecondary(p)
	s.Go(func() (func(), error) {
		resp, err := open.Forecast(ctx)
		return func() {
			out = resp
		}, err
	})

	if deadline, ok := ctx.Deadline(); ok {
		timer := time.AfterFunc(time.Until(deadline)/4, s.Shift)
		defer timer.Stop()
	}

	if s.Wait() {
		return out, nil
	}

	return nil, errors.New("nothing helped")
}
