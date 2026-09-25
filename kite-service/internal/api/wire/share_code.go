package wire

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type ShareCodeCreateRequest struct {
	Data string `json:"data"`
}

func (req ShareCodeCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Data, validation.Required, validation.Length(1, 512*1024)),
	)
}

type ShareCodeCreateResponse struct {
	Code string `json:"code"`
}

type ShareCodeGetResponse struct {
	Data string `json:"data"`
}
