package store

import (
	"context"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ObjectStore interface {
	CreateBucketIfNotExists(ctx context.Context, bucket string) error
	UploadObject(ctx context.Context, bucket string, object *model.Object) error
	UploadObjectIfNotExists(ctx context.Context, bucket string, object *model.Object) error
	DownloadObject(ctx context.Context, bucket string, name string) (*model.Object, error)
	DeleteObject(ctx context.Context, bucket string, name string) error
}
