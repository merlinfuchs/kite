package internal

import (
	"unsafe"

	"github.com/merlinfuchs/kite/kite-sdk-go/internal/sys"
)

func Log(level int, msg string) {
	buf := []byte(msg)
	offset := uint32(uintptr(unsafe.Pointer(&buf[0])))

	sys.KiteLog(uint32(level), offset, uint32(len(buf)))
}
