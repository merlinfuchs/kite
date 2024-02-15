package kv

type KVKeyGetCall struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type KVKeyGetResponse = TypedKVValue

type KVKeySetCall struct {
	Namespace string       `json:"namespace"`
	Key       string       `json:"key"`
	Value     TypedKVValue `json:"value"`
}

type KVKeySetResponse struct{}

type KVKeyDeleteCall struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
}

type KVKeyDeleteResponse = TypedKVValue

type KVKeyIncreaseCall struct {
	Namespace string `json:"namespace"`
	Key       string `json:"key"`
	Increment int    `json:"increment"`
}

type KVKeyIncreaseResponse = TypedKVValue
