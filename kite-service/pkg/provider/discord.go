package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// DiscordProvider provides access to the Discord API.
type DiscordProvider interface {
	Guild(ctx context.Context, guildID discord.GuildID) (*discord.Guild, error)
	GuildRoles(ctx context.Context, guildID discord.GuildID) ([]discord.Role, error)
	Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error)
	User(ctx context.Context, userID discord.UserID) (*discord.User, error)
	Role(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) (*discord.Role, error)
	Member(ctx context.Context, guildID discord.GuildID, userID discord.UserID) (*discord.Member, error)
	Message(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error)

	CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*InteractionResponseResource, error)
	EditInteractionResponse(ctx context.Context, applicationID discord.AppID, token string, response api.EditInteractionResponseData) (*discord.Message, error)
	DeleteInteractionResponse(ctx context.Context, applicationID discord.AppID, token string) error
	CreateInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, data api.InteractionResponseData) (*discord.Message, error)
	EditInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID, data api.EditInteractionResponseData) (*discord.Message, error)
	DeleteInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID) error
	CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error)
	// CreatePersonaMessage sends a message with another name and avatar
	// through a webhook of the app in the channel.
	CreatePersonaMessage(ctx context.Context, channelID discord.ChannelID, persona Persona, message api.SendMessageData) (*discord.Message, error)
	EditMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, message api.EditMessageData) (*discord.Message, error)
	CreateMessageReaction(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, emoji discord.APIEmoji) error
	CreatePoll(ctx context.Context, channelID discord.ChannelID, data CreatePollData) (*discord.Message, error)
	ForwardMessage(ctx context.Context, channelID discord.ChannelID, messageChannelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error)
	EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error
	CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error)
	EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) (*discord.Channel, error)
	CreatePrivateChannel(ctx context.Context, userID discord.UserID) (*discord.Channel, error)
	StartThreadWithMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.StartThreadData) (*discord.Channel, error)
	StartThreadWithoutMessage(ctx context.Context, channelID discord.ChannelID, data api.StartThreadData) (*discord.Channel, error)
	StartThreadInForum(ctx context.Context, channelID discord.ChannelID, data api.StartForumThreadData) (*discord.Channel, error)
	APIRequest(ctx context.Context, req DiscordAPIRequest) ([]byte, error)
	// BotToken is for services that call Discord for the app, like Cookie
	// API's transcripts, and only goes to integrations the app enabled.
	BotToken() string

	UpdateVoiceState(ctx context.Context, guildID discord.GuildID, channelID discord.ChannelID, selfMute bool, selfDeaf bool) error
	// PlayAudio joins the voice channel and plays the uploaded sound in the
	// background, replacing what the bot plays in the server. It returns once
	// playback started.
	PlayAudio(ctx context.Context, opts PlayAudioOpts) error
	// SpeakText turns the text into speech and plays it like PlayAudio.
	SpeakText(ctx context.Context, opts SpeakTextOpts) error
	// StopAudio stops what the bot plays in the server. The bot stays in the
	// voice channel.
	StopAudio(ctx context.Context, guildID discord.GuildID) error
	UpdatePresence(ctx context.Context, status discord.Status, activity discord.Activity) error
	LeaveGuild(ctx context.Context, guildID discord.GuildID) error
	EditServerProfile(ctx context.Context, guildID discord.GuildID, data ServerProfileData) error
	// AppVerifyKey is the app's public key as hex, which Discord signs
	// interactions and webhook events with.
	AppVerifyKey(ctx context.Context) (string, error)

	HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error)
	// MarkInteractionResponded is for interactions a previous execution
	// already responded to, e.g. before a durable sleep.
	MarkInteractionResponded(interactionID discord.InteractionID)
	AutoDeferInteraction(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse)
}

// PlayAudioOpts are the options of DiscordProvider.PlayAudio.
type PlayAudioOpts struct {
	GuildID   discord.GuildID
	ChannelID discord.ChannelID
	// AssetID is the ID of the asset of the uploaded sound.
	AssetID string
	// LeaveAfter makes the bot leave the voice channel once the sound ended.
	LeaveAfter bool
}

// SpeakTextOpts are the options of DiscordProvider.SpeakText.
type SpeakTextOpts struct {
	GuildID   discord.GuildID
	ChannelID discord.ChannelID
	Text      string
	// Language is a language code like en or pt-BR, empty for English.
	Language string
	// LeaveAfter makes the bot leave the voice channel once it spoke.
	LeaveAfter bool
}

// Persona is the name and avatar a message is sent with instead of the bot's.
type Persona struct {
	Name      string
	AvatarURL string
}

