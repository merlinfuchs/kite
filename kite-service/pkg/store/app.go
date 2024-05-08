package store

import (
	"context"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type AppStore interface {
	CreateApp(ctx context.Context, app *model.App) error
	UpdateApp(ctx context.Context, bot *model.App) (*model.App, error)
	UpdateAppStatus(ctx context.Context, bot *model.App) (*model.App, error)
	DeleteApp(ctx context.Context, appID distype.Snowflake) error
	GetApp(ctx context.Context, appID distype.Snowflake, ownerUserID distype.Snowflake) (*model.App, error)
	GetAppsWithValidToken(ctx context.Context) ([]model.App, error)
	GetAppsForOwnerUser(ctx context.Context, ownerUserID distype.Snowflake) ([]model.App, error)
	GetDistinctAppIDs(ctx context.Context) ([]distype.Snowflake, error)
	CheckUserIsOwnerOfApp(ctx context.Context, appID distype.Snowflake, ownerUserID distype.Snowflake) (bool, error)
}
