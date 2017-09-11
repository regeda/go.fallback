# go.failover

[![Build Status](https://travis-ci.org/regeda/go.failover.svg?branch=master)](https://travis-ci.org/regeda/go.failover)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.failover)](https://goreportcard.com/report/github.com/regeda/go.failover)

Failover approaches and algorithms aimed to make your request stable and reliable.

## Master-Master

The approach emits successful response from a faster master.

```go
func slow(context.Context) (error, func()) {
    time.Sleep(time.Second)
    return nil, func() {
        fmt.Println("slow")
    }
}

func fast(context.Context) (error, func()) {
    return nil, func() {
        fmt.Println("fast")
    }
}

err := failover.MasterMaster(
    context.Background(),
    failover.Handler(slow),
    failover.Handler(fast),
)

// console prints "fast"
```

## Master-Slave

The approach emits successful master response otherwise a slave result will be acquired.
A slave can shift early before a master complete a job. But if a master was lucky then a slave result will be declined anyway.
> You shouldn't care about locks in callback functions because they are thread-safe executed.

```go
func HandleAccuWeather(weather *Weather) failover.Handler {
    return func(ctx context.Context) (error, func()) {
        resp, err := accuWeather.Forecast(ctx, AccuWeatherRequest())
        return err, func() {
            accuWeatcherResponseToWeather(resp, weather)
        }
    }
}

func HandleOpenWeather(weather *Weather) failover.Handler {
    return func(ctx context.Context) (error, func()) {
        resp, err := openWeather.Forecast(ctx, OpenWeatherRequest())
        return err, func() {
            openWeatherResponseToWeather(resp, weather)
        }
    }
}

var weather Weather

err := failover.MasterSlave(
    context.Background(),
    time.Second, // slave shift time
    HandleAccuWeather(&weather), // master
    HandleOpenWeather(&weather), // slave
)
```

### Contributing
If you know another failover approaches or algorithms then feel free to send them in a pull request. Unit tests are required.
