package host

import (
	"context"
	"math"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/go-types/dismodel"
	"github.com/merlinfuchs/kite/go-types/fail"
)

func (h HostEnvironment) callDiscordBanList(ctx context.Context, guildID string, data dismodel.BanListCall) (dismodel.BanListResponse, error) {
	bans, err := h.bot.Session.GuildBans(guildID, data.Limit, data.Before, data.After, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.BanListResponse{}, modelError(err)
	}

	res := make([]dismodel.Ban, len(bans))
	for i, ban := range bans {
		res[i] = dismodel.Ban{
			Reason: ban.Reason,
			User:   modelUser(ban.User),
		}
	}

	return res, nil
}

func (h HostEnvironment) callDiscordBanGet(ctx context.Context, guildID string, data dismodel.BanGetCall) (dismodel.BanGetResponse, error) {
	ban, err := h.bot.Session.GuildBan(guildID, data.UserID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.BanGetResponse{}, modelError(err)
	}

	return dismodel.Ban{
		Reason: ban.Reason,
		User:   modelUser(ban.User),
	}, nil
}

func (h HostEnvironment) callDiscordBanCreate(ctx context.Context, guildID string, data dismodel.BanCreateCall) (dismodel.BanCreateResponse, error) {
	days := int(math.Max(0, math.Min(7, float64(data.DeleteMessageSeconds/86400))))

	err := h.bot.Session.GuildBanCreate(guildID, data.UserID, days, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.BanCreateResponse{}, modelError(err)
	}

	return dismodel.BanCreateResponse{}, nil
}

func (h HostEnvironment) callDiscordBanRemove(ctx context.Context, guildID string, data dismodel.BanRemoveCall) (dismodel.BanRemoveResponse, error) {
	err := h.bot.Session.GuildBanDelete(guildID, data.UserID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.BanRemoveResponse{}, modelError(err)
	}

	return dismodel.BanRemoveResponse{}, nil
}

func (h HostEnvironment) callDiscordChannelGet(ctx context.Context, guildID string, data dismodel.ChannelGetCall) (dismodel.ChannelGetResponse, error) {
	channel, err := h.bot.Session.Channel(data.ID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.ChannelGetResponse{}, modelError(err)
	}

	return modelChannel(channel), nil
}

func (h HostEnvironment) callDiscordChannelList(ctx context.Context, guildID string, data dismodel.ChannelListCall) (dismodel.ChannelListResponse, error) {
	channels, err := h.bot.Session.GuildChannels(guildID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.ChannelListResponse{}, modelError(err)
	}

	res := make([]dismodel.Channel, len(channels))
	for i, channel := range channels {
		res[i] = modelChannel(channel)
	}

	return res, nil
}

func (h HostEnvironment) callDiscordChannelCreate(ctx context.Context, guildID string, data dismodel.ChannelCreateCall) (dismodel.ChannelCreateResponse, error) {
	channel, err := h.bot.Session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:  data.Name,
		Topic: data.Topic,
		NSFW:  data.NSFW,
	}, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.ChannelCreateResponse{}, modelError(err)
	}

	return modelChannel(channel), nil
}

func (h HostEnvironment) callDiscordChannelUpdate(ctx context.Context, guildID string, data dismodel.ChannelUpdateCall) (dismodel.ChannelUpdateResponse, error) {
	channel, err := h.bot.Session.ChannelEditComplex(data.ID, &discordgo.ChannelEdit{
		// TODO
	}, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.ChannelUpdateResponse{}, modelError(err)
	}

	return modelChannel(channel), nil
}

func (h HostEnvironment) callDiscordChannelUpdatePositions(ctx context.Context, guildID string, data dismodel.ChannelUpdatePositionsCall) (dismodel.ChannelUpdatePositionsResponse, error) {
	return dismodel.ChannelUpdatePositionsResponse{}, fail.NewHostError(fail.HostErrorTypeUnimplemented)
}

