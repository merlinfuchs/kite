package kvstorage

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type KVStorageHandler struct {
	kvStorage store.KVStorageStore
}

func NewHandler(kvStorage store.KVStorageStore) *KVStorageHandler {
	return &KVStorageHandler{
		kvStorage: kvStorage,
	}
}

func (h *KVStorageHandler) HandleKVStorageNamespaceList(c *fiber.Ctx) error {
	guildID := c.Params("guildID")

	namespaces, err := h.kvStorage.GetKVStorageNamespaces(c.Context(), guildID)
	if err != nil {
		return err
	}

	res := make([]wire.KVStorageNamespace, len(namespaces))
	for i, namespace := range namespaces {
		res[i] = wire.KVStorageNamespaceToWire(&namespace)
	}

	return c.JSON(wire.KVStorageNamespaceListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *KVStorageHandler) HandleKVStorageNamespaceKeyList(c *fiber.Ctx) error {
	guildID := c.Params("guildID")
	namespace := c.Params("namespace")

	values, err := h.kvStorage.GetKVStorageKeys(c.Context(), guildID, namespace)
	if err != nil {
		return err
	}

	res := make([]wire.KVStorageValue, len(values))
	for i, key := range values {
		res[i] = wire.KVStorageValueToWire(&key)
	}

	return c.JSON(wire.KVStorageNamespaceKeyListResponse{
		Success: true,
		Data:    res,
	})
}
