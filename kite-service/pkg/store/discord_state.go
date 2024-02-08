package store

import (
	"context"

	"github.com/merlinfuchs/dismod/distype"
)

type DiscordStateStore interface {
	GetGuildBotMember(ctx context.Context, guildID string) (*distype.Member, error)
	GetGuildMember(ctx context.Context, guildID string, userID string) (*distype.Member, error)
	GetGuildOwnerID(ctx context.Context, guildID string) (string, error)
	GetGuildRoles(ctx context.Context, guildID string) ([]*distype.Role, error)
}
