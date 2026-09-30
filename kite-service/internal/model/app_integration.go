package model

import "time"

// AppIntegration is an app's choice to turn an integration without a
// credential on or off.
type AppIntegration struct {
	AppID         string
	IntegrationID string
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
