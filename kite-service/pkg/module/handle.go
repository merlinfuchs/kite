package module

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/merlinfuchs/kite/go-types/event"
)

type HandleResult struct {
	TotalDuration     time.Duration
	ExecutionDuration time.Duration
}

func (p *Module) Handle(ctx context.Context, e *event.Event) (HandleResult, error) {
	if p.config.TotalTimeLimit != 0 {
		ctx, p.cancel = context.WithTimeout(ctx, p.config.TotalTimeLimit)
	} else {
		ctx, p.cancel = context.WithCancel(ctx)
	}

	raw, err := json.Marshal(e)
	if err != nil {
		return HandleResult{}, fmt.Errorf("failed to marshal event: %w", err)
	}
	p.currentEvent = raw
	p.currentGuildID = e.GuildID

	fn := p.m.ExportedFunction("kite_handle")
	if fn == nil {
		return HandleResult{}, fmt.Errorf("kite_handle not defined")
	}

	p.startHandle()

	_, err = fn.Call(ctx, uint64(len(raw)))
	if err != nil {
		p.endHandle()
		return HandleResult{}, fmt.Errorf("failed to call kite_handle: %w", err)
	}

	p.endHandle()

	res := HandleResult{
		TotalDuration:     p.totalDuration(),
		ExecutionDuration: p.executionDuration(),
	}

	if p.currentEventResponse == nil {
		return res, nil
	}

	if !p.currentEventResponse.Success {
		err := p.currentEventResponse.Error
		p.currentEventResponse = nil
		return res, err
	}

	p.currentEventResponse = nil
	return res, nil
}

func (p *Module) Close(ctx context.Context) error {
	if err := p.m.Close(ctx); err != nil {
		return err
	}
	return p.r.Close(ctx)
}

func (p *Module) totalDuration() time.Duration {
	return time.Since(p.handleStartAt)
}

func (p *Module) executionDuration() time.Duration {
	totalTime := p.totalDuration()

	hostCallDuration := p.hostCallDuration
	if !p.hostCallStartAt.IsZero() {
		hostCallDuration += time.Since(p.hostCallStartAt)
	}

	return totalTime - hostCallDuration
}

func (p *Module) startHandle() {
	p.state = ModuleStateEvent
	p.ticker = time.NewTicker(time.Millisecond * 5)
	p.handleStartAt = time.Now()
	p.hostCallDuration = 0

	if p.config.ExecutionTimeLimit != 0 {
		go func() {
			for {
				_, ok := <-p.ticker.C
				if !ok {
					return
				}

				executionDuration := p.executionDuration()
				if executionDuration > p.config.ExecutionTimeLimit {
					p.cancel()
				}
			}
		}()
	}
}

func (p *Module) endHandle() {
	p.state = ModuleStateReady
	p.ticker.Stop()
}