func (h HostEnvironment) callDiscordChannelDelete(ctx context.Context, guildID string, data dismodel.ChannelDeleteCall) (dismodel.ChannelDeleteResponse, error) {
	channel, err := h.bot.Session.ChannelDelete(data.ID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.ChannelDeleteResponse{}, modelError(err)
	}

	return modelChannel(channel), nil
}

func (h HostEnvironment) callDiscordChannelUpdatePermissions(ctx context.Context, guildID string, data dismodel.ChannelUpdatePermissionsCall) (dismodel.ChannelUpdatePermissionsResponse, error) {
	return dismodel.ChannelUpdatePermissionsResponse{}, fail.NewHostError(fail.HostErrorTypeUnimplemented)
}

func (h HostEnvironment) callDiscordChannelDeletePermissions(ctx context.Context, guildID string, data dismodel.ChannelDeletePermissionsCall) (dismodel.ChannelDeletePermissionsResponse, error) {
	return dismodel.ChannelDeletePermissionsResponse{}, fail.NewHostError(fail.HostErrorTypeUnimplemented)
}

func (h HostEnvironment) callDiscordMessageCreate(ctx context.Context, guildID string, data dismodel.MessageCreateCall) (dismodel.MessageCreateResponse, error) {
	msg, err := h.bot.Session.ChannelMessageSend(data.ChannelID, data.Content, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.MessageCreateResponse{}, err
	}

	return modelMessage(msg), nil
}

func (h HostEnvironment) callDiscordMessageUpdate(ctx context.Context, guildID string, data dismodel.MessageUpdateCall) (dismodel.MessageUpdateResponse, error) {
	msg, err := h.bot.Session.ChannelMessageEdit(data.ChannelID, data.MessageID, data.Content, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.MessageUpdateResponse{}, modelError(err)
	}

	return modelMessage(msg), nil
}

func (h HostEnvironment) callDiscordMessageDelete(ctx context.Context, guildID string, data dismodel.MessageDeleteCall) (dismodel.MessageDeleteResponse, error) {
	err := h.bot.Session.ChannelMessageDelete(data.ChannelID, data.MessageID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.MessageDeleteResponse{}, modelError(err)
	}

	return dismodel.MessageDeleteResponse{}, nil
}

func (h HostEnvironment) callDiscordMessageGet(ctx context.Context, guildID string, data dismodel.MessageGetCall) (dismodel.MessageGetResponse, error) {
	msg, err := h.bot.Session.ChannelMessage(data.ChannelID, data.MessageID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.MessageGetResponse{}, modelError(err)
	}

	return modelMessage(msg), nil
}

func (h HostEnvironment) callDiscordInteractionResponseCreate(ctx context.Context, guildID string, data dismodel.InteractionResponseCreateCall) (dismodel.InteractionResponseCreateResponse, error) {
	err := h.bot.Session.InteractionRespond(&discordgo.Interaction{
		ID:    data.ID,
		Token: data.Token,
	}, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: data.Data.Content,
		},
	})

	return dismodel.InteractionResponseCreateResponse{}, err
}

func (h HostEnvironment) callDiscordGuildGet(ctx context.Context, guildID string, data dismodel.GuildGetCall) (dismodel.GuildGetResponse, error) {
	guild, err := h.bot.Session.Guild(guildID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.GuildGetResponse{}, modelError(err)
	}

	return modelGuild(guild), nil
}

func (h HostEnvironment) callDiscordRoleList(ctx context.Context, guildID string, data dismodel.RoleListCall) (dismodel.RoleListResponse, error) {
	roles, err := h.bot.Session.GuildRoles(guildID, discordgo.WithContext(ctx))
	if err != nil {
		return dismodel.RoleListResponse{}, modelError(err)
	}

	res := make([]dismodel.Role, len(roles))
	for i, role := range roles {
		res[i] = modelRole(role)
	}

	return res, nil
}
