# go.fallback

[![Build Status](https://travis-ci.org/regeda/go.fallback.svg?branch=master)](https://travis-ci.org/regeda/go.fallback)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.fallback)](https://goreportcard.com/report/github.com/regeda/go.fallback)
[![GoDoc](https://godoc.org/github.com/regeda/go.fallback?status.svg)](https://godoc.org/github.com/regeda/go.fallback)
[![codecov](https://codecov.io/gh/regeda/go.fallback/branch/master/graph/badge.svg)](https://codecov.io/gh/regeda/go.fallback)

Remote calls can often hang until they timed out. To avoid a failure, Fallback makes a set of standby providers that work with the same task.
It allows creating a hierarchy from a group of primary and secondary providers.
Thus if none of the primary providers solved a task, secondary providers take a chance return a successful response.
In the meantime, Fallback controls thread-safe execution and failover synchronization.

### Primary
Primary approach resolves the first non-error result. A group is successful if any of goroutines was completed without an error.
```go
var result string

p := fallback.NewPrimary()
p.Go(func() (func(), error) {
  return nil, errors.New("broken")
})
p.Go(func() (func(), error) {
  return func() {
    result = "ok"
  }, nil
})

if p.Wait() {
  fmt.Printf("result = ", result)
}
// Output:
// result = ok
```
> Go accepts Func type with `func() (func(), error)` signature.
> Func perfoms a task and returns an error or "done" function.
> "Done" function will be executed in thread-safe mode. There is you can do assignments in shared memory without locks or semaphores.
> Basically, "done" function is performed once.

### Secondary
Secondary approach waits for primary's result and performs secondary goroutines if a primary was failed.
```go
var result string

p := fallback.NewPrimary()
p.Go(func() (func(), error) {
  fmt.Println("primary is broken")
  return nil, errors.New("broken")
})

s := fallback.NewSecondary(p)
s.Go(func() (func(), error) {
  return nil, errors.New("broken")
})
s.Go(func() (func(), error) {
  fmt.Println("secondary helps")
  return func() {
    result = "ok"
  }, nil
})

if s.Wait() {
  fmt.Printf("result = ", result)
}
// Output:
// primary is broken
// secondary helps
// result = ok
```

Also, you can run a secondary without a primary wait. It helps getting a fallback result early even if a primary will failed unexpectedly.
```go
var result string

p := fallback.NewPrimary()
p.Go(func() (func(), error) {
  time.Sleep(time.Second)
  fmt.Println("primary is broken")
  return nil, errors.New("broken")
})

s := fallback.NewSecondary(p)
s.Go(func() (func(), error) {
  fmt.Println("secondary helps")
  return func() {
    result = "ok"
  }, nil
})

s.Shift() // start a secondary immediately

if s.Wait() {
  fmt.Printf("result = ", result)
}
// Output:
// secondary helps
// primary is broken
// result = ok
```

### Context
You can create Primary or Secondary with a context. It allows cancelling a group if any of goroutines completed successfully.
```go
var result string

p, ctx := fallback.NewPrimaryWithContext(context.Background())
p.Go(func() (func(), error) {
  fmt.Println("the first is good")
  return func() {
    result = "A"
  }, nil
})
p.Go(func() (func(), error) {
  select {
  case <-time.After(time.Second):
    return func() {
      result = "B"
    }, nil
  case <-ctx.Done():
    fmt.Println("the second is canceled")
    return nil, ctx.Err()
  }
})

if p.Wait() {
  fmt.Printf("result = ", result)
}
// Output:
// the first is good
// the second is canceled
// result = A
```

### Benchmark
I have run a benchmark on MacBook Pro with CPU 2.7 GHz Intel Core i5 and RAM 8 GB 1867 MHz DDR3:
```
BenchmarkPrimary-4                         2000000         899 ns/op         48 B/op        1 allocs/op
BenchmarkPrimaryWithCanceledSecondary-4    1000000        2037 ns/op        176 B/op        4 allocs/op
BenchmarkSecondaryWithFailedPrimary-4      1000000        2529 ns/op        192 B/op        5 allocs/op
```

### Contributing
If you know another fallback approaches or algorithms then feel free to send them in a pull request. Unit tests are required.
