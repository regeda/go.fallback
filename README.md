# go.fallback

[![Build Status](https://travis-ci.org/regeda/go.fallback.svg?branch=master)](https://travis-ci.org/regeda/go.fallback)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.fallback)](https://goreportcard.com/report/github.com/regeda/go.fallback)

Fallback approaches and algorithms aimed to make your request stable and reliable.

## Primary

The approach emits successful response from a faster primary.

```go
func slow(context.Context) (error, func()) {
    time.Sleep(time.Second)
    return nil, func() {
        fmt.Println("slow")
    }
}

func fast(ctx context.Context) (error, func()) {
    return nil, func() {
        fmt.Println("fast")
    }
}

err := fallback.Primary(context.Background(), slow, fast)

// console prints "fast"
```

## Secondary

The approach emits successful primary response otherwise a secondary result will be acquired.
A secondary can shift early before a primary complete a job. But if a primary was lucky then a secondary result will be declined anyway.
> You shouldn't care about locks in callback functions because they are thread-safe executed.

```go
func HandleAccuWeather(weather *Weather) fallback.Func {
    return func(ctx context.Context) (error, func()) {
        resp, err := accuWeather.Forecast(ctx, AccuWeatherRequest())
        return err, func() {
            accuWeatcherResponseToWeather(resp, weather)
        }
    }
}

func HandleOpenWeather(weather *Weather) fallback.Func {
    return func(ctx context.Context) (error, func()) {
        resp, err := openWeather.Forecast(ctx, OpenWeatherRequest())
        return err, func() {
            openWeatherResponseToWeather(resp, weather)
        }
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
