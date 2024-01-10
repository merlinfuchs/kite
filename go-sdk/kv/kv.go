package kv

import (
	"encoding/json"

	"github.com/merlinfuchs/kite/go-sdk/internal"
	"github.com/merlinfuchs/kite/go-sdk/internal/util"
	"github.com/merlinfuchs/kite/go-types/call"
	"github.com/merlinfuchs/kite/go-types/kvmodel"
)

type KVNamespace struct {
	namespace string
}

func New(namespace ...string) *KVNamespace {
	if len(namespace) == 0 {
		return &KVNamespace{namespace: "default"}
	}

	return &KVNamespace{namespace: namespace[0]}
}

func (k *KVNamespace) Get(key string) (kvmodel.KVValue, error) {
	res, err := internal.MakeCall(call.Call{
		Type: call.KVKeyGet,
		Data: kvmodel.KVKeyGetCall{
			Namespace: k.namespace,
			Key:       key,
		},
	})
	if err != nil {
		return nil, err
	}

	val, err := util.DecodeT[kvmodel.KVKeyGetResponse](res.Data.(json.RawMessage))
	if err != nil {
		return nil, err
	}

	return val.Value, nil
}

func (k *KVNamespace) Set(key string, value kvmodel.KVValue) error {
	_, err := internal.MakeCall(call.Call{
		Type: call.KVKeySet,
		Data: kvmodel.KVKeySetCall{
			Namespace: k.namespace,
			Key:       key,
			Value: kvmodel.TypedKVValue{
				Type:  value.Type(),
				Value: value,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (k *KVNamespace) Delete(key string) (kvmodel.KVValue, error) {
	res, err := internal.MakeCall(call.Call{
		Type: call.KVKeyDelete,
		Data: kvmodel.KVKeyDeleteCall{
			Namespace: k.namespace,
			Key:       key,
		},
	})
	if err != nil {
		return nil, err
	}

	val, err := util.DecodeT[kvmodel.KVKeyDeleteResponse](res.Data.(json.RawMessage))
	if err != nil {
		return nil, err
	}

	return val.Value, nil
}

func (k *KVNamespace) Increase(key string, increment int) (kvmodel.KVValue, error) {
	res, err := internal.MakeCall(call.Call{
		Type: call.KVKeyIncrease,
		Data: kvmodel.KVKeyIncreaseCall{
			Namespace: k.namespace,
			Key:       key,
			Increment: increment,
		},
	})
	if err != nil {
		return nil, err
	}

	val, err := util.DecodeT[kvmodel.KVKeyIncreaseResponse](res.Data.(json.RawMessage))
	if err != nil {
		return nil, err
	}

	return val.Value, nil
}
