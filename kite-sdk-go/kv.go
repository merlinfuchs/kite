package kite

import (
	"encoding/json"

	"github.com/merlinfuchs/kite/kite-sdk-go/call"
	"github.com/merlinfuchs/kite/kite-sdk-go/internal"
	"github.com/merlinfuchs/kite/kite-sdk-go/internal/util"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
)

type KVNamespace struct {
	namespace string
}

func KV(namespace ...string) *KVNamespace {
	if len(namespace) == 0 {
		return &KVNamespace{namespace: "default"}
	}

	return &KVNamespace{namespace: namespace[0]}
}

func (k *KVNamespace) Get(key string) (kv.KVValue, error) {
	res, err := internal.MakeCall(call.Call{
		Type: call.KVKeyGet,
		Data: kv.KVKeyGetCall{
			Namespace: k.namespace,
			Key:       key,
		},
	})
	if err != nil {
		return nil, err
	}

	val, err := util.DecodeT[kv.KVKeyGetResponse](res.Data.(json.RawMessage))
	if err != nil {
		return nil, err
	}

	return val.Value, nil
}

func (k *KVNamespace) Set(key string, value kv.KVValue) error {
	_, err := internal.MakeCall(call.Call{
		Type: call.KVKeySet,
		Data: kv.KVKeySetCall{
			Namespace: k.namespace,
			Key:       key,
			Value: kv.TypedKVValue{
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

func (k *KVNamespace) Delete(key string) (kv.KVValue, error) {
	res, err := internal.MakeCall(call.Call{
		Type: call.KVKeyDelete,
		Data: kv.KVKeyDeleteCall{
			Namespace: k.namespace,
			Key:       key,
		},
	})
	if err != nil {
		return nil, err
	}

	val, err := util.DecodeT[kv.KVKeyDeleteResponse](res.Data.(json.RawMessage))
	if err != nil {
		return nil, err
	}

	return val.Value, nil
}

func (k *KVNamespace) Increase(key string, increment int) (kv.KVValue, error) {
	res, err := internal.MakeCall(call.Call{
		Type: call.KVKeyIncrease,
		Data: kv.KVKeyIncreaseCall{
			Namespace: k.namespace,
			Key:       key,
			Increment: increment,
		},
	})
	if err != nil {
		return nil, err
	}

	val, err := util.DecodeT[kv.KVKeyIncreaseResponse](res.Data.(json.RawMessage))
	if err != nil {
		return nil, err
	}

	return val.Value, nil
}
