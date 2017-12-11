package forecaster

import (
	"context"
	"errors"

	fallback "github.com/regeda/go.fallback"
)

var Primary Forecaster = forecasterFn(primary)

func primary(ctx context.Context) (*Response, error) {
	var out *Response
	f := fallback.NewPrimary()
	f.Go(func() (func(), error) {
		resp, err := accurate.Forecast(ctx)
		return func() {
			out = resp
		}, err
	})
	f.Go(func() (func(), error) {
		resp, err := open.Forecast(ctx)
		return func() {
			out = resp
		}, err
	})
	if f.Wait() {
		return out, nil
	}
	return nil, errors.New("nothing helped")
}
