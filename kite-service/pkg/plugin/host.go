package plugin

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/go-types/call"
	"github.com/merlinfuchs/kite/go-types/logmodel"
)

type HostEnvironment interface {
	Log(ctx context.Context, level logmodel.LogLevel, msg string)
	TrackEventHandled(ctx context.Context, eventType string, success bool, totalDuration time.Duration, executionDuration time.Duration)
	Call(ctx context.Context, req call.Call) (interface{}, error)
}
