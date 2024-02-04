package internal

import "unsafe"

func Log(level int, msg string) {
	buf := []byte(msg)
	offset := uint32(uintptr(unsafe.Pointer(&buf[0])))

	kiteLog(uint32(level), offset, uint32(len(buf)))
}
