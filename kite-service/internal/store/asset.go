package store

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type AssetStore interface {
	CreateAsset(ctx context.Context, asset *model.Asset) (*model.Asset, error)
	Asset(ctx context.Context, id string) (*model.Asset, error)
	AssetWithContent(ctx context.Context, id string) (*model.Asset, error)
	DeleteAsset(ctx context.Context, id string) error
	// DeleteAppAssetObjects deletes the stored files of an app's assets that
	// no other app uses. The assets themselves are deleted with the app.
	DeleteAppAssetObjects(ctx context.Context, appID string) error
	DeleteExpiredAssets(ctx context.Context, timestamp time.Time) error
}
