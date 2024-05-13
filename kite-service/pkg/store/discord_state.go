package store

import (
	"context"

	"github.com/merlinfuchs/dismod/disrest"
	"github.com/merlinfuchs/dismod/distype"
)

type AppProvider interface {
	AppState(appID distype.Snowflake) (AppStateProvider, error)
}

type AppStateProvider interface {
	State() DiscordStateStore
	Client() *disrest.Client
}

type DiscordStateStore interface {
	GetGuild(ctx context.Context, guildID distype.Snowflake) (*distype.Guild, error)
	GetGuildChannels(ctx context.Context, guildID distype.Snowflake) ([]distype.Channel, error)
	GetGuildRoles(ctx context.Context, guildID distype.Snowflake) ([]distype.Role, error)
	GetChannel(ctx context.Context, channelID distype.Snowflake) (*distype.Channel, error)
	GetRole(ctx context.Context, guildID distype.Snowflake, roleID distype.Snowflake) (*distype.Role, error)
	GetGuildBotMember(ctx context.Context, guildID distype.Snowflake) (*distype.Member, error)
	GetGuildMember(ctx context.Context, guildID distype.Snowflake, userID distype.Snowflake) (*distype.Member, error)
	GetGuildOwnerID(ctx context.Context, guildID distype.Snowflake) (distype.Snowflake, error)
}
