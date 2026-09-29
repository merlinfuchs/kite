package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// AppIntegration is an integration the app connected with a credential. The
// credential can't be read back.
type AppIntegration struct {
	IntegrationID string    `json:"integration_id"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AppIntegrationListResponse = []*AppIntegration

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

func AppIntegrationToWire(secret *model.AppSecret) *AppIntegration {
	if secret == nil {
		return nil
	}

	return &AppIntegration{
		IntegrationID: secret.IntegrationID,
		UpdatedAt:     secret.UpdatedAt,
	}
}
