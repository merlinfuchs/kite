package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type APIKey struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	Key           string    `json:"key"`
	KeyHash       string    `json:"key_hash"`
	AppID         string    `json:"app_id"`
	CreatorUserID string    `json:"creator_user_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ExpiresAt     null.Time `json:"expires_at"`
}

type APIKeyCreateRequest struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func (req APIKeyCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Type, validation.Required, validation.In(string(model.APIKeyTypeIFTTT))),
		validation.Field(&req.Name, validation.Required),
	)
}

type APIKeyCreateResponse = APIKey

func APIKeyToWire(key *model.APIKey) *APIKey {
	return &APIKey{
		ID:            key.ID,
		Type:          string(key.Type),
		Name:          key.Name,
		Key:           key.Key,
		KeyHash:       key.KeyHash,
		AppID:         key.AppID,
		CreatorUserID: key.CreatorUserID,
		CreatedAt:     key.CreatedAt,
		UpdatedAt:     key.UpdatedAt,
		ExpiresAt:     key.ExpiresAt,
	}
}
