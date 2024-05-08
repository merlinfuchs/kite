package host

import (
	"context"
	"fmt"

	"github.com/merlinfuchs/kite/kite-sdk-go/fail"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func (h HostEnvironment) callKVKeyGet(ctx context.Context, data kv.KVKeyGetCall) (kv.KVKeyGetResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	res, err := h.kvStorage.GetKVStorageKey(ctx, h.AppID, data.Namespace, data.Key)
	if err != nil {
		if err == store.ErrNotFound {
			return kv.KVKeyGetResponse{}, &fail.HostError{
				Code:    fail.HostErrorTypeKVKeyNotFound,
				Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
			}
		}
		return kv.KVKeyGetResponse{}, err
	}

	return res.Value, nil
}

func (h HostEnvironment) callKVKeySet(ctx context.Context, data kv.KVKeySetCall) (kv.KVKeySetResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	err := h.kvStorage.SetKVStorageKey(ctx, h.AppID, data.Namespace, data.Key, data.Value)
	if err != nil {
		return kv.KVKeySetResponse{}, err
	}

	return kv.KVKeySetResponse{}, nil
}

func (h HostEnvironment) callKVKeyDelete(ctx context.Context, data kv.KVKeyDeleteCall) (kv.KVKeyDeleteResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	res, err := h.kvStorage.DeleteKVStorageKey(ctx, h.AppID, data.Namespace, data.Key)
	if err != nil {
		if err == store.ErrNotFound {
			return kv.KVKeyDeleteResponse{}, &fail.HostError{
				Code:    fail.HostErrorTypeKVKeyNotFound,
				Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
			}
		}
		return kv.KVKeyDeleteResponse{}, err
	}

	return res.Value, nil
}

func (h HostEnvironment) callKVKeyIncrease(ctx context.Context, data kv.KVKeyIncreaseCall) (kv.KVKeyIncreaseResponse, error) {
	if data.Namespace == "" {
		data.Namespace = "default"
	}

	res, err := h.kvStorage.IncreaseKVStorageKey(ctx, h.AppID, data.Namespace, data.Key, data.Increment)
	if err != nil {
		if err == store.ErrNotFound {
			return kv.KVKeyIncreaseResponse{}, &fail.HostError{
				Code:    fail.HostErrorTypeKVKeyNotFound,
				Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
			}
		}
		return kv.KVKeyIncreaseResponse{}, err
	}

	return res.Value, nil
}
