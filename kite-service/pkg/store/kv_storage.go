package store

import (
	"context"

	"github.com/merlinfuchs/kite/go-types/kvmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type KVStorageStore interface {
	GetKVStorageNamespaces(ctx context.Context, guildID string) ([]model.KVStorageNamespace, error)
	GetKVStorageKeys(ctx context.Context, guildID, namespace string) ([]model.KVStorageValue, error)
	SetKVStorageKey(ctx context.Context, guildID, namespace, key string, value kvmodel.TypedKVValue) error
	GetKVStorageKey(ctx context.Context, guildID, namespace, key string) (*model.KVStorageValue, error)
	DeleteKVStorageKey(ctx context.Context, guildID, namespace, key string) (*model.KVStorageValue, error)
	IncreaseKVStorageKey(ctx context.Context, guildID, namespace, key string, increment int) (*model.KVStorageValue, error)
}
