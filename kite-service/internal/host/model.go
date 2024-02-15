package host

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"

	"github.com/merlinfuchs/kite/kite-sdk-go/fail"
)

func modelError(err error) *fail.HostError {
	var derr *discordgo.RESTError
	if errors.As(err, &derr) {
		if derr.Message == nil {
			return &fail.HostError{
				Code:    fail.HostErrorTypeDiscordUnknown,
				Message: derr.Error(),
			}
		}

		code := fail.HostErrorTypeDiscordUnknown
		switch derr.Message.Code {
		case discordgo.ErrCodeUnknownGuild:
			code = fail.HostErrorTypeDiscordGuildNotFound
		case discordgo.ErrCodeUnknownChannel:
			code = fail.HostErrorTypeDiscordChannelNotFound
		case discordgo.ErrCodeUnknownMessage:
			code = fail.HostErrorTypeDiscordMessageNotFound
		case discordgo.ErrCodeUnknownBan:
			code = fail.HostErrorTypeDiscordBanNotFound
		}

		return &fail.HostError{
			Code:    code,
			Message: derr.Message.Message,
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
