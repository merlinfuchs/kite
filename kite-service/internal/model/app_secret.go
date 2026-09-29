package model

import "time"

// AppSecret is a value flows can use without it being in the flow data. The
// value stays encrypted until a flow needs it.
type AppSecret struct {
	ID             string
	AppID          string
	Name           string
	ValueEncrypted string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
