package host

import (
	"context"
	"fmt"

	"github.com/merlinfuchs/kite/go-types/fail"
	"github.com/merlinfuchs/kite/go-types/kvmodel"
)

func (h HostEnvironment) callKVKeyGet(ctx context.Context, data kvmodel.KVKeyGetCall) (kvmodel.KVKeyGetResponse, error) {
	val, ok := h.kv[data.Key]
	if !ok {
		return kvmodel.KVKeyGetResponse{}, &fail.HostError{
			Code:    fail.HostErrorTypeKVKeyNotFound,
			Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
		}
	}

	return val, nil
}

func (h HostEnvironment) callKVKeySet(ctx context.Context, data kvmodel.KVKeySetCall) (kvmodel.KVKeySetResponse, error) {
	h.kv[data.Key] = data.Value

	return kvmodel.KVKeySetResponse{}, nil
}

func (h HostEnvironment) callKVKeyDelete(ctx context.Context, data kvmodel.KVKeyDeleteCall) (kvmodel.KVKeyDeleteResponse, error) {
	val, ok := h.kv[data.Key]
	if !ok {
		return kvmodel.KVKeyDeleteResponse{}, &fail.HostError{
			Code:    fail.HostErrorTypeKVKeyNotFound,
			Message: fmt.Sprintf("key %s not found in namespace %s", data.Key, data.Namespace),
		}
	}

	delete(h.kv, data.Key)

	return val, nil
}

func (h HostEnvironment) callKVKeyIncrease(ctx context.Context, data kvmodel.KVKeyIncreaseCall) (kvmodel.KVKeyIncreaseResponse, error) {
	val, ok := h.kv[data.Key]
	if !ok {
		val = kvmodel.TypedKVValue{
			Type:  kvmodel.KVValueTypeInt,
			Value: kvmodel.KVInt(0),
		}
	}

	if val.Type != kvmodel.KVValueTypeInt {
		return kvmodel.KVKeyIncreaseResponse{}, &fail.HostError{
			Code:    fail.HostErrorTypeKVValueTypeMismatch,
			Message: fmt.Sprintf("key %s in namespace %s is not an int", data.Key, data.Namespace),
		}
	}

	val.Value = kvmodel.KVInt(val.Value.Int() + data.Increment)
	h.kv[data.Key] = val

	return val, nil
}
