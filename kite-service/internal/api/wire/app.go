package wire

import (
	"errors"
	"regexp"
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

// AppProfile is how the app's bot user looks in Discord. It's read from
// Discord on demand instead of being stored.
type AppProfile struct {
	AvatarURL null.String `json:"avatar_url"`
	BannerURL null.String `json:"banner_url"`
}

type AppProfileGetResponse = AppProfile

// appProfileImageMaxLength bounds an image data URI. It leaves room for one
// image of about 5 MB below handler.MaxJSONBodySize.
const appProfileImageMaxLength = 7 * 1024 * 1024

var appProfileImagePattern = regexp.MustCompile(`^data:image/(png|jpeg|gif|webp);base64,[A-Za-z0-9+/]+=*$`)

// AppProfileUpdateRequest changes the avatar and banner of the app's bot user.
// A missing field is left unchanged, an empty string removes the image and
// anything else must be an image data URI.
type AppProfileUpdateRequest struct {
	Avatar *string `json:"avatar,omitempty"`
	Banner *string `json:"banner,omitempty"`
}

func (req AppProfileUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Avatar, validation.By(validateAppProfileImage)),
		validation.Field(&req.Banner, validation.By(validateAppProfileImage)),
	)
}

func validateAppProfileImage(value interface{}) error {
	image, _ := value.(*string)
	if image == nil || *image == "" {
		return nil
	}

	if len(*image) > appProfileImageMaxLength {
		return errors.New("the image is too large, the limit is 5 MB")
	}
	if !appProfileImagePattern.MatchString(*image) {
		return errors.New("must be a PNG, JPEG, GIF or WebP image")
	}

	return nil
}

type AppProfileUpdateResponse = AppProfile

type AppStatusUpdateRequest struct {
	DiscordStatus *AppDiscordStatus `json:"discord_status,omitempty"`
}

func (req AppStatusUpdateRequest) Validate() error {
	if req.DiscordStatus == nil {
		return nil
	}

	return validation.ValidateStruct(req.DiscordStatus,
		validation.Field(&req.DiscordStatus.Statuses, validation.Length(0, 10)),
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
			entries[i] = AppDiscordStatusEntry(e)
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
