# go.failover

[![Build Status](https://travis-ci.org/regeda/go.failover.svg?branch=master)](https://travis-ci.org/regeda/go.failover)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.failover)](https://goreportcard.com/report/github.com/regeda/go.failover)

Failover approaches and algorithms aimed to make your request stable and reliable.

## Master-Slave

The approach returns successful master response otherwise a slave result will be acquired.
A slave can shift early before a master complete a job. But if a master was lucky then a slave result will be omitted anyway.

```go
type AccuWeather struct{}
func (w *AccuWeather) Request(ctx context.Context, in interface{}) (interface{}, error) {
    return w.Forecast(ctx, parseAccuWeatherRequest(in))
}

type OpenWeather struct{}
func (w *OpenWeather) Request(ctx context.Context, in interface{}) (interface{}, error) {
    return w.Forecast(ctx, parseOpenWeatherRequest(in))
}

service := failover.MasterSlave(&AccuWeather{}, &OpenWeather{}, time.Second)
weather, err := service.Request(context.Background(), NewWeatherRequest())
...
```

If you need more that two providers you can create a tree of master-slave requests:
```go
failover.MasterSlave(
    failover.MasterSlave(&AccuWeather{}, &OpenWeather{}, time.Second),
    &YahooWeather{},
    2*time.Second,
)
```

### Contributing
If you know another failover approaches or algorithms then feel free to send them in a pull request. Unit tests are required.
