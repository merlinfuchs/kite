//go:build wasi || wasm

package sys

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/merlinfuchs/kite/kite-sdk-go/command"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
)

//export kite_set_manifest
func KiteSetManifest(offset uint32, length uint32) uint32

//export kite_get_config_size
func KiteGetConfigSize() uint32

//export kite_get_config
func KiteGetConfig(offset uint32) uint32

//export kite_log
func KiteLog(level uint32, offset uint32, length uint32) uint32

//export kite_call
func KiteCall(offset uint32, length uint32) uint32

//export kite_get_event
func KiteGetEvent(offset uint32) uint32

//export kite_set_event_response
func KiteSetEventResponse(offset uint32, length uint32) uint32

//export kite_get_call_response
func KiteGetCallResponse(offset uint32) uint32

//export kite_get_api_version
func kiteGetAPIVersion() uint32 {
	return uint32(manifest.APIVersionAlpha)
}

//export kite_get_api_encoding
func kiteGetAPIEncoding() uint32 {
	return uint32(manifest.APIEncodingJSON)
}

var Manifest = manifest.Manifest{}
var Commands = []*command.Command{}

//export kite_describe
func kiteDescribe() {
	manifest := Manifest

	for _, cmd := range Commands {
		manifest.DiscordCommands = append(manifest.DiscordCommands, cmd.DiscordCommand)
	}

	raw, err := json.Marshal(manifest)
	if err != nil {
		panic(err)
	}

	offset := uint32(uintptr(unsafe.Pointer(&raw[0])))
	ok := KiteSetManifest(offset, uint32(len(raw)))
	if ok != 0 {
		panic(fmt.Errorf("failed to set manifest"))
	}
}

var EventHandlers = make(map[event.EventType][]event.EventHandler)

//export kite_handle
func kiteHandle(length uint32) {
	buf := make([]byte, length)

	offset := uint32(uintptr(unsafe.Pointer(&buf[0])))
	ok := KiteGetEvent(offset)
	if ok != 0 {
		return
	}

	var e event.Event
	err := json.Unmarshal(buf, &e)
	if err != nil {
		SetEventResponse(event.EventError(err))
		return
	}

	handlers, exists := EventHandlers[e.Type]
	if !exists || len(handlers) == 0 {
		SetEventResponse(event.EventError(fmt.Errorf("no handler for event type '%s'", e.Type)))
	}

	for _, handler := range handlers {
		err = handler(e)
		if err != nil {
			SetEventResponse(event.EventError(err))
			return
		}
	}

	err = SetEventResponse(event.EventResponse{
		Success: true,
	})
	if err != nil {
		SetEventResponse(event.EventError(err))
		return
	}
}

func SetEventResponse(v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	offset := uint32(uintptr(unsafe.Pointer(&raw[0])))
	ok := KiteSetEventResponse(offset, uint32(len(raw)))
	if ok != 0 {
		return fmt.Errorf("failed to set event response")
	}

	return nil
}
