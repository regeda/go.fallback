package fallback

import (
	"context"
	"fmt"
	"time"
)

// WeatherGather illustrates the use of Secondary for a result acquisition
// from a primary function in spite of a secondary execution.
func ExampleSecondary_weatherGather() {
	type weather struct {
		celsius float64
	}

	provider := func(w *weather, sleep time.Duration, celsius float64) Func {
		return func(context.Context) (error, func()) {
			time.Sleep(sleep)
			return nil, func() {
				w.celsius = celsius
			}
		}
	}

	var w weather

	err := Secondary(context.Background(),
		time.Second/2,
		provider(&w, time.Second, 21.1),     // slow provider with high priority
		provider(&w, time.Millisecond, -10), // fast provider with low priority
	)

	if err != nil {
		fmt.Println("error:", err)
	}

	fmt.Println(w.celsius)

	// Output:
	// 21.1
}
