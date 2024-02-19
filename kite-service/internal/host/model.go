package host

import (
	"context"
	"errors"

	"github.com/disgoorg/disgo/rest"

	"github.com/merlinfuchs/kite/kite-sdk-go/fail"
)

func modelError(err error) *fail.HostError {
	var derr rest.Error
	if errors.As(err, &derr) {
		if derr.Code == 0 {
			return &fail.HostError{
				Code:    fail.HostErrorTypeDiscordUnknown,
				Message: derr.Error(),
			}
		}

		// TODO: map all error codes
		code := fail.HostErrorTypeDiscordUnknown
		switch derr.Code {
		case 10004:
			code = fail.HostErrorTypeDiscordGuildNotFound
		case 10003:
			code = fail.HostErrorTypeDiscordChannelNotFound
		case 10008:
			code = fail.HostErrorTypeDiscordMessageNotFound
		case 10026:
			code = fail.HostErrorTypeDiscordBanNotFound
		}

		return &fail.HostError{
			Code:    code,
			Message: derr.Message,
		}
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return &fail.HostError{
			Code:    fail.HostErrorTypeTimeout,
			Message: err.Error(),
		}
	}

	if errors.Is(err, context.Canceled) {
		return &fail.HostError{
			Code:    fail.HostErrorTypeCanceled,
			Message: err.Error(),
		}
	}

	return &fail.HostError{
		Message: err.Error(),
	}
}
