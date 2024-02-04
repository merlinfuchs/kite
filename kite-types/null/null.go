package null

import "encoding/json"

type Null[T any] struct {
	Valid bool
	Value T
}

func (n *Null[T]) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		n.Valid = false
		return nil
	}

	n.Valid = true
	return json.Unmarshal(data, &n.Value)
}

func (n Null[T]) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(n.Value)
}
