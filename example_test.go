package fallback

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func ExamplePrimary_successfullyCompleted() {
	p := NewPrimary()
	p.Go(func() error {
		return errors.New("broken")
	})
	p.Go(func() error {
		return nil
	})

	if p.Wait() {
		fmt.Println("successfully completed")
	}
	// Output:
	// successfully completed
}

func ExamplePrimary_withContext() {
	p, ctx := NewPrimaryWithContext(context.Background())
	p.Go(func() error {
		fmt.Println("the first is good")
		return nil
	})
	p.Go(func() error {
		time.Sleep(time.Second)
		if ctx.Err() == context.Canceled {
			fmt.Println("the second is canceled")
		}
		return nil
	})

	if p.Wait() {
		fmt.Println("successfully completed")
	}
	// Output:
	// the first is good
	// the second is canceled
	// successfully completed
}

func ExampleSecondary_successfullyCompletedIfPrimaryFailed() {
	p := NewPrimary()
	p.Go(func() error {
		fmt.Println("primary is broken")
		return errors.New("broken")
	})

	s := NewSecondary(p)
	s.Go(func() error {
		return errors.New("broken")
	})
	s.Go(func() error {
		fmt.Println("secondary helps")
		return nil
	})

	if s.Wait() {
		fmt.Println("successfully completed")
	}
	// Output:
	// primary is broken
	// secondary helps
	// successfully completed
}

func ExampleSecondary_shiftBeforePrimaryFailed() {
	p := NewPrimary()
	p.Go(func() error {
		time.Sleep(time.Second)
		fmt.Println("primary is broken")
		return errors.New("broken")
	})

	s := NewSecondary(p)
	s.Go(func() error {
		fmt.Println("secondary helps")
		return nil
	})

	s.Shift() // start a secondary immediately

	if s.Wait() {
		fmt.Println("successfully completed")
	}

	// Output:
	// secondary helps
	// primary is broken
	// successfully completed
}
