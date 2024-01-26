package plugin

import (
	"context"

	"github.com/merlinfuchs/kite/go-types/call"
	"github.com/merlinfuchs/kite/go-types/logmodel"
)

type HostEnvironment interface {
	Log(ctx context.Context, level logmodel.LogLevel, msg string)
	Call(ctx context.Context, req call.Call) (interface{}, error)
}
