package wire

import (
	"regexp"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
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
	// Empty if the bot has no banner.
	BannerURL string `json:"banner_url"`
}

func StateUserToWire(user *discord.User) *StateUser {
	return &StateUser{
		ID:        user.ID.String(),
		Username:  user.Username,
		AvatarURL: user.AvatarURL(),
		BannerURL: user.BannerURL(),
	}
}

type StateUserGetResponse = StateUser

// StateGuildProfile is the bot's profile in a server. Fields are empty where
// it uses its own profile. Discord doesn't return the About Me.
type StateGuildProfile struct {
	Nick      string     `json:"nick"`
	AvatarURL string     `json:"avatar_url"`
	BannerURL string     `json:"banner_url"`
	NameStyle *NameStyle `json:"name_style"`
}

// NameStyle is the font, effect and colors of the bot's name in a server, by
// the names of provider.NameStyleFonts and provider.NameStyleEffects.
type NameStyle struct {
	Font   string `json:"font"`
	Effect string `json:"effect"`
	Colors []int  `json:"colors"`
}

func (s NameStyle) Validate() error {
	_, err := s.ToProvider()
	return err
}

func (s NameStyle) ToProvider() (*provider.NameStyle, error) {
	return provider.NewNameStyle(s.Font, s.Effect, s.Colors)
}

// NameStyleToWire returns nil for styles with a font or effect Kite doesn't
// offer, like the ones Discord no longer offers.
func NameStyleToWire(style *provider.NameStyle) *NameStyle {
	if style == nil {
		return nil
	}
	font, effect, ok := style.Names()
	if !ok {
		return nil
	}
	return &NameStyle{Font: font, Effect: effect, Colors: style.Colors}
}

type StateGuildProfileGetResponse = StateGuildProfile

// StateUserUpdateRequest changes the bot's avatar and banner. A field that
// isn't set stays as it is, and an empty one removes the image.
type StateUserUpdateRequest struct {
	// Avatar and Banner are data URIs of the image.
	Avatar *string `json:"avatar,omitempty"`
	Banner *string `json:"banner,omitempty"`
}

func (req StateUserUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Avatar, validation.Match(profileImageRegex)),
		validation.Field(&req.Banner, validation.Match(profileImageRegex)),
	)
}

type StateUserUpdateResponse = StateUser

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
// own profile. Reset resets all of them, the other fields are ignored then.
type StateGuildProfileUpdateRequest struct {
	Nick *string `json:"nick,omitempty"`
	Bio  *string `json:"bio,omitempty"`
	// Avatar and Banner are data URIs of the image.
	Avatar    *string    `json:"avatar,omitempty"`
	Banner    *string    `json:"banner,omitempty"`
	NameStyle *NameStyle `json:"name_style,omitempty"`
	Reset     bool       `json:"reset,omitempty"`
}

var profileImageRegex = regexp.MustCompile(`^data:image/(png|jpeg|gif|webp);base64,`)

func (req StateGuildProfileUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Nick, validation.RuneLength(0, 32)),
		validation.Field(&req.Bio, validation.RuneLength(0, 190)),
		validation.Field(&req.Avatar, validation.Match(profileImageRegex)),
		validation.Field(&req.Banner, validation.Match(profileImageRegex)),
		validation.Field(&req.NameStyle),
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
