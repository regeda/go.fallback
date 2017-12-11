package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/regeda/go.fallback/example/forecaster"
)

var approach = flag.String("approach", "", "naive|primary|secondary")

func main() {
	flag.Parse()

	ctx := context.Background()

	s := forecasterFactory(*approach)

	for {
		exec(ctx, s)
		time.Sleep(100 * time.Millisecond)
	}
}

func forecasterFactory(approach string) forecaster.Forecaster {
	switch approach {
	case "naive":
		return forecaster.Naive
	case "primary":
		return forecaster.Primary
	case "secondary":
		return forecaster.Secondary
	}
	panic(fmt.Sprintf("unknown \"%s\" forecaster", approach))
}

func exec(ctx context.Context, s forecaster.Forecaster) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	resp, err := s.Forecast(ctx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%s: OK\n", resp.Name)
	}
}
