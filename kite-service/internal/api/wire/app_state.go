package wire

import (
	"regexp"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	IconURL        null.String `json:"icon_url"`
	CreatedAt      time.Time   `json:"created_at"`
	BotPermissions null.String `json:"bot_permissions,omitempty"`
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

// StateGuildProfileUpdateRequest changes the bot's profile in a server. A
// field that isn't set stays as it is, and an empty one resets to the bot's
// own profile.
type StateGuildProfileUpdateRequest struct {
	Nick *string `json:"nick,omitempty"`
	Bio  *string `json:"bio,omitempty"`
	// Avatar and Banner are data URIs of the image.
	Avatar *string `json:"avatar,omitempty"`
	Banner *string `json:"banner,omitempty"`
}

var profileImageRegex = regexp.MustCompile(`^data:image/(png|jpeg|gif|webp);base64,`)

func (req StateGuildProfileUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Nick, validation.RuneLength(0, 32)),
		validation.Field(&req.Bio, validation.RuneLength(0, 190)),
		validation.Field(&req.Avatar, validation.Match(profileImageRegex)),
		validation.Field(&req.Banner, validation.Match(profileImageRegex)),
	)
}

type StateGuildProfileUpdateResponse = Empty

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

// StateGuildEmojis are the custom emojis of one guild the app is in. The app
// can use them in messages and components like its own emojis.
type StateGuildEmojis struct {
	GuildID   string      `json:"guild_id"`
	GuildName string      `json:"guild_name"`
	Emojis    []*AppEmoji `json:"emojis"`
}

type StateEmojiListResponse = []*StateGuildEmojis
