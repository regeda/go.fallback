package fallback

import "context"

// Func executes something and returns an error if something went wrong
// or a function-resolver. A function-resolver will be called once if no errors
// were fired during an execution.
// All fallback algorithms guarantee that a function-resolver will be called
// in thread-safe mode and a shared data wouldn't be corrupted. Thus, an update
// of a shared data should be done inside a function-resolver:
// 		func resolve(something *Something) fallback.Func {
//			return func(ctx context.Context) (error, func()) {
//				resp, err := getSomething(ctx, Request{})
//				return err, func() {
//					responseToSomething(resp, something)
//				}
//			}
//		}
type Func func(context.Context) (error, func())
