package fail

import (
	"errors"
	"fmt"
)

type HostErrorType int

const (
	HostErrorTypeUnknown                HostErrorType = 0
	HostErrorTypeTimeout                HostErrorType = 1
	HostErrorTypeCanceled               HostErrorType = 2
	HostErrorTypeUnimplemented          HostErrorType = 3
	HostErrorTypeValidationFailed       HostErrorType = 4
	HostErrorTypeGuildAccessMissing     HostErrorType = 5
	HostErrorTypeDiscordUnknown         HostErrorType = 100
	HostErrorTypeDiscordGuildNotFound   HostErrorType = 101
	HostErrorTypeDiscordChannelNotFound HostErrorType = 102
	HostErrorTypeDiscordMessageNotFound HostErrorType = 103
	HostErrorTypeDiscordBanNotFound     HostErrorType = 104
	HostErrorTypeKVUnknown              HostErrorType = 200
	HostErrorTypeKVKeyNotFound          HostErrorType = 201
	HostErrorTypeKVValueTypeMismatch    HostErrorType = 202
)

type HostError struct {
	Code    HostErrorType `json:"code"`
	Message string        `json:"message"`
}

func (e *HostError) Error() string {
	return fmt.Sprintf("Host error %d: %s", e.Code, e.Message)
}

func NewHostError(code HostErrorType, message ...string) *HostError {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}

	return &HostError{
		Code:    code,
		Message: msg,
	}
}

func IsHostErrorCode(err error, code HostErrorType) bool {
	if err == nil {
		return false
	}

	var hostError *HostError
	ok := errors.As(err, &hostError)
	if !ok {
		return false
	}

	return hostError.Code == code
}
