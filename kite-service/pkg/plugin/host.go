package plugin

import (
	"context"

	"github.com/merlinfuchs/kite/go-types/call"
	"github.com/merlinfuchs/kite/go-types/logmodel"
)

type HostEnvironment interface {
	Log(ctx context.Context, deploymentID string, level logmodel.LogLevel, msg string)
	Call(ctx context.Context, deploymentID string, guildID string, req call.Call) (interface{}, error)
}