// ServerProfileData is the body of a Modify Current Member request, the bot's
// profile in one server. arikawa only supports its nickname. A field that's
// nil stays as it is, and one that's null resets to the bot's own profile.
type ServerProfileData struct {
	Nick option.NullableString `json:"nick,omitempty"`
	// Avatar and Banner are data URIs of the image.
	Avatar option.NullableString `json:"avatar,omitempty"`
	Banner option.NullableString `json:"banner,omitempty"`
	Bio    option.NullableString `json:"bio,omitempty"`
	// NameStyle changes the style of the bot's name, and an empty one removes
	// it. Discord ignores its fields unless all three are sent, so it's sent
	// as all of them by MarshalJSON.
	NameStyle *NameStyle `json:"-"`
}

func (d ServerProfileData) MarshalJSON() ([]byte, error) {
	type fields ServerProfileData
	body := struct {
		fields
		FontID   option.NullableInt `json:"display_name_font_id,omitempty"`
		EffectID option.NullableInt `json:"display_name_effect_id,omitempty"`
		Colors   *[]int             `json:"display_name_colors,omitempty"`
	}{fields: fields(d)}

	if style := d.NameStyle; style != nil {
		if style.FontID == 0 {
			body.FontID, body.EffectID, body.Colors = option.NullInt, option.NullInt, new([]int)
		} else {
			body.FontID = option.NewNullableInt(style.FontID)
			body.EffectID = option.NewNullableInt(style.EffectID)
			body.Colors = &style.Colors
		}
	}
	return json.Marshal(body)
}

// NameStyle is the font, effect and colors of a name, which Discord shows in
// the profile. Servers only show the font, and the role color instead of the
// colors.
type NameStyle struct {
	FontID   int   `json:"font_id"`
	EffectID int   `json:"effect_id"`
	Colors   []int `json:"colors"`
}

// NewNameStyle returns the style with the font and effect of the names Kite
// uses for them.
func NewNameStyle(font, effect string, colors []int) (*NameStyle, error) {
	fontID, ok := NameStyleFonts[font]
	if !ok {
		return nil, fmt.Errorf("unknown name font %q", font)
	}
	effectID, ok := NameStyleEffects[effect]
	if !ok {
		return nil, fmt.Errorf("unknown name effect %q", effect)
	}
	if len(colors) == 0 || len(colors) > MaxNameStyleColors {
		return nil, fmt.Errorf("a name style needs 1 to %d colors", MaxNameStyleColors)
	}
	for _, color := range colors {
		if color < 0 || color > 0xFFFFFF {
			return nil, fmt.Errorf("name color %d isn't a color", color)
		}
	}
	return &NameStyle{FontID: fontID, EffectID: effectID, Colors: colors}, nil
}

// Names returns the names Kite uses for the font and effect, and false for
// ones Kite doesn't offer.
func (s NameStyle) Names() (font string, effect string, ok bool) {
	for name, id := range NameStyleFonts {
		if id == s.FontID {
			font = name
		}
	}
	for name, id := range NameStyleEffects {
		if id == s.EffectID {
			effect = name
		}
	}
	return font, effect, font != "" && effect != ""
}

// StyledMember is a member as Discord sends it, with the style of its name,
// which arikawa doesn't know.
type StyledMember struct {
	discord.Member
	NameStyle *NameStyle `json:"display_name_styles"`
}

// MaxNameStyleColors is how many colors Discord takes for a name style.
const MaxNameStyleColors = 5

// NameStyleFonts are the fonts of name styles by the name Kite uses for them,
// which is the one Discord shows. Discord still takes some it no longer
// offers, which Kite leaves out.
var NameStyleFonts = map[string]int{
	"default":     11,
	"sakura":      3,
	"jellybean":   4,
	"modern":      6,
	"medieval":    7,
	"8bit":        8,
	"vampyre":     10,
	"tempo":       12,
	"monkey_bars": 13,
	"mainframe":   14,
	"headbang":    15,
	"journal":     16,
}

// NameStyleEffects are the effects of name styles by the name Kite uses for
// them.
var NameStyleEffects = map[string]int{
	"solid":    1,
	"gradient": 2,
	"neon":     3,
	"toon":     4,
	"pop":      5,
	"prism":    7,
	"gummy":    8,
}

// CreatePollData is the poll object of a create message request. arikawa
// doesn't support polls yet, so these mirror Discord's API types.
type CreatePollData struct {
	Question PollMedia    `json:"question"`
	Answers  []PollAnswer `json:"answers"`
	// Duration is the number of hours the poll is open for.
	Duration         int            `json:"duration"`
	AllowMultiselect bool           `json:"allow_multiselect"`
	LayoutType       PollLayoutType `json:"layout_type,omitempty"`
}

