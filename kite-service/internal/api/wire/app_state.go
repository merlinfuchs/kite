package wire

import (
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"gopkg.in/guregu/null.v4"
)

type AppStateStatus struct {
	Online bool `json:"online"`
}

type StateStatusGetResponse = AppStateStatus

// Guild is a server the app is in. MemberCount is approximate and only set
// when the list is requested with counts, JoinedAt is when the app joined.
type Guild struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	IconURL     null.String `json:"icon_url"`
	OwnerID     string      `json:"owner_id"`
	MemberCount null.Int    `json:"member_count"`
	JoinedAt    null.Time   `json:"joined_at"`
	CreatedAt   time.Time   `json:"created_at"`
}

type StateGuildListResponse = []*Guild

func GuildToWire(guild *discord.Guild) *Guild {
	if guild == nil {
		return nil
	}

	iconURL := guild.IconURL()

	return &Guild{
		ID:          guild.ID.String(),
		Name:        guild.Name,
		Description: guild.Description,
		IconURL:     null.NewString(iconURL, iconURL != ""),
		OwnerID:     guild.OwnerID.String(),
		CreatedAt:   guild.CreatedAt(),
	}
}

type GuildOwner struct {
	ID          string      `json:"id"`
	Username    string      `json:"username"`
	DisplayName string      `json:"display_name"`
	AvatarURL   null.String `json:"avatar_url"`
}

type GuildDetails struct {
	Owner GuildOwner `json:"owner"`
	// Permissions is the bitset of the permissions the app has in the guild
	// through its roles, before channel overwrites.
	Permissions string `json:"permissions"`
}

type StateGuildGetResponse = GuildDetails

type Channel struct {
	ID    string `json:"id"`
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

type StateGuildChannelListResponse = []*Channel

type StateGuildLeaveResponse = Empty

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

type Role struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Managed roles belong to integrations, like other bots, and can't be
	// given to members.
	Managed  bool `json:"managed"`
	Position int  `json:"position"`
}

type StateGuildRoleListResponse = []*Role

func RoleToWire(role *discord.Role) *Role {
	return &Role{
		ID:       role.ID.String(),
		Name:     role.Name,
		Managed:  role.Managed,
		Position: role.Position,
	}
}
