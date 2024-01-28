package store

import (
	"context"

	"github.com/merlinfuchs/kite/go-types/dismodel"
)

type DiscordStateStore interface {
	GetGuildBotMember(ctx context.Context, guildID string) (*dismodel.Member, error)
	GetGuildMember(ctx context.Context, guildID string, userID string) (*dismodel.Member, error)
	GetGuildOwnerID(ctx context.Context, guildID string) (string, error)
	GetGuildRoles(ctx context.Context, guildID string) ([]*dismodel.Role, error)
}
