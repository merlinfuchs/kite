package wire

import (
	"encoding/base64"
	"encoding/json"
)

type APIResponse[Data any] struct {
	Success bool   `json:"success"`
	Data    Data   `json:"data"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Status  int         `json:"-"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e Error) MarshalJSON() ([]byte, error) {
	type Alias Error

	wrapped := struct {
		Success bool  `json:"success"`
		Error   Alias `json:"error,omitempty"`
	}{
		Success: false,
		Error:   Alias(e),
	}

	return json.Marshal(wrapped)
}

type Base64 []byte

func (b Base64) MarshalJSON() ([]byte, error) {
	d := base64.StdEncoding.EncodeToString(b)
	return json.Marshal(d)
}

func (b *Base64) UnMarshalJSON(d []byte) error {
	var s string
	err := json.Unmarshal(d, &s)
	if err != nil {
		return err
	}

	*b, err = base64.StdEncoding.DecodeString(s)
	return err
}
