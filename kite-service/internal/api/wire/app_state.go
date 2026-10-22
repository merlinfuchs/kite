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
	// AvailableTags are the tags posts in a forum or media channel can have.
	AvailableTags []ForumTag `json:"available_tags,omitempty"`
	// RequireTag is whether forum posts must have at least one tag.
	RequireTag bool `json:"require_tag,omitempty"`
}

type ForumTag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// Moderated tags can only be applied by members with Manage Threads.
	Moderated bool `json:"moderated"`
}

type StateGuildChannelListResponse = []*Channel

type StateGuildLeaveResponse = Empty

func ChannelToWire(channel *discord.Channel) *Channel {
	if channel == nil {
		return nil
	}

	var tags []ForumTag
	for _, tag := range channel.AvailableTags {
		tags = append(tags, ForumTag{
			ID:        tag.ID.String(),
			Name:      tag.Name,
			Moderated: tag.Moderated,
		})
	}

	return &Channel{
		ID:            channel.ID.String(),
		Type:          int(channel.Type),
		Name:          channel.Name,
		Topic:         channel.Topic,
		AvailableTags: tags,
		RequireTag:    channel.Flags&discord.ThreadRequireTag != 0,
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
