package model

import (
	"time"

	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
)

type KVStorageNamespace struct {
	AppID     string
	Namespace string
	KeyCount  int
}

type KVStorageValue struct {
	AppID     string
	Namespace string
	Key       string
	Value     kv.TypedKVValue
	CreatedAt time.Time
	UpdatedAt time.Time
}
