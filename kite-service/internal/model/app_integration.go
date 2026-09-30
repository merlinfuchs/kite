package model

import "time"

// AppIntegration is an integration the app set up, and whether it's enabled.
// IntegrationID is the integration defined in code, like cookie_api.
type AppIntegration struct {
	ID            string
	AppID         string
	IntegrationID string
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
