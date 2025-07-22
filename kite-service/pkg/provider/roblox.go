package provider

import (
	"context"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type RobloxProvider interface {
	UserByID(ctx context.Context, id int64) (*thing.RobloxUserValue, error)
	UsersByUsername(ctx context.Context, username string) ([]thing.RobloxUserValue, error)
}
