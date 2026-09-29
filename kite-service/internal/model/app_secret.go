package model

import "time"

// AppSecret is a value flows can use without it being in the flow data. The
// value stays encrypted until a flow needs it. It either has a Name, for
// {{secrets.NAME}}, or an IntegrationID, when it's an integration's credential.
type AppSecret struct {
	ID             string
	AppID          string
	Name           string
	IntegrationID  string
	ValueEncrypted string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
