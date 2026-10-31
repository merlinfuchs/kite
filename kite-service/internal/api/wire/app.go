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
	// NoCredits is set when the app is paused because it used all of its
	// credits for the month. It stays enabled and connected, but runs no
	// flows.
	NoCredits bool `json:"no_credits"`

	CommandsDeployPending bool                 `json:"commands_deploy_pending"`
	CommandsDeployError   *CommandsDeployError `json:"commands_deploy_error"`

	// Discord's approximate counts, null until first fetched.
	DiscordGuildCount       null.Int `json:"discord_guild_count"`
	DiscordUserInstallCount null.Int `json:"discord_user_install_count"`

	// CreditNoticesEnabled is whether the owner gets a DM from the bot when
	// the app used 80% or all of its credits for the month.
	CreditNoticesEnabled bool `json:"credit_notices_enabled"`
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
	Invited bool `json:"invited"`
	// HasFlow is whether the app has something that can run, like a command.
	HasFlow bool `json:"has_flow"`
	RanFlow bool `json:"ran_flow"`
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

type AppCreditNoticesUpdateRequest struct {
	Enabled bool `json:"enabled"`
}

func (req AppCreditNoticesUpdateRequest) Validate() error {
	return nil
}

type AppCreditNoticesUpdateResponse = App

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
		NoCredits:      model.HasNoCredits(app),

		CommandsDeployPending: app.CommandsDeployAt.Valid,
		CommandsDeployError:   (*CommandsDeployError)(app.CommandsDeployError),

		DiscordGuildCount:       app.DiscordGuildCount,
		DiscordUserInstallCount: app.DiscordUserInstallCount,

		CreditNoticesEnabled: app.CreditNoticesEnabled,
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
