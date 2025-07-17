package eval

import (
	"context"
	"time"
)

// proxyContext is a wrapper around a context.Context that prevents access to additional properties
// This prevents users from accessing additional properties or methods that might be present
// on custom context implementations, which could lead to information leakage.
type proxyContext struct {
	ctx context.Context
}

func (sc proxyContext) Deadline() (time.Time, bool) {
	return sc.ctx.Deadline()
}

func (sc proxyContext) Done() <-chan struct{} {
	return sc.ctx.Done()
}

func (sc proxyContext) Err() error {
	return sc.ctx.Err()
}

func (sc proxyContext) Value(key interface{}) interface{} {
	return sc.ctx.Value(key)
}
