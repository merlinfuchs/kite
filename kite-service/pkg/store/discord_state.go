package store

import (
	"context"

	"github.com/merlinfuchs/dismod/distype"
)

type DiscordStateStore interface {
	GetGuildBotMember(ctx context.Context, guildID distype.Snowflake) (*distype.Member, error)
	GetGuildMember(ctx context.Context, guildID distype.Snowflake, userID distype.Snowflake) (*distype.Member, error)
	GetGuildOwnerID(ctx context.Context, guildID distype.Snowflake) (distype.Snowflake, error)
	GetGuildRoles(ctx context.Context, guildID distype.Snowflake) ([]distype.Role, error)
}
