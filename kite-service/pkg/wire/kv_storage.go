package wire

import (
	"time"

	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

type KVStorageNamespace struct {
	Namespace string `json:"namespace"`
	KeyCount  int    `json:"key_count"`
}

type KVStorageValue struct {
	Namespace string          `json:"namespace"`
	Key       string          `json:"key"`
	Value     kv.TypedKVValue `json:"value"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type KVStorageNamespaceListResponse APIResponse[[]KVStorageNamespace]

type KVStorageNamespaceKeyListResponse APIResponse[[]KVStorageValue]

func KVStorageNamespaceToWire(namespace *model.KVStorageNamespace) KVStorageNamespace {
	return KVStorageNamespace{
		Namespace: namespace.Namespace,
		KeyCount:  namespace.KeyCount,
	}
}

func KVStorageValueToWire(value *model.KVStorageValue) KVStorageValue {
	return KVStorageValue{
		Namespace: value.Namespace,
		Key:       value.Key,
		Value:     value.Value,
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}
}
