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

type Guild struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	IconURL     null.String `json:"icon_url"`
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
		CreatedAt:   guild.CreatedAt(),
	}
}

type Channel struct {
	ID    string `json:"id"`
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

type StateGuildChannelListResponse = []*Channel

type StateGuildLeaveResponse = Empty

type StateGuildInviteCreateResponse struct {
	URL string `json:"url"`
}

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
