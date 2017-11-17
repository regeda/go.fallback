# go.fallback

[![Build Status](https://travis-ci.org/regeda/go.failover.svg?branch=master)](https://travis-ci.org/regeda/go.failover)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.failover)](https://goreportcard.com/report/github.com/regeda/go.failover)

Fallback approaches and algorithms aimed to make your request stable and reliable.

## Primary

The approach emits successful response from a faster primary.

```go
func slow(ctx context.Context) error {
    time.Sleep(time.Second)
    fallback.Resolve(ctx, func() {
      fmt.Println("slow")
    })
    return nil
}

func fast(ctx context.Context) error {
    fallback.Resolve(ctx, func() {
      fmt.Println("fast")
    })
    return nil
}

err := fallback.Primary(context.Background(), slow, fast)

// console prints "fast"
```

## Secondary

The approach emits successful primary response otherwise a secondary result will be acquired.
A secondary can shift early before a primary complete a job. But if a primary was lucky then a secondary result will be declined anyway.
> You shouldn't care about locks in callback functions because they are thread-safe executed.

```go
func HandleAccuWeather(weather *Weather) func(context.Context) error {
    return func(ctx context.Context) error {
        resp, err := accuWeather.Forecast(ctx, AccuWeatherRequest())
        fallback.Resolve(ctx, func() {
            accuWeatcherResponseToWeather(resp, weather)
        })
        return err
    }
}

func HandleOpenWeather(weather *Weather) func(context.Context) error {
    return func(ctx context.Context) error {
        resp, err := openWeather.Forecast(ctx, OpenWeatherRequest())
        fallback.Resolve(ctx, func() {
            openWeatherResponseToWeather(resp, weather)
        })
        return err
    }
}

var weather Weather

err := fallback.Secondary(
    context.Background(),
    time.Second, // secondary shift time
    HandleAccuWeather(&weather), // primary
    HandleOpenWeather(&weather), // secondary
)
```

### Contributing
If you know another fallback approaches or algorithms then feel free to send them in a pull request. Unit tests are required.
