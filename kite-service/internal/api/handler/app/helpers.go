package app

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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
