package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type App struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Description   null.String `json:"description"`
	Enabled       bool        `json:"enabled"`
	DiscordID     string      `json:"discord_id"`
	OwnerUserID   string      `json:"owner_user_id"`
	CreatorUserID string      `json:"creator_user_id"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type AppGetResponse = App

type AppCreateRequest struct {
	DiscordToken string `json:"discord_token"`
}

func (req AppCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.DiscordToken, validation.Required),
	)
}

type AppCreateResponse = App

type AppUpdateRequest struct {
	Name        string      `json:"name"`
	Description null.String `json:"description"`
	Enabled     bool        `json:"enabled"`
}

func (req AppUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(0, 80)),
		validation.Field(&req.Description, validation.Length(0, 200)),
	)
}

type AppUpdateResponse = App

type AppTokenUpdateRequest struct {
	DiscordToken string `json:"discord_token"`
}

func (req AppTokenUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.DiscordToken, validation.Required),
	)
}

type AppTokenUpdateResponse = App

type AppDeleteResponse = Empty

type AppListResponse = []*App

func AppToWire(app *model.App) *App {
	if app == nil {
		return nil
	}

	return &App{
		ID:            app.ID,
		Name:          app.Name,
		Description:   app.Description,
		Enabled:       app.Enabled,
		DiscordID:     app.DiscordID,
		OwnerUserID:   app.OwnerUserID,
		CreatorUserID: app.CreatorUserID,
		CreatedAt:     app.CreatedAt,
		UpdatedAt:     app.UpdatedAt,
	}
}
