package model

import (
	"time"

	"github.com/merlinfuchs/kite/go-types/kvmodel"
)

type KVStorageNamespace struct {
	GuildID   string
	Namespace string
	KeyCount  int
}

type KVStorageValue struct {
	GuildID   string
	Namespace string
	Key       string
	Value     kvmodel.TypedKVValue
	CreatedAt time.Time
	UpdatedAt time.Time
}
