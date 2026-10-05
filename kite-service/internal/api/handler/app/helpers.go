package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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
}
