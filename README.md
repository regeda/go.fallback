# go.fallback

[![Build Status](https://travis-ci.org/regeda/go.fallback.svg?branch=master)](https://travis-ci.org/regeda/go.fallback)
[![Go Report Card](https://goreportcard.com/badge/github.com/regeda/go.fallback)](https://goreportcard.com/report/github.com/regeda/go.fallback)

Fallback algorithm aimed to make your requests stable and reliable.

### Primary
Primary approach resolves the first non-error result. A group is successful if any of goroutines was completed without an error.
```go
p := fallback.NewPrimary()
p.Go(func() error {
  return errors.New("broken")
})
p.Go(func() error {
  return nil
})
ok := p.Wait() // true
```

### Secondary
Secondary approach waits for primary's result and performs secondary goroutines if a primary was failed.
A group is successful if a primary exited without an error or any of secondary goroutines was completed without an error after primary failure.
```go
p := fallback.NewPrimary()
p.Go(func() error {
  fmt.Println("primary is broken")
  return errors.New("broken")
})

s := fallback.NewSecondary(p)
s.Go(func() error {
  return errors.New("broken")
})
s.Go(func() error {
  fmt.Println("secondary helps")
  return nil
})

ok := s.Wait() // true
// output:
//    primary is broken
//    secondary helps
```

Also, you can run a secondary without a primary wait. It helps getting a fallback result early even if a primary will failed unexpectedly.
```
p := fallback.NewPrimary()
p.Go(func() error {
  time.Sleep(time.Second)
  fmt.Println("primary is broken")
  return errors.New("broken")
})

s := fallback.NewSecondary(p)
s.Go(func() error {
  fmt.Println("secondary helps")
  return nil
})

s.Shift() // start a secondary immediately

ok := s.Wait() // true
// output:
//    secondary helps
//    primary is broken
```

### Context
You can create Primary or Secondary with a context. It allows cancelling a group if any of goroutines completed successfully.
```go
p, ctx := fallback.NewPrimaryWithContext(context.Background())
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

ok := p.Wait()
// output:
//    the first is good
//    the second is canceled
```

### Benchmark
I have run a benchmark on MacBook Pro with CPU 2.7 GHz Intel Core i5 and RAM 8 GB 1867 MHz DDR3:
```
BenchmarkPrimary-4                               2000000               790 ns/op
BenchmarkPrimaryWithCanceledSecondary-4          1000000              1989 ns/op
BenchmarkSecondaryWithFailedPrimary-4            1000000              2205 ns/op
```

### Contributing
If you know another fallback approaches or algorithms then feel free to send them in a pull request. Unit tests are required.
