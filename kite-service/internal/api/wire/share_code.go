package wire

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

type ShareCode struct {
	Code string          `json:"code"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ShareCodeCreateRequest struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (req ShareCodeCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Type, validation.Required, validation.In(
			string(model.ShareCodeTypeCommand),
			string(model.ShareCodeTypeEventListener),
		)),
		validation.Field(&req.Data, validation.Required, validation.Length(1, 512*1024)),
	)
}

type ShareCodeCreateResponse struct {
	Code string `json:"code"`
}

type ShareCodeGetResponse = ShareCode

func ShareCodeToWire(shareCode *model.ShareCode) *ShareCode {
	if shareCode == nil {
		return nil
	}

	return &ShareCode{
		Code: shareCode.Code,
		Type: string(shareCode.Type),
		Data: shareCode.Data,
	}
}