type PollLayoutType int

const PollLayoutTypeDefault PollLayoutType = 1

type PollAnswer struct {
	PollMedia PollMedia `json:"poll_media"`
}

type PollMedia struct {
	Text  string     `json:"text,omitempty"`
	Emoji *PollEmoji `json:"emoji,omitempty"`
}

// PollEmoji holds either the ID of a custom emoji or the unicode of a standard
// one, never both.
type PollEmoji struct {
	ID   discord.EmojiID `json:"id,omitempty"`
	Name string          `json:"name,omitempty"`
}

// DiscordAPIRequest is a request to any endpoint of the Discord API, sent
// with the bot's token. A response with an error status is returned as error.
type DiscordAPIRequest struct {
	Method string
	// Path is relative to api.Endpoint and includes the query.
	Path   string
	Body   []byte
	Reason api.AuditLogReason
}

type InteractionResponseResource struct {
	Type    api.InteractionResponseType
	Message *discord.Message
}

type MockDiscordProvider struct{}

func (p *MockDiscordProvider) Guild(ctx context.Context, guildID discord.GuildID) (*discord.Guild, error) {
	return nil, nil
}

func (p *MockDiscordProvider) GuildRoles(ctx context.Context, guildID discord.GuildID) ([]discord.Role, error) {
	return nil, nil
}

func (p *MockDiscordProvider) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) User(ctx context.Context, userID discord.UserID) (*discord.User, error) {
	return nil, nil
}

func (p *MockDiscordProvider) Role(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) (*discord.Role, error) {
	return nil, nil
}

func (p *MockDiscordProvider) Member(ctx context.Context, guildID discord.GuildID, userID discord.UserID) (*discord.Member, error) {
	return nil, nil
}

func (p *MockDiscordProvider) Message(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*InteractionResponseResource, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditInteractionResponse(ctx context.Context, applicationID discord.AppID, token string, response api.EditInteractionResponseData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) DeleteInteractionResponse(ctx context.Context, applicationID discord.AppID, token string) error {
	return nil
}

func (p *MockDiscordProvider) CreateInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, data api.InteractionResponseData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID, data api.EditInteractionResponseData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) DeleteInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID) error {
	return nil
}

func (p *MockDiscordProvider) CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, message api.EditMessageData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) CreatePersonaMessage(ctx context.Context, channelID discord.ChannelID, persona Persona, message api.SendMessageData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) CreateMessageReaction(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, emoji discord.APIEmoji) error {
	return nil
}

func (p *MockDiscordProvider) CreatePoll(ctx context.Context, channelID discord.ChannelID, data CreatePollData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) ForwardMessage(ctx context.Context, channelID discord.ChannelID, messageChannelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error {
	return nil
}

func (p *MockDiscordProvider) CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) CreatePrivateChannel(ctx context.Context, userID discord.UserID) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) StartThreadWithMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.StartThreadData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) StartThreadWithoutMessage(ctx context.Context, channelID discord.ChannelID, data api.StartThreadData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) StartThreadInForum(ctx context.Context, channelID discord.ChannelID, data api.StartForumThreadData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) BotToken() string {
	return ""
}

func (p *MockDiscordProvider) APIRequest(ctx context.Context, req DiscordAPIRequest) ([]byte, error) {
	return nil, nil
}

func (p *MockDiscordProvider) UpdateVoiceState(ctx context.Context, guildID discord.GuildID, channelID discord.ChannelID, selfMute bool, selfDeaf bool) error {
	return nil
}

func (p *MockDiscordProvider) PlayAudio(ctx context.Context, opts PlayAudioOpts) error {
	return nil
}

func (p *MockDiscordProvider) SpeakText(ctx context.Context, opts SpeakTextOpts) error {
	return nil
}

func (p *MockDiscordProvider) StopAudio(ctx context.Context, guildID discord.GuildID) error {
	return nil
}

func (p *MockDiscordProvider) UpdatePresence(ctx context.Context, status discord.Status, activity discord.Activity) error {
	return nil
}

func (p *MockDiscordProvider) LeaveGuild(ctx context.Context, guildID discord.GuildID) error {
	return nil
}

func (p *MockDiscordProvider) EditServerProfile(ctx context.Context, guildID discord.GuildID, data ServerProfileData) error {
	return nil
}

func (p *MockDiscordProvider) AppVerifyKey(ctx context.Context) (string, error) {
	return "", nil
}

func (p *MockDiscordProvider) MarkInteractionResponded(interactionID discord.InteractionID) {}

func (p *MockDiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	return false, nil
}

func (p *MockDiscordProvider) AutoDeferInteraction(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) {
}
