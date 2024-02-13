package module

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/merlinfuchs/kite/kite-sdk-go/call"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-sdk-go/fail"
	"github.com/merlinfuchs/kite/kite-sdk-go/log"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
)

func (p *Module) kiteSetManifest(offset uint32, length uint32) uint32 {
	if err := p.checkState(ModuleStateReady); err != nil {
		return 1
	}

	p.currentManifest = manifest.Manifest{}

	err := p.readJSON(offset, length, &p.currentManifest)
	if err != nil {
		return 1
	}

	return 0
}

func (p *Module) kiteGetConfigSize(ctx context.Context) uint32 {
	if err := p.checkState(ModuleStateEvent); err != nil {
		return p.resError(err)
	}

	p.startHostCall()
	defer p.endHostCall()

	config, err := p.env.GetConfig(ctx)
	if err != nil {
		slog.With("err", err).Error("failed to get config")
		return 0
	}

	body, err := json.Marshal(config)
	if err != nil {
		return 0
	}

	return uint32(len(body))
}

func (p *Module) kiteGetConfig(ctx context.Context, offset uint32) uint32 {
	if err := p.checkState(ModuleStateEvent); err != nil {
		return p.resError(err)
	}

	p.startHostCall()
	defer p.endHostCall()

	config, err := p.env.GetConfig(ctx)
	if err != nil {
		slog.With("err", err).Error("failed to get config")
		return 0
	}

	body, err := json.Marshal(config)
	if err != nil {
		return 1
	}

	err = p.writeBytes(offset, body)
	if err != nil {
		return 1
	}

	return 0
}

func (p *Module) kiteLog(ctx context.Context, level uint32, offset uint32, length uint32) uint32 {
	if err := p.checkState(ModuleStateEvent); err != nil {
		return 1
	}

	msg, err := p.readBytes(offset, length)
	if err != nil {
		return 1
	}

	lvl := log.LogLevel(level)
	if !lvl.Valid() {
		return 1
	}

	go p.env.Log(context.Background(), lvl, string(msg))
	return 0
}

func (p *Module) kiteCall(ctx context.Context, offset uint32, length uint32) uint32 {
	if err := p.checkState(ModuleStateEvent); err != nil {
		return p.resError(err)
	}

	p.startHostCall()
	defer p.endHostCall()

	var req call.Call
	if err := p.readJSON(offset, length, &req); err != nil {
		return p.resError(err)
	}

	res, err := p.env.Call(ctx, req)
	if err != nil {
		return p.resError(err)
	}

	return p.resSuccess(res)
}

func (p *Module) kiteGetEvent(offset uint32) uint32 {
	if err := p.checkState(ModuleStateEvent); err != nil {
		return 1
	}

	if p.currentEvent == nil {
		return 1
	}

	err := p.writeBytes(offset, p.currentEvent)
	if err != nil {
		return 1
	}

	p.currentEvent = nil
	return 0
}

func (p *Module) kiteSetEventResponse(offset uint32, length uint32) uint32 {
	if err := p.checkState(ModuleStateEvent); err != nil {
		return 1
	}

	p.currentEventResponse = &event.EventResponse{}

	err := p.readJSON(offset, length, p.currentEventResponse)
	if err != nil {
		return 1
	}

	return 0
}

func (p *Module) kiteGetCallResponse(offset uint32) uint32 {
	if err := p.checkState(ModuleStateEvent); err != nil {
		return 1
	}

	if p.currentCallResponse == nil {
		return 1
	}

	err := p.writeBytes(offset, p.currentCallResponse)
	if err != nil {
		return 1
	}

	p.currentCallResponse = nil
	return 0
}

func (p *Module) resError(err error) uint32 {
	var hostError *fail.HostError

	if herr, ok := err.(*fail.HostError); ok {
		hostError = herr
	} else {
		hostError = &fail.HostError{
			Code:    0,
			Message: err.Error(),
		}
	}

	body, err := json.Marshal(call.CallResponse{
		Success: false,
		Error:   hostError,
	})
	if err != nil {
		return 0
	}

	p.currentCallResponse = body
	return uint32(len(body))
}

func (p *Module) resSuccess(v interface{}) uint32 {
	body, err := json.Marshal(call.CallResponse{
		Success: true,
		Data:    v,
	})
	if err != nil {
		return 0
	}

	p.currentCallResponse = body
	return uint32(len(body))
}

func (p *Module) readJSON(offset uint32, length uint32, v interface{}) error {
	body, err := p.readBytes(offset, length)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

func (p *Module) readBytes(offset uint32, length uint32) ([]byte, error) {
	mem := p.m.Memory()
	if mem == nil {
		return nil, fmt.Errorf("memory not initialized")
	}

	body, ok := mem.Read(offset, length)
	if !ok {
		return nil, fmt.Errorf("failed to read memory")
	}

	return body, nil
}

func (p *Module) writeBytes(offset uint32, body []byte) error {
	mem := p.m.Memory()
	if mem == nil {
		return fmt.Errorf("memory not initialized")
	}

	ok := mem.Write(offset, body)
	if !ok {
		return fmt.Errorf("failed to write memory")
	}

	return nil
}

func (p *Module) checkState(state ModuleState) error {
	if p.state != state {
		return fmt.Errorf("this is only allowed in %s state", state)
	}
	return nil
}
