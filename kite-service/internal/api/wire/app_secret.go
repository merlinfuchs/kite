package wire

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

var appSecretNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// MaxAppSecretValueLength is the maximum size of a secret's value in bytes.
const MaxAppSecretValueLength = 4096

// AppSecret never includes the value, it can't be read back once saved.
type AppSecret struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AppSecretListResponse = []*AppSecret

type AppSecretCreateRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (req AppSecretCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, appSecretNameRules()...),
		validation.Field(&req.Value, validation.Required, validation.Length(1, MaxAppSecretValueLength)),
	)
}

type AppSecretCreateResponse = AppSecret

// AppSecretUpdateRequest renames a secret, and replaces its value if one is
// given.
type AppSecretUpdateRequest struct {
	Name  string      `json:"name"`
	Value null.String `json:"value"`
}

func (req AppSecretUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, appSecretNameRules()...),
		validation.Field(&req.Value, validation.When(req.Value.Valid,
			validation.Required,
			validation.Length(1, MaxAppSecretValueLength),
		)),
	)
}

type AppSecretUpdateResponse = AppSecret

type AppSecretDeleteResponse = Empty

func appSecretNameRules() []validation.Rule {
	return []validation.Rule{
		validation.Required,
		validation.Length(1, 100),
		validation.Match(appSecretNameRegex).
			Error("must only consist of letters, numbers, and underscores"),
	}
}

func AppSecretToWire(secret *model.AppSecret) *AppSecret {
	if secret == nil {
		return nil
	}

	return &AppSecret{
		ID:        secret.ID,
		Name:      secret.Name,
		CreatedAt: secret.CreatedAt,
		UpdatedAt: secret.UpdatedAt,
	}
}
