package host

import (
	"context"

	"github.com/disgoorg/disgo/rest"
	"github.com/merlinfuchs/dismod/distype"
)

func (h HostEnvironment) callDiscordChannelGet(ctx context.Context, data distype.ChannelGetRequest) (*distype.ChannelGetResponse, error) {
	channel, err := h.discordState.GetChannel(ctx, data.ChannelID)
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (h HostEnvironment) callDiscordChannelList(ctx context.Context, data distype.GuildChannelListRequest) (*distype.GuildChannelListResponse, error) {
	channels, err := h.discordState.GetGuildChannels(ctx, data.GuildID)
	if err != nil {
		return nil, err
	}

	return &channels, nil
}

func (h HostEnvironment) callDiscordMessageCreate(ctx context.Context, data distype.MessageCreateRequest) (*distype.MessageCreateResponse, error) {
	var res *distype.MessageCreateResponse
	err := h.discordClient.Request(
		rest.CreateMessage.Compile(nil, data.ChannelID),
		data, &res, rest.WithCtx(ctx),
	)
	return res, err
}

func (h HostEnvironment) callDiscordMessageUpdate(ctx context.Context, data distype.MessageEditRequest) (*distype.MessageEditResponse, error) {
	var res *distype.MessageEditResponse
	err := h.discordClient.Request(
		rest.UpdateMessage.Compile(nil, data.ChannelID, data.MessageID),
		data, &res, rest.WithCtx(ctx),
	)
	return res, err
}

func (h HostEnvironment) callDiscordMessageDelete(ctx context.Context, data distype.MessageDeleteRequest) (*distype.MessageDeleteResponse, error) {
	err := h.discordClient.Request(
		rest.DeleteMessage.Compile(nil, data.ChannelID, data.MessageID),
		nil, nil, rest.WithCtx(ctx),
	)
	return &distype.MessageDeleteResponse{}, err
}

func (h HostEnvironment) callDiscordMessageGet(ctx context.Context, data distype.MessageGetRequest) (*distype.MessageGetResponse, error) {
	var res *distype.MessageGetResponse
	err := h.discordClient.Request(
		rest.GetMessage.Compile(nil, data.ChannelID, data.MessageID),
		nil, &res, rest.WithCtx(ctx),
	)
	return res, err
}

func (h HostEnvironment) callDiscordInteractionResponseCreate(ctx context.Context, data distype.InteractionResponseCreateRequest) (*distype.InteractionResponseCreateResponse, error) {
	err := h.discordClient.Request(
		rest.CreateInteractionResponse.Compile(nil, data.InteractionID, data.InteractionToken),
		data, nil, rest.WithCtx(ctx),
	)
	return &distype.InteractionResponseCreateResponse{}, err
}

func (h HostEnvironment) callDiscordGuildGet(ctx context.Context, data distype.GuildGetRequest) (*distype.GuildGetResponse, error) {
	return nil, nil
}
