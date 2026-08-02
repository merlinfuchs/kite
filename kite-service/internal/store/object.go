package store

import (
	"context"
	"errors"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ObjectStore interface {
	CreateBucketIfNotExists(ctx context.Context, bucket string) error
	UploadObject(ctx context.Context, bucket string, object *model.Object) error
	UploadObjectIfNotExists(ctx context.Context, bucket string, object *model.Object) error
	DownloadObject(ctx context.Context, bucket string, name string) (*model.Object, error)
	DeleteObject(ctx context.Context, bucket string, name string) error
}

var ErrObjectStoreDisabled = errors.New("object store is not configured")

// DisabledObjectStore stands in for the object store when S3 isn't configured.
// Every operation fails with ErrObjectStoreDisabled instead of panicking, so
// that features which don't need object storage keep working.
type DisabledObjectStore struct{}

var _ ObjectStore = DisabledObjectStore{}

func (DisabledObjectStore) CreateBucketIfNotExists(ctx context.Context, bucket string) error {
	return ErrObjectStoreDisabled
}

func (DisabledObjectStore) UploadObject(ctx context.Context, bucket string, object *model.Object) error {
	return ErrObjectStoreDisabled
}

func (DisabledObjectStore) UploadObjectIfNotExists(ctx context.Context, bucket string, object *model.Object) error {
	return ErrObjectStoreDisabled
}

func (DisabledObjectStore) DownloadObject(ctx context.Context, bucket string, name string) (*model.Object, error) {
	return nil, ErrObjectStoreDisabled
}

func (DisabledObjectStore) DeleteObject(ctx context.Context, bucket string, name string) error {
	return ErrObjectStoreDisabled
}
