package host

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/dismod/distype"
)

func (h HostEnvironment) callDiscordChannelGet(ctx context.Context, data distype.ChannelGetRequest) (distype.ChannelGetResponse, error) {
	channel, err := h.bot.Session.Channel(data.ChannelID.String(), discordgo.WithContext(ctx))
	if err != nil {
		return distype.ChannelGetResponse{}, modelError(err)
	}

	return modelChannel(channel), nil
}

func (h HostEnvironment) callDiscordChannelList(ctx context.Context, data distype.GuildChannelListRequest) (distype.GuildChannelListResponse, error) {
	channels, err := h.bot.Session.GuildChannels(data.GuildID.String(), discordgo.WithContext(ctx))
	if err != nil {
		return distype.GuildChannelListResponse{}, modelError(err)
	}

	res := make([]distype.Channel, len(channels))
	for i, channel := range channels {
		res[i] = modelChannel(channel)
	}

	return res, nil
}

func (h HostEnvironment) callDiscordMessageCreate(ctx context.Context, data distype.MessageCreateRequest) (distype.MessageCreateResponse, error) {
	var content string
	if data.Content != nil {
		content = *data.Content
	}

	msg, err := h.bot.Session.ChannelMessageSend(data.ChannelID.String(), content, discordgo.WithContext(ctx))
	if err != nil {
		return distype.MessageCreateResponse{}, err
	}

	return modelMessage(msg), nil
}

func (h HostEnvironment) callDiscordMessageUpdate(ctx context.Context, data distype.MessageEditRequest) (distype.MessageEditResponse, error) {
	var content string
	if data.Content != nil {
		content = *data.Content
	}

	msg, err := h.bot.Session.ChannelMessageEdit(data.ChannelID.String(), data.MessageID.String(), content, discordgo.WithContext(ctx))
	if err != nil {
		return distype.MessageEditResponse{}, modelError(err)
	}

	return modelMessage(msg), nil
}

func (h HostEnvironment) callDiscordMessageDelete(ctx context.Context, data distype.MessageDeleteRequest) (distype.MessageDeleteResponse, error) {
	err := h.bot.Session.ChannelMessageDelete(data.ChannelID.String(), data.MessageID.String(), discordgo.WithContext(ctx))
	if err != nil {
		return distype.MessageDeleteResponse{}, modelError(err)
	}

	return distype.MessageDeleteResponse{}, nil
}

func (h HostEnvironment) callDiscordMessageGet(ctx context.Context, data distype.MessageGetRequest) (distype.MessageGetResponse, error) {
	msg, err := h.bot.Session.ChannelMessage(data.ChannelID.String(), data.MessageID.String(), discordgo.WithContext(ctx))
	if err != nil {
		return distype.MessageGetResponse{}, modelError(err)
	}

	return modelMessage(msg), nil
}

func (h HostEnvironment) callDiscordInteractionResponseCreate(ctx context.Context, data distype.InteractionResponseCreateRequest) (distype.InteractionResponseCreateResponse, error) {
	err := h.bot.Session.InteractionRespond(&discordgo.Interaction{
		ID:    data.InteractionID.String(),
		Token: data.InteractionToken,
	}, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			// TODO: Content: data.Data.Content,
		},
	})

	return distype.InteractionResponseCreateResponse{}, err
}

func (h HostEnvironment) callDiscordGuildGet(ctx context.Context, data distype.GuildGetRequest) (distype.GuildGetResponse, error) {
	guild, err := h.bot.Session.Guild(h.GuildID, discordgo.WithContext(ctx))
	if err != nil {
		return distype.GuildGetResponse{}, modelError(err)
	}

	return modelGuild(guild), nil
}
