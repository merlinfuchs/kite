package model

import (
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type PluginValue struct {
	ID               uint64      `json:"id"`
	PluginInstanceID string      `json:"plugin_instance_id"`
	Key              string      `json:"key"`
	Value            thing.Thing `json:"value"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type PluginValueOperation = provider.VariableOperation
