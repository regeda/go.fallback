# go.fallback

[![Build Status](https://travis-ci.org/regeda/go.fallback.svg?branch=master)](https://travis-ci.org/regeda/go.fallback)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.fallback)](https://goreportcard.com/report/github.com/regeda/go.fallback)
[![GoDoc](https://godoc.org/github.com/regeda/go.fallback?status.svg)](https://godoc.org/github.com/regeda/go.fallback)

Fallback provides an easy way to make remote calls to different providers under the same task.
Remote calls can often hang until they timed out. To avoid a failure, Fallback makes a set of standby providers for a reliable response.
It allows creating a hierarchy from a group of primary and secondary providers.
Thus if none of the primary providers solved a task, secondary providers might reach the goal.
In the meantime, Fallback controls thread-safe execution and failover synchronization.

### Primary
Primary approach resolves the first non-error result. A group is successful if any of goroutines was completed without an error.
```go
var result string

p := fallback.NewPrimary()
p.Go(func() (error, func()) {
  return errors.New("broken"), func() {
    result = "A"
  }
})
p.Go(func() (error, func()) {
  return nil, func() {
    result = "B"
  }
})

if p.Wait() {
  fmt.Printf("result = ", result)
}
// Output:
// result = B
```
> Go accepts Func type with `func() (error, func())` signature.
> Func perfoms a task and returns an error or "done" function.
> "Done" function will be executed in thread-safe mode. There is you can do assignments in shared memory without locks or semaphores.
> Basically, "done" function is performed once.

### Secondary
Secondary approach waits for primary's result and performs secondary goroutines if a primary was failed.
```go
var result string

p := fallback.NewPrimary()
p.Go(func() (error, func()) {
  fmt.Println("primary is broken")
  return errors.New("broken"), func() {
    result = "A"
  }
})

s := fallback.NewSecondary(p)
s.Go(func() (error, func()) {
  return errors.New("broken"), func() {
    result = "B"
  }
})
s.Go(func() (error, func()) {
  fmt.Println("secondary helps")
  return nil, func() {
    result = "C"
  }
})

if s.Wait() {
  fmt.Printf("result = ", result)
}
// Output:
// primary is broken
// secondary helps
// result = C
```

Also, you can run a secondary without a primary wait. It helps getting a fallback result early even if a primary will failed unexpectedly.
```go
var result string

p := fallback.NewPrimary()
p.Go(func() (error, func()) {
  time.Sleep(time.Second)
  fmt.Println("primary is broken")
  return errors.New("broken"), func() {
    result = "A"
  }
})

s := fallback.NewSecondary(p)
s.Go(func() (error, func()) {
  fmt.Println("secondary helps")
  return nil, func() {
    result = "B"
  }
})

s.Shift() // start a secondary immediately

if s.Wait() {
  fmt.Printf("result = ", result)
}
// Output:
// secondary helps
// primary is broken
// result = B
```

### Context
You can create Primary or Secondary with a context. It allows cancelling a group if any of goroutines completed successfully.
```go
var result string

p, ctx := fallback.NewPrimaryWithContext(context.Background())
p.Go(func() (error, func()) {
  fmt.Println("the first is good")
  return nil, func() {
    result = "A"
  }
})
p.Go(func() (error, func()) {
  select {
  case <-time.After(time.Second):
    return nil, func() {
      result = "B"
    }
  case <-ctx.Done():
    fmt.Println("the second is canceled")
    return ctx.Err(), fallback.NoopFunc
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
BenchmarkPrimary-4                               2000000               804 ns/op
BenchmarkPrimaryWithCanceledSecondary-4          1000000              1988 ns/op
BenchmarkSecondaryWithFailedPrimary-4            1000000              2332 ns/op
```

### Contributing
If you know another fallback approaches or algorithms then feel free to send them in a pull request. Unit tests are required.
