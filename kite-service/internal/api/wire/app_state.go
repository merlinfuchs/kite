package wire

import (
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"gopkg.in/guregu/null.v4"
)

type AppStateStatus struct {
	Online bool `json:"online"`
}

type StateStatusGetResponse = AppStateStatus

// StateUser is the Discord user the app's bot sends messages as. It's empty
// while the app isn't running.
type StateUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type StateUserGetResponse = StateUser

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
	// ParentID is the category a channel is in, or the channel a thread is in.
	ParentID string `json:"parent_id,omitempty"`
	Position int    `json:"position,omitempty"`
	// BotPermissions is the bot's permission bit set in the channel, as a
	// string because it doesn't fit in a JavaScript number. It's empty when
	// the permissions couldn't be worked out.
	BotPermissions string `json:"bot_permissions,omitempty"`
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

	var parentID string
	if channel.ParentID.IsValid() {
		parentID = channel.ParentID.String()
	}

	return &Channel{
		ID:            channel.ID.String(),
		Type:          int(channel.Type),
		Name:          channel.Name,
		Topic:         channel.Topic,
		ParentID:      parentID,
		Position:      channel.Position,
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
	// Color is the role's color as an RGB integer, or 0 for no color.
	Color int `json:"color"`
	// IconURL is the URL of the role's custom icon, if it has one.
	IconURL string `json:"icon_url,omitempty"`
	// UnicodeEmoji is the emoji used as the role's icon, if it has one.
	UnicodeEmoji string `json:"unicode_emoji,omitempty"`
	// Permissions is the role's permission bit set, as a string because it
	// doesn't fit in a JavaScript number.
	Permissions string `json:"permissions"`
	// Hoist is whether members with the role are shown separately.
	Hoist       bool      `json:"hoist"`
	Mentionable bool      `json:"mentionable"`
	CreatedAt   time.Time `json:"created_at"`
}

type StateGuildRoleListResponse = []*Role

func RoleToWire(role *discord.Role) *Role {
	return &Role{
		ID:           role.ID.String(),
		Name:         role.Name,
		Managed:      role.Managed,
		Position:     role.Position,
		Color:        role.Color.Int(),
		IconURL:      role.IconURL(),
		UnicodeEmoji: role.UnicodeEmoji,
		Permissions:  strconv.FormatUint(uint64(role.Permissions), 10),
		Hoist:        role.Hoist,
		Mentionable:  role.Mentionable,
		CreatedAt:    role.CreatedAt(),
	}
}

type Member struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	// DisplayName is the member's server nickname, or their global display
	// name, or their username.
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bot         bool      `json:"bot"`
	RoleIDs     []string  `json:"role_ids"`
	JoinedAt    null.Time `json:"joined_at"`
}

// StateGuildMemberListResponse holds at most MaxStateGuildMembers members.
type StateGuildMemberListResponse = []*Member

// MaxStateGuildMembers is how many members the server explorer lists for a
// server, so a huge server doesn't need hundreds of requests to Discord.
const MaxStateGuildMembers = 1000

func MemberToWire(member *discord.Member, guildID discord.GuildID) *Member {
	displayName := member.Nick
	if displayName == "" {
		displayName = member.User.DisplayOrUsername()
	}

	roleIDs := make([]string, len(member.RoleIDs))
	for i, id := range member.RoleIDs {
		roleIDs[i] = id.String()
	}

	// Members without a server avatar use their own, or the default one.
	avatarURL := member.AvatarURL(guildID)
	if avatarURL == "" {
		avatarURL = member.User.AvatarURL()
	}

	return &Member{
		ID:          member.User.ID.String(),
		Username:    member.User.Username,
		DisplayName: displayName,
		AvatarURL:   avatarURL,
		Bot:         member.User.Bot,
		RoleIDs:     roleIDs,
		JoinedAt:    null.NewTime(member.Joined.Time(), member.Joined.IsValid()),
	}
}
