package model

import "time"

// AppSecret is a value flows can use without it being in the flow data. The
// value stays encrypted until a flow needs it. It either has a Name, for
// {{secrets.NAME}}, or an AppIntegrationID, when it's the credential of an
// integration the app set up. IntegrationID is that integration's ID in code.
type AppSecret struct {
	ID               string
	AppID            string
	Name             string
	AppIntegrationID string
	IntegrationID    string
	ValueEncrypted   string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
