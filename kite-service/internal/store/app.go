package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

type AppUpdateOpts struct {
	ID           string
	Name         string
	Description  null.String
	DiscordToken string
	Enabled      bool
	UpdatedAt    time.Time
}

type AppStore interface {
	AppsByUser(ctx context.Context, userID string) ([]*model.App, error)
	CountAppsByUser(ctx context.Context, userID string) (int, error)
	App(ctx context.Context, id string) (*model.App, error)
	AppCredentials(ctx context.Context, id string) (*model.AppCredentials, error)
	CreateApp(ctx context.Context, app *model.App) (*model.App, error)
	UpdateApp(ctx context.Context, opts AppUpdateOpts) (*model.App, error)
	DeleteApp(ctx context.Context, id string) error
	EnabledAppIDs(ctx context.Context) ([]string, error)
	EnabledAppsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.App, error)
}
