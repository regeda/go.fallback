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

	primary := fallback.NewPrimary()
	primary.Go(func() (func(), error) {
		resp, err := accurate.Forecast(ctx)
		return func() {
			out = resp
		}, err
	})

	secondary := fallback.NewSecondary(primary)
	secondary.Go(func() (func(), error) {
		resp, err := quick.Forecast(ctx)
		return func() {
			out = resp
		}, err
	})

	if deadline, ok := ctx.Deadline(); ok {
		timer := time.AfterFunc(time.Until(deadline)/4, secondary.Shift)
		defer timer.Stop()
	}

	if secondary.Wait() {
		return out, nil
	}

	return nil, errors.New("nothing helped")
}
