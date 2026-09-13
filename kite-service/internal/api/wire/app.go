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
	Statuses      []AppDiscordStatusEntry `json:"statuses,omitempty"`
	ActiveID      string                  `json:"active_id,omitempty"`
	RotateEnabled bool                    `json:"rotate_enabled,omitempty"`
}

type AppDiscordStatusEntry struct {
	ID            string `json:"id"`
	Label         string `json:"label,omitempty"`
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
	if req.DiscordStatus == nil {
		return nil
	}

	return validation.ValidateStruct(req.DiscordStatus,
		validation.Field(&req.DiscordStatus.Statuses, validation.Length(0, 20)),
	)
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
		entries := make([]AppDiscordStatusEntry, len(app.DiscordStatus.Statuses))
		for i, e := range app.DiscordStatus.Statuses {
			entries[i] = AppDiscordStatusEntry{
				ID:            e.ID,
				Label:         e.Label,
				Status:        e.Status,
				ActivityType:  e.ActivityType,
				ActivityName:  e.ActivityName,
				ActivityState: e.ActivityState,
				ActivityURL:   e.ActivityURL,
			}
		}

		status = &AppDiscordStatus{
			Statuses:      entries,
			ActiveID:      app.DiscordStatus.ActiveID,
			RotateEnabled: app.DiscordStatus.RotateEnabled,
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

type AppEmojiListResponse = []*AppEmoji

type AppEmoji struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Animated  bool   `json:"animated"`
	Available bool   `json:"available"`
}

type AppEntityListResponse = []*AppEntity

type AppEntity struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}
