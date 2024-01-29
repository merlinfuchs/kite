package internal

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/go-types/manifest"
)

//export kite_set_manifest
func kiteSetManifest(offset uint32, length uint32) uint32

//export kite_get_config_size
func kiteGetConfigSize() uint32

//export kite_get_config
func kiteGetConfig(offset uint32) uint32

//export kite_log
func kiteLog(level uint32, offset uint32, length uint32) uint32

//export kite_call
func kiteCall(offset uint32, length uint32) uint32

//export kite_get_event
func kiteGetEvent(offset uint32) uint32

//export kite_set_event_response
func kiteSetEventResponse(offset uint32, length uint32) uint32

//export kite_get_call_response
func kiteGetCallResponse(offset uint32) uint32

//export kite_get_api_version
func kiteGetAPIVersion() uint32 {
	return uint32(manifest.APIVersionAlpha)
}

//export kite_get_api_encoding
func kiteGetAPIEncoding() uint32 {
	return uint32(manifest.APIEncodingJSON)
}

var Manifest = manifest.Manifest{}

//export kite_describe
func kiteDescribe() {
	raw, err := json.Marshal(Manifest)
	if err != nil {
		panic(err)
	}

	offset := uint32(uintptr(unsafe.Pointer(&raw[0])))
	ok := kiteSetManifest(offset, uint32(len(raw)))
	if ok != 0 {
		panic(fmt.Errorf("failed to set manifest"))
	}
}

var EventHandlers = make(map[event.EventType][]event.EventHandler)

//export kite_handle
func kiteHandle(length uint32) {
	buf := make([]byte, length)

	offset := uint32(uintptr(unsafe.Pointer(&buf[0])))
	ok := kiteGetEvent(offset)
	if ok != 0 {
		return
	}

	var e event.Event
	err := json.Unmarshal(buf, &e)
	if err != nil {
		setEventResponse(event.EventError(err))
		return
	}

	handlers, exists := EventHandlers[e.Type]
	if !exists || len(handlers) == 0 {
		setEventResponse(event.EventError(fmt.Errorf("no handler for event type '%s'", e.Type)))
	}

	for _, handler := range handlers {
		err = handler(e)
		if err != nil {
			setEventResponse(event.EventError(err))
			return
		}
	}

	err = setEventResponse(event.EventResponse{
		Success: true,
	})
	if err != nil {
		setEventResponse(event.EventError(err))
		return
	}
}

func setEventResponse(v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}

	offset := uint32(uintptr(unsafe.Pointer(&raw[0])))
	ok := kiteSetEventResponse(offset, uint32(len(raw)))
	if ok != 0 {
		return fmt.Errorf("failed to set event response")
	}

	return nil
}
