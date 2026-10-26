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
	AssetStorageByApp(ctx context.Context, appID string) (int64, error)
	DeleteExpiredAssets(ctx context.Context, timestamp time.Time) error
	// DeleteUnusedObjects deletes the stored files no asset uses anymore, like
	// the ones of deleted apps, if they are older than minAge.
	DeleteUnusedObjects(ctx context.Context, minAge time.Duration) error
}
