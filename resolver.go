package fallback

import "context"

type resolver interface {
	resolve(func())
}

func Resolve(ctx context.Context, f func()) {
	if r, ok := ctx.(resolver); ok {
		r.resolve(f)
	} else {
		f()
	}
}
