package app

import (
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
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

type DiscordAppInfo struct {
	ID          string
	Name        string
	Description null.String
}
