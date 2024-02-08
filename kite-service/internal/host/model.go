package host

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/dismod/distype"

	"github.com/merlinfuchs/kite/kite-types/fail"
)

func modelMessage(msg *discordgo.Message) distype.Message {
	return distype.Message{
		ID:        distype.Snowflake(msg.ID),
		ChannelID: distype.Snowflake(msg.ChannelID),
		Content:   msg.Content,
	}
}

func modelChannel(channel *discordgo.Channel) distype.Channel {
	return distype.Channel{
		ID:   distype.Snowflake(channel.ID),
		Name: &channel.Name,
	}
}

func modelUser(user *discordgo.User) distype.User {
	return distype.User{
		ID:       distype.Snowflake(user.ID),
		Username: user.Username,
	}
}

func modelGuild(guild *discordgo.Guild) distype.Guild {
	return distype.Guild{
		ID:   distype.Snowflake(guild.ID),
		Name: guild.Name,
	}
}

func modelRole(role *discordgo.Role) distype.Role {
	return distype.Role{
		ID:   distype.Snowflake(role.ID),
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
