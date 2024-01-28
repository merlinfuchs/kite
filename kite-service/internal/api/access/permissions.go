package access

import (
	"context"
	"slices"
	"strconv"

	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type GuildPermissions struct {
	UserIsMember    bool
	UserIsOwner     bool
	UserPermissions int64
	BotIsMember     bool
	BotPermissions  int64
}

func (m *AccessManager) GetGuildPermissionsForUser(ctx context.Context, guildID string, userID string) (*GuildPermissions, error) {
	res := &GuildPermissions{}

	botMember, err := m.state.GetGuildBotMember(ctx, guildID)
	if err != nil {
		if err == store.ErrNotFound {
			return res, nil
		}
		return res, err
	}

	res.BotIsMember = true

	member, err := m.state.GetGuildMember(ctx, guildID, userID)
	if err != nil {
		if err == store.ErrNotFound {
			return res, nil
		}
		return res, err
	}

	res.UserIsMember = true

	ownerID, err := m.state.GetGuildOwnerID(ctx, guildID)
	if err != nil {
		return res, err
	}

	res.UserIsOwner = ownerID == userID

	roles, err := m.state.GetGuildRoles(ctx, guildID)
	if err != nil {
		return res, err
	}

	for _, role := range roles {
		if role.ID == guildID || slices.Contains(member.Roles, role.ID) {
			perms, _ := strconv.ParseInt(role.Permissions, 10, 64)
			res.UserPermissions |= perms
		}

		if role.ID == guildID || slices.Contains(botMember.Roles, role.ID) {
			perms, _ := strconv.ParseInt(role.Permissions, 10, 64)
			res.BotPermissions |= perms
		}
	}

	return res, nil
}
