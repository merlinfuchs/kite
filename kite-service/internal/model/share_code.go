package model

import "time"

type ShareCode struct {
	Code      string
	Data      string
	CreatedAt time.Time
}
