package forecaster

import (
	"context"
	"time"
)

type Response struct {
	Name string
	Now  time.Time
}

type Forecaster interface {
	Forecast(context.Context) (*Response, error)
}

type forecasterFn func(context.Context) (*Response, error)

func (f forecasterFn) Forecast(ctx context.Context) (*Response, error) {
	return f(ctx)
}
