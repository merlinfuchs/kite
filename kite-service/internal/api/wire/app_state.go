package wire

import "github.com/diamondburned/arikawa/v3/discord"

type Guild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type StateGuildListResponse = []*Guild

func GuildToWire(guild *discord.Guild) *Guild {
	if guild == nil {
		return nil
	}

	return &Guild{
		ID:          guild.ID.String(),
		Name:        guild.Name,
		Description: guild.Description,
	}
}

type Channel struct {
	ID    string `json:"id"`
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

type StateGuildChannelListResponse = []*Channel

func ChannelToWire(channel *discord.Channel) *Channel {
	if channel == nil {
		return nil
	}

	return &Channel{
		ID:    channel.ID.String(),
		Type:  int(channel.Type),
		Name:  channel.Name,
		Topic: channel.Topic,
	}
}
