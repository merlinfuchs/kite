package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func (h *AppHandler) getAppClient(ctx context.Context, app *model.App) (*api.Client, error) {
	token, err := h.tokenCrypt.DecryptString(app.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	return api.NewClient("Bot " + token).WithContext(ctx), nil
}

func (h *AppHandler) getDiscordAppInfo(ctx context.Context, token string) (*DiscordAppInfo, error) {
	client := api.NewClient("Bot " + token).WithContext(ctx)

	app, err := client.CurrentApplication()
	if err != nil {
		return nil, fmt.Errorf("failed to get current application: %w", err)
	}

	return &DiscordAppInfo{
		ID:          app.ID.String(),
		Name:        app.Name,
		Description: null.NewString(app.Description, app.Description != ""),
		VerifyKey:   app.VerifyKey,
	}, nil
}

func (h *AppHandler) updateDiscordApp(ctx context.Context, app *model.App) error {
	client, err := h.getAppClient(ctx, app)
	if err != nil {
		return fmt.Errorf("failed to get app client: %w", err)
	}

	req := struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{
		Name:        app.Name,
		Description: app.Description.String,
	}

	_, err = client.Request("PATCH", api.EndpointApplications+app.DiscordID, httputil.WithJSONBody(req))
	if err != nil {
		return err
	}

	return nil
}

func (h *AppHandler) updateDiscordBotUser(ctx context.Context, app *model.App) error {
	client, err := h.getAppClient(ctx, app)
	if err != nil {
		return fmt.Errorf("failed to get app client: %w", err)
	}

	_, err = client.ModifyCurrentUser(api.ModifyCurrentUserData{
		Username: option.NewString(app.Name),
	})
	if err != nil {
		return err
	}

	return nil
}

// updateDiscordBotProfile sets the avatar and banner of the bot user. A nil
// image is left unchanged and an empty one is removed.
func (h *AppHandler) updateDiscordBotProfile(ctx context.Context, app *model.App, avatar *string, banner *string) (*discord.User, error) {
	client, err := h.getAppClient(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to get app client: %w", err)
	}

	// Discord removes an image when its field is null, so the fields can't be
	// a struct with omitempty.
	req := map[string]*string{}
	if avatar != nil {
		req["avatar"] = emptyToNil(avatar)
	}
	if banner != nil {
		req["banner"] = emptyToNil(banner)
	}

	var user *discord.User
	err = client.RequestJSON(&user, "PATCH", api.EndpointMe, httputil.WithJSONBody(req))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func emptyToNil(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

func appProfileToWire(user *discord.User) *wire.AppProfile {
	// AvatarURL falls back to the default avatar, which can't be removed.
	var avatarURL string
	if user.Avatar != "" {
		avatarURL = user.AvatarURL()
	}
	bannerURL := user.BannerURL()

	return &wire.AppProfile{
		AvatarURL: null.NewString(avatarURL, avatarURL != ""),
		BannerURL: null.NewString(bannerURL, bannerURL != ""),
	}
}

// discordProfileErrorMessage returns the reason Discord gave for rejecting a
// profile change, like changing the avatar too fast.
func discordProfileErrorMessage(err error) string {
	var restErr *httputil.HTTPError
	if !errors.As(err, &restErr) {
		return err.Error()
	}

	var fields map[string]struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"_errors"`
	}
	if json.Unmarshal(restErr.Errors, &fields) == nil {
		for _, name := range []string{"avatar", "banner"} {
			if errs := fields[name].Errors; len(errs) > 0 && errs[0].Message != "" {
				return errs[0].Message
			}
		}
	}

	if restErr.Message != "" {
		return restErr.Message
	}
	return "Discord rejected the image"
}

// discordInstallParams is an install params object of Discord.
type discordInstallParams struct {
	Scopes      []string `json:"scopes"`
	Permissions string   `json:"permissions"`
}

// discordIntegrationTypesConfig is keyed by the integration type, 0 for
// installing to a server and 1 for installing to an account. A type that is
// turned off is missing.
type discordIntegrationTypesConfig map[string]struct {
	OAuth2InstallParams *discordInstallParams `json:"oauth2_install_params,omitempty"`
}

const (
	discordGuildInstall = "0"
	discordUserInstall  = "1"

	// defaultGuildPermissions is Administrator.
	defaultGuildPermissions = "8"
)

func (h *AppHandler) getDiscordInstallation(ctx context.Context, app *model.App) (*wire.AppInstallation, error) {
	client, err := h.getAppClient(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to get app client: %w", err)
	}

	var res struct {
		IntegrationTypesConfig discordIntegrationTypesConfig `json:"integration_types_config"`
	}
	err = client.RequestJSON(&res, "GET", api.EndpointApplications+"@me")
	if err != nil {
		return nil, err
	}

	return installationToWire(res.IntegrationTypesConfig), nil
}

// updateDiscordInstallation replaces the installation contexts of the app.
// Discord turns off every context that isn't in the request.
func (h *AppHandler) updateDiscordInstallation(ctx context.Context, app *model.App, req wire.AppInstallationUpdateRequest) (*wire.AppInstallation, error) {
	client, err := h.getAppClient(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to get app client: %w", err)
	}

	config := discordIntegrationTypesConfig{}
	for key, params := range map[string]*wire.AppInstallParams{
		discordGuildInstall: req.GuildInstall,
		discordUserInstall:  req.UserInstall,
	} {
		if params == nil {
			continue
		}
		entry := config[key]
		entry.OAuth2InstallParams = &discordInstallParams{
			Scopes:      params.Scopes,
			Permissions: params.Permissions,
		}
		config[key] = entry
	}

	body := struct {
		IntegrationTypesConfig discordIntegrationTypesConfig `json:"integration_types_config"`
	}{config}

	var res struct {
		IntegrationTypesConfig discordIntegrationTypesConfig `json:"integration_types_config"`
	}
	err = client.RequestJSON(&res, "PATCH", api.EndpointApplications+"@me", httputil.WithJSONBody(body))
	if err != nil {
		return nil, err
	}

	return installationToWire(res.IntegrationTypesConfig), nil
}

func installationToWire(config discordIntegrationTypesConfig) *wire.AppInstallation {
	res := &wire.AppInstallation{}
	for key, entry := range config {
		// A context can be on without default install params. Servers are then
		// asked for Administrator, like Kite's invite links always did.
		permissions := "0"
		if key == discordGuildInstall {
			permissions = defaultGuildPermissions
		}
		params := &wire.AppInstallParams{Scopes: []string{}, Permissions: permissions}
		if entry.OAuth2InstallParams != nil {
			if entry.OAuth2InstallParams.Scopes != nil {
				params.Scopes = entry.OAuth2InstallParams.Scopes
			}
			if entry.OAuth2InstallParams.Permissions != "" {
				params.Permissions = entry.OAuth2InstallParams.Permissions
			}
		}
		// New Discord apps ask servers for no permissions, which Kite never
		// did, so no permissions count as not chosen yet.
		if key == discordGuildInstall && params.Permissions == "0" {
			params.Permissions = defaultGuildPermissions
		}

		switch key {
		case discordGuildInstall:
			res.GuildInstall = params
		case discordUserInstall:
			res.UserInstall = params
		}
	}
	return res
}

// discordErrorMessage returns the reason Discord gave for rejecting a request.
func discordErrorMessage(err error) string {
	var restErr *httputil.HTTPError
	if errors.As(err, &restErr) && restErr.Message != "" {
		return restErr.Message
	}
	return err.Error()
}

func (h *AppHandler) getAppEmojis(ctx context.Context, app *model.App) ([]discord.Emoji, error) {
	client, err := h.getAppClient(ctx, app)
	if err != nil {
		return nil, fmt.Errorf("failed to get app client: %w", err)
	}

	var res struct {
		Items []discord.Emoji `json:"items"`
	}

	err = client.RequestJSON(&res, "GET", api.EndpointApplications+app.DiscordID+"/emojis")
	if err != nil {
		return nil, err
	}

	return res.Items, nil
}

type DiscordAppInfo struct {
	ID          string
	Name        string
	Description null.String
	VerifyKey   string
}

// retryFailedCommandsDeploy deploys the commands again if the last deploy
// failed, since a new token or restarting the app may have fixed it.
func (h *AppHandler) retryFailedCommandsDeploy(c *handler.Context) {
	if c.App.CommandsDeployError == nil {
		return
	}

	if err := h.appStore.QueueAppCommandsDeploy(c.Context(), c.App.ID, false); err != nil {
		slog.Error(
			"Failed to queue commands deploy",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
	}
}
