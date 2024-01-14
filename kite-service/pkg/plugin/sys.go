package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/merlinfuchs/kite/go-types/call"
	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/go-types/fail"
	"github.com/merlinfuchs/kite/go-types/logmodel"
)

func (p *Plugin) kiteGetConfigSize() uint32 {
	body, err := json.Marshal(p.config.UserConfig)
	if err != nil {
		return 0
	}

	return uint32(len(body))
}

func (p *Plugin) kiteGetConfig(offset uint32) uint32 {
	body, err := json.Marshal(p.config.UserConfig)
	if err != nil {
		return 1
	}

	err = p.writeBytes(offset, body)
	if err != nil {
		return 1
	}

	return 0
}

func (p *Plugin) kiteLog(level uint32, offset uint32, length uint32) uint32 {
	msg, err := p.readBytes(offset, length)
	if err != nil {
		return 1
	}

	lvl := logmodel.LogLevel(level)
	if !lvl.Valid() {
		return 1
	}

	p.env.Log(p.ctx, lvl, string(msg))
	return 0
}

func (p *Plugin) kiteCall(offset uint32, length uint32) uint32 {
	if p.state != PluginStateEvent {
		return p.resError(fmt.Errorf("host calls are only allowed in event handlers"))
	}

	p.startHostCall()
	defer p.endHostCall()

	var req call.Call
	if err := p.readJSON(offset, length, &req); err != nil {
		return p.resError(err)
	}

	res, err := p.env.Call(p.ctx, p.currentGuildID, req)
	if err != nil {
		return p.resError(err)
	}

	return p.resSuccess(res)
}

func (p *Plugin) kiteGetEvent(offset uint32) uint32 {
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

func (p *Plugin) kiteSetEventResponse(offset uint32, length uint32) uint32 {
	p.currentEventResponse = &event.EventResponse{}

	err := p.readJSON(offset, length, p.currentEventResponse)
	if err != nil {
		return 1
	}

	return 0
}

func (p *Plugin) kiteGetCallResponse(offset uint32) uint32 {
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

func (p *Plugin) resError(err error) uint32 {
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

func (p *Plugin) resSuccess(v interface{}) uint32 {
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

func (p *Plugin) readJSON(offset uint32, length uint32, v interface{}) error {
	body, err := p.readBytes(offset, length)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, v)
}

func (p *Plugin) readBytes(offset uint32, length uint32) ([]byte, error) {
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

func (p *Plugin) writeBytes(offset uint32, body []byte) error {
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
