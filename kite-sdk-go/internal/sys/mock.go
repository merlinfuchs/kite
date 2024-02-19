//go:build !wasi && !wasm

package sys

import (
	"github.com/merlinfuchs/kite/kite-sdk-go/command"
	"github.com/merlinfuchs/kite/kite-sdk-go/event"
	"github.com/merlinfuchs/kite/kite-sdk-go/manifest"
)

var Manifest = manifest.Manifest{}
var Commands = []*command.Command{}

var EventHandlers = make(map[event.EventType][]event.EventHandler)

func KiteSetManifest(offset uint32, length uint32) uint32 {
	return 0
}

func KiteGetConfigSize() uint32 {
	return 0
}

func KiteGetConfig(offset uint32) uint32 {
	return 0
}

func KiteLog(level uint32, offset uint32, length uint32) uint32 {
	return 0
}

func KiteCall(offset uint32, length uint32) uint32 {
	return 0
}

func KiteGetEvent(offset uint32) uint32 {
	return 0
}

func KiteSetEventResponse(offset uint32, length uint32) uint32 {
	return 0
}

func KiteGetCallResponse(offset uint32) uint32 {
	return 0
}
