package fallback

import (
	"context"
	"fmt"
	"time"
)

// Fastslow illustrates the use of Primary for a faster result acquisition
// from a collection of goroutines.
func ExamplePrimary_fastslow() {
	err := Primary(context.Background(),
		func(context.Context) (error, func()) {
			time.Sleep(time.Second)
			return nil, func() {
				fmt.Println("slow")
			}
		},
		func(context.Context) (error, func()) {
			return nil, func() {
				fmt.Println("fast")
			}
		})

	if err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// fast
}
