package kvmodel

import (
	"encoding/json"
	"fmt"
)

type KVValueType string

const (
	KVValueTypeString KVValueType = "STRING"
	KVValueTypeInt    KVValueType = "INT"
)

type TypedKVValue struct {
	Type  KVValueType `json:"type"`
	Value KVValue     `json:"value"`
}

func (t *TypedKVValue) UnmarshalJSON(b []byte) error {
	var raw struct {
		Type  KVValueType     `json:"type"`
		Value json.RawMessage `json:"value"`
	}

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	t.Type = raw.Type

	switch t.Type {
	case KVValueTypeString:
		var val KVString
		if err := json.Unmarshal(raw.Value, &val); err != nil {
			return err
		}
		t.Value = val
		break
	case KVValueTypeInt:
		var val KVInt
		if err := json.Unmarshal(raw.Value, &val); err != nil {
			return err
		}
		t.Value = val
		break
	default:
		return fmt.Errorf("unknown type %s", t.Type)
	}

	return nil
}

type KVValue interface {
	Type() KVValueType
	String() string
	Int() int
}

type KVString string

func (k KVString) Type() KVValueType {
	return KVValueTypeString
}

func (k KVString) String() string {
	return string(k)
}

func (k KVString) Int() int {
	return 0
}

type KVInt int

func (k KVInt) Type() KVValueType {
	return KVValueTypeInt
}

func (k KVInt) String() string {
	return fmt.Sprint(int(k))
}

func (k KVInt) Int() int {
	return int(k)
}
