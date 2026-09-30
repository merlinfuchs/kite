package wire

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gopkg.in/guregu/null.v4"
)

// AppIntegration is whether the app can use an integration. Credentials can't
// be read back.
type AppIntegration struct {
	IntegrationID string `json:"integration_id"`
	Enabled       bool   `json:"enabled"`
	// When the app last set the credential, for integrations that need one.
	CredentialUpdatedAt null.Time `json:"credential_updated_at"`
}

// AppIntegrationListResponse has an entry for every integration.
type AppIntegrationListResponse = []*AppIntegration

// AppIntegrationUpdateRequest turns an integration without a credential on or
// off.
type AppIntegrationUpdateRequest struct {
	// A pointer, so a missing value isn't taken as turning it off.
	Enabled *bool `json:"enabled"`
}

func (req AppIntegrationUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Enabled, validation.NotNil),
	)
}

type AppIntegrationUpdateResponse = AppIntegration

type AppIntegrationConnectRequest struct {
	Credential string `json:"credential"`
}

func (req AppIntegrationConnectRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Credential, validation.Required, validation.Length(1, MaxAppSecretValueLength)),
	)
}

type AppIntegrationConnectResponse = AppIntegration

type AppIntegrationDisconnectResponse = Empty
