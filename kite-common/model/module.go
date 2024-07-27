package model

import (
	"time"
)

type Module struct {
	ID        string
	WasmBytes []byte
	WasmHash  string
	CreatedAt time.Time
}
