package wire

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
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
	// TokenInvalid is set when the app was disabled because Discord doesn't
	// accept its token.
	TokenInvalid bool `json:"token_invalid"`

	CommandsDeployPending bool                 `json:"commands_deploy_pending"`
	CommandsDeployError   *CommandsDeployError `json:"commands_deploy_error"`

	// Discord's approximate counts, null until first fetched.
	DiscordGuildCount       null.Int `json:"discord_guild_count"`
	DiscordUserInstallCount null.Int `json:"discord_user_install_count"`
}

type CommandsDeployError struct {
	Message string `json:"message"`
	Command string `json:"command,omitempty"`
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

func (e AppDiscordStatusEntry) Validate() error {
	return validation.ValidateStruct(&e,
		validation.Field(&e.ActivityName, validation.Length(0, 128)),
		validation.Field(&e.ActivityState, validation.Length(0, 128)),
		validation.Field(&e.ActivityURL, validation.When(
			e.ActivityType == int(discord.StreamingActivity),
			validation.Required.Error("Streaming needs a Twitch or YouTube URL"),
			validation.By(func(value interface{}) error {
				if !flow.IsStreamURL(value.(string)) {
					return errors.New("Discord only accepts Twitch or YouTube URLs for streaming")
				}
				return nil
			}),
		)),
	)
}

type AppGetResponse = App

// AppSetupGetResponse is how far a new app is through the setup checklist.
type AppSetupGetResponse struct {
	Invited    bool `json:"invited"`
	HasCommand bool `json:"has_command"`
	RanCommand bool `json:"ran_command"`
}

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

// AppInstallation is how the app can be installed in Discord. A nil context
// is turned off. It's read from Discord on demand instead of being stored.
type AppInstallation struct {
	GuildInstall *AppInstallParams `json:"guild_install"`
	UserInstall  *AppInstallParams `json:"user_install"`
}

// AppInstallParams are the scopes and bot permissions an installation context
// asks for by default.
type AppInstallParams struct {
	Scopes      []string `json:"scopes"`
	Permissions string   `json:"permissions"`
}

type AppInstallationGetResponse = AppInstallation

var appPermissionsPattern = regexp.MustCompile(`^\d{1,20}$`)

// AppInstallationUpdateRequest replaces how the app can be installed. At least
// one context has to stay on.
type AppInstallationUpdateRequest struct {
	GuildInstall *AppInstallParams `json:"guild_install"`
	UserInstall  *AppInstallParams `json:"user_install"`
}

func (req AppInstallationUpdateRequest) Validate() error {
	if req.GuildInstall == nil && req.UserInstall == nil {
		return validation.Errors{
			"guild_install": errors.New("turn on at least one way to install your app"),
		}
	}

	return validation.ValidateStruct(&req,
		validation.Field(&req.GuildInstall, validation.By(func(value interface{}) error {
			params, _ := value.(*AppInstallParams)
			if params == nil {
				return nil
			}
			if !slices.Contains(params.Scopes, "bot") {
				return errors.New("installing to a server always adds the bot")
			}
			// No permissions mean Administrator, the default, when read back.
			if params.Permissions == "0" {
				return errors.New("choose at least one permission")
			}
			return params.validate([]string{"applications.commands", "bot"})
		})),
		validation.Field(&req.UserInstall, validation.By(func(value interface{}) error {
			params, _ := value.(*AppInstallParams)
			if params == nil {
				return nil
			}
			// Apps installed to an account have no bot user in the server.
			return params.validate([]string{"applications.commands"})
		})),
	)
}

func (p *AppInstallParams) validate(allowedScopes []string) error {
	seen := map[string]bool{}
	for _, scope := range p.Scopes {
		if !slices.Contains(allowedScopes, scope) {
			return fmt.Errorf("unsupported scope %q", scope)
		}
		if seen[scope] {
			return fmt.Errorf("scope %q is listed twice", scope)
		}
		seen[scope] = true
	}
	if !seen["applications.commands"] {
		return errors.New("the applications.commands scope is needed for commands")
	}

	if !appPermissionsPattern.MatchString(p.Permissions) {
		return errors.New("permissions must be a permissions integer")
	}
	if !seen["bot"] && p.Permissions != "0" {
		return errors.New("permissions can only be asked for with the bot scope")
	}

	return nil
}

type AppInstallationUpdateResponse = AppInstallation

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

// AppDuplicateRequest copies the app to the Discord bot of the token, which
// can't be the bot of the app itself.
type AppDuplicateRequest struct {
	DiscordToken string `json:"discord_token"`
}

func (req AppDuplicateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.DiscordToken, validation.Required),
	)
}

type AppDuplicateResponse = App

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
		TokenInvalid:   model.HasInvalidToken(app),

		CommandsDeployPending: app.CommandsDeployAt.Valid,
		CommandsDeployError:   (*CommandsDeployError)(app.CommandsDeployError),

		DiscordGuildCount:       app.DiscordGuildCount,
		DiscordUserInstallCount: app.DiscordUserInstallCount,
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
