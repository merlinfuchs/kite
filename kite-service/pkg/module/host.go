package module

import (
	"context"
	"time"

	"github.com/merlinfuchs/kite/kite-types/call"
	"github.com/merlinfuchs/kite/kite-types/logmodel"
)

type HostEnvironment interface {
	Log(ctx context.Context, level logmodel.LogLevel, msg string)
	TrackEventHandled(ctx context.Context, eventType string, success bool, totalDuration time.Duration, executionDuration time.Duration)
	GetConfig(ctx context.Context) (map[string]interface{}, error)
	Call(ctx context.Context, req call.Call) (interface{}, error)
}

type moduleLogForwarder struct {
	env    HostEnvironment
	level  logmodel.LogLevel
	buffer []byte
}

func (p *moduleLogForwarder) Write(b []byte) (int, error) {
	// Buffer the log message until a newline is encountered
	p.buffer = append(p.buffer, b...)
	if b[len(b)-1] != '\n' && len(p.buffer) < 1000 {
		return len(b), nil
	}

	// Send the log message to the host
	go p.env.Log(context.Background(), p.level, string(p.buffer))
	p.buffer = p.buffer[:0]
	return len(b), nil
}
