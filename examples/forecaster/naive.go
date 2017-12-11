package forecaster

import (
	"context"
	"fmt"
)

var Naive Forecaster = forecasterFn(naive)

func naive(ctx context.Context) (*Response, error) {
	resp, err := accurate.Forecast(ctx)
	if err != nil {
		fmt.Printf("accurate: %v, goto quick: ", err)
		return quick.Forecast(ctx)
	}
	return resp, err
}
