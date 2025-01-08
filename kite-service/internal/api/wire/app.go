package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type App struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    null.String       `json:"description"`
	Enabled        bool              `json:"enabled"`
	DisabledReason null.String       `json:"disabled_reason"`
	DiscordID      string            `json:"discord_id"`
	DiscordStatus  *AppDiscordStatus `json:"discord_status,omitempty"`
	OwnerUserID    string            `json:"owner_user_id"`
	CreatorUserID  string            `json:"creator_user_id"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type AppDiscordStatus struct {
	Status        string `json:"status,omitempty"`
	ActivityType  int    `json:"activity_type,omitempty"`
	ActivityName  string `json:"activity_name,omitempty"`
	ActivityState string `json:"activity_state,omitempty"`
	ActivityURL   string `json:"activity_url,omitempty"`
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
type AppStatusUpdateRequest struct {
	DiscordStatus *AppDiscordStatus `json:"discord_status,omitempty"`
}

func (req AppStatusUpdateRequest) Validate() error {
	return nil
}

type AppStatusUpdateResponse = App

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

	var status *AppDiscordStatus
	if app.DiscordStatus != nil {
		status = &AppDiscordStatus{
			Status:        app.DiscordStatus.Status,
			ActivityType:  app.DiscordStatus.ActivityType,
			ActivityName:  app.DiscordStatus.ActivityName,
			ActivityState: app.DiscordStatus.ActivityState,
			ActivityURL:   app.DiscordStatus.ActivityURL,
		}
	}

	return &App{
		ID:             app.ID,
		Name:           app.Name,
		Description:    app.Description,
		Enabled:        app.Enabled,
		DisabledReason: app.DisabledReason,
		DiscordID:      app.DiscordID,
		DiscordStatus:  status,
		OwnerUserID:    app.OwnerUserID,
		CreatorUserID:  app.CreatorUserID,
		CreatedAt:      app.CreatedAt,
		UpdatedAt:      app.UpdatedAt,
	}
}

type AppEntityListResponse = []*AppEntity

type AppEntity struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}
