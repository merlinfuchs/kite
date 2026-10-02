package provider

import (
	"context"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
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
	EditMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, message api.EditMessageData) (*discord.Message, error)
	CreateMessageReaction(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, emoji discord.APIEmoji) error
	CreatePoll(ctx context.Context, channelID discord.ChannelID, data CreatePollData) (*discord.Message, error)
	BanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.BanData) error
	UnbanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error
	KickMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error
	PruneMembers(ctx context.Context, guildID discord.GuildID, data api.PruneData) (uint, error)
	EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error
	CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error)
	EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) error
	CreatePrivateChannel(ctx context.Context, userID discord.UserID) (*discord.Channel, error)
	StartThreadWithMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.StartThreadData) (*discord.Channel, error)
	StartThreadWithoutMessage(ctx context.Context, channelID discord.ChannelID, data api.StartThreadData) (*discord.Channel, error)
	APIRequest(ctx context.Context, req DiscordAPIRequest) ([]byte, error)

	UpdateVoiceState(ctx context.Context, guildID discord.GuildID, channelID discord.ChannelID, selfMute bool, selfDeaf bool) error
	UpdatePresence(ctx context.Context, status discord.Status, activity discord.Activity) error

	HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error)
	// MarkInteractionResponded is for interactions a previous execution
	// already responded to, e.g. before a durable sleep.
	MarkInteractionResponded(interactionID discord.InteractionID)
	AutoDeferInteraction(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse)
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

func (p *MockDiscordProvider) CreateMessageReaction(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, emoji discord.APIEmoji) error {
	return nil
}

func (p *MockDiscordProvider) CreatePoll(ctx context.Context, channelID discord.ChannelID, data CreatePollData) (*discord.Message, error) {
	return nil, nil
}

func (p *MockDiscordProvider) BanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.BanData) error {
	return nil
}

func (p *MockDiscordProvider) UnbanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error {
	return nil
}

func (p *MockDiscordProvider) KickMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error {

	return nil
}

func (p *MockDiscordProvider) PruneMembers(ctx context.Context, guildID discord.GuildID, data api.PruneData) (uint, error) {
	return 0, nil
}

func (p *MockDiscordProvider) EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error {
	return nil
}

func (p *MockDiscordProvider) CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) error {
	return nil
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

func (p *MockDiscordProvider) APIRequest(ctx context.Context, req DiscordAPIRequest) ([]byte, error) {
	return nil, nil
}

func (p *MockDiscordProvider) UpdateVoiceState(ctx context.Context, guildID discord.GuildID, channelID discord.ChannelID, selfMute bool, selfDeaf bool) error {
	return nil
}

func (p *MockDiscordProvider) UpdatePresence(ctx context.Context, status discord.Status, activity discord.Activity) error {
	return nil
}

func (p *MockDiscordProvider) MarkInteractionResponded(interactionID discord.InteractionID) {}

func (p *MockDiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	return false, nil
}

func (p *MockDiscordProvider) AutoDeferInteraction(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) {
}
