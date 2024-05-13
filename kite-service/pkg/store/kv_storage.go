package store

import (
	"context"

	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type KVStorageStore interface {
	GetKVStorageNamespaces(ctx context.Context, appID string) ([]model.KVStorageNamespace, error)
	GetKVStorageKeys(ctx context.Context, appID, namespace string) ([]model.KVStorageValue, error)
	SetKVStorageKey(ctx context.Context, appID, namespace, key string, value kv.TypedKVValue) error
	GetKVStorageKey(ctx context.Context, appID, namespace, key string) (*model.KVStorageValue, error)
	DeleteKVStorageKey(ctx context.Context, appID, namespace, key string) (*model.KVStorageValue, error)
	IncreaseKVStorageKey(ctx context.Context, appID, namespace, key string, increment int) (*model.KVStorageValue, error)
}
