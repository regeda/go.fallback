package failover

import "context"

// Requester returns some output when some input was processed.
type Requester interface {
	Request(context.Context, interface{}) (interface{}, error)
}

// RequesterFunc is an adapter to allow the use of ordinary functions as failover requester.
type RequesterFunc func(context.Context, interface{}) (interface{}, error)

// Request calls f(ctx, in)
func (f RequesterFunc) Request(ctx context.Context, in interface{}) (interface{}, error) {
	return f(ctx, in)
}
