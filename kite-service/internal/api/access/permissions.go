package access

import (
	"context"
	"slices"

	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type GuildPermissions struct {
	UserIsMember    bool
	UserIsOwner     bool
	UserPermissions uint64
	BotIsMember     bool
	BotPermissions  uint64
}

func (m *AccessManager) GetGuildPermissionsForUser(ctx context.Context, guildID distype.Snowflake, userID distype.Snowflake) (*GuildPermissions, error) {
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
		if role.ID == distype.Snowflake(guildID) || slices.Contains(member.Roles, role.ID) {
			res.UserPermissions |= role.Permissions.Parse()
		}

		if role.ID == distype.Snowflake(guildID) || slices.Contains(botMember.Roles, role.ID) {
			res.BotPermissions |= role.Permissions.Parse()
		}
	}

	return res, nil
}
