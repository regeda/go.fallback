package failover

import "context"

// Handler executes something and postpones the result in future function.
type Handler func(context.Context) (error, func())
