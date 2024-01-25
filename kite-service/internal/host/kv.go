package host

import (
	"context"
	"fmt"

	"github.com/merlinfuchs/kite/go-types/fail"
	"github.com/merlinfuchs/kite/go-types/kvmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func (h HostEnvironment) callKVKeyGet(ctx context.Context, guildID string, data kvmodel.KVKeyGetCall) (kvmodel.KVKeyGetResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	res, err := h.kvStorage.GetKVStorageKey(ctx, guildID, data.Namespace, data.Key)
	if err != nil {
		if err == store.ErrNotFound {
			return kvmodel.KVKeyGetResponse{}, &fail.HostError{
				Code:    fail.HostErrorTypeKVKeyNotFound,
				Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
			}
		}
		return kvmodel.KVKeyGetResponse{}, err
	}

	return res.Value, nil
}

func (h HostEnvironment) callKVKeySet(ctx context.Context, guildID string, data kvmodel.KVKeySetCall) (kvmodel.KVKeySetResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	err := h.kvStorage.SetKVStorageKey(ctx, guildID, data.Namespace, data.Key, data.Value)
	if err != nil {
		return kvmodel.KVKeySetResponse{}, err
	}

	return kvmodel.KVKeySetResponse{}, nil
}

func (h HostEnvironment) callKVKeyDelete(ctx context.Context, guildID string, data kvmodel.KVKeyDeleteCall) (kvmodel.KVKeyDeleteResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	res, err := h.kvStorage.DeleteKVStorageKey(ctx, guildID, data.Namespace, data.Key)
	if err != nil {
		if err == store.ErrNotFound {
			return kvmodel.KVKeyDeleteResponse{}, &fail.HostError{
				Code:    fail.HostErrorTypeKVKeyNotFound,
				Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
			}
		}
		return kvmodel.KVKeyDeleteResponse{}, err
	}

	return res.Value, nil
}

func (h HostEnvironment) callKVKeyIncrease(ctx context.Context, guildID string, data kvmodel.KVKeyIncreaseCall) (kvmodel.KVKeyIncreaseResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	res, err := h.kvStorage.IncreaseKVStorageKey(ctx, guildID, data.Namespace, data.Key, data.Increment)
	if err != nil {
		if err == store.ErrNotFound {
			return kvmodel.KVKeyIncreaseResponse{}, &fail.HostError{
				Code:    fail.HostErrorTypeKVKeyNotFound,
				Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
			}
		}
		return kvmodel.KVKeyIncreaseResponse{}, err
	}

	return res.Value, nil
}
