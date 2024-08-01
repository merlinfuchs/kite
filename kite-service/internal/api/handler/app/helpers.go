package app

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"gopkg.in/guregu/null.v4"
)

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

func (h *AppHandler) updateDiscordAppName(ctx context.Context, token string, name string) error {
	client := api.NewClient("Bot " + token).WithContext(ctx)

	_, err := client.ModifyCurrentUser(api.ModifyCurrentUserData{
		Username: option.NewString(name),
		// TODO: avatar
	})
	if err != nil {
		return fmt.Errorf("failed to update current user: %w", err)
	}

	return nil
}

type DiscordAppInfo struct {
	ID          string
	Name        string
	Description null.String
}
