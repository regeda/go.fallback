# go.fallback

[![Build Status](https://travis-ci.org/regeda/go.fallback.svg?branch=master)](https://travis-ci.org/regeda/go.fallback)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.fallback)](https://goreportcard.com/report/github.com/regeda/go.fallback)

Fallback algorithm aimed to make your requests stable and reliable.

### Benchmark
I have run a benchmark on MacBook Pro with CPU 2.7 GHz Intel Core i5 and RAM 8 GB 1867 MHz DDR3:
```
BenchmarkSuccessfulPrimary-4     2000000               801 ns/op              48 B/op          1 allocs/op
BenchmarkFailedPrimary-4         1000000              2180 ns/op             200 B/op          4 allocs/op
```

### Contributing
If you know another fallback approaches or algorithms then feel free to send them in a pull request. Unit tests are required.
