package store

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type DiscordStateStore interface {
	GetGuildBotMember(ctx context.Context, guildID string) (*discordgo.Member, error)
	GetGuildMember(ctx context.Context, guildID string, userID string) (*discordgo.Member, error)
	GetGuildOwnerID(ctx context.Context, guildID string) (string, error)
	GetGuildRoles(ctx context.Context, guildID string) ([]*discordgo.Role, error)
}
