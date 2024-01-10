package host

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/fail"
)

func modelMessage(msg *discordgo.Message) dismodel.Message {
	return dismodel.Message{
		ID:        msg.ID,
		ChannelID: msg.ChannelID,
		Content:   msg.Content,
	}
}

func modelChannel(channel *discordgo.Channel) dismodel.Channel {
	return dismodel.Channel{
		ID:   channel.ID,
		Name: channel.Name,
	}
}

func modelUser(user *discordgo.User) dismodel.User {
	return dismodel.User{
		ID:       user.ID,
		Username: user.Username,
	}
}

func modelGuild(guild *discordgo.Guild) dismodel.Guild {
	return dismodel.Guild{
		ID:   guild.ID,
		Name: guild.Name,
	}
}

func modelRole(role *discordgo.Role) dismodel.Role {
	return dismodel.Role{
		ID:   role.ID,
		Name: role.Name,
	}
}

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
