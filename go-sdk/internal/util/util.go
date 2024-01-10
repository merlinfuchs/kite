package util

import "encoding/json"

func DecodeT[T any](raw []byte) (T, error) {
	var data T
	err := json.Unmarshal(raw, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}
