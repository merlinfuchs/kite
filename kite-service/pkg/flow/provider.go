package flow

import (
	"context"
	"errors"
	"net/http"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

var (
	ErrNotFound = errors.New("not found")
)

type FlowProviders struct {
	Discord         FlowDiscordProvider
	HTTP            FlowHTTPProvider
	Log             FlowLogProvider
	Variable        FlowVariableProvider
	MessageTemplate FlowMessageTemplateProvider
}

type FlowDiscordProvider interface {
	Guild(ctx context.Context, guildID discord.GuildID) (*discord.Guild, error)
	GuildChannels(ctx context.Context, guildID discord.GuildID) ([]discord.Channel, error)
	GuildRoles(ctx context.Context, guildID discord.GuildID) ([]discord.Role, error)
	Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error)
	User(ctx context.Context, userID discord.UserID) (*discord.User, error)
	Role(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) (*discord.Role, error)
	Member(ctx context.Context, guildID discord.GuildID, userID discord.UserID) (*discord.Member, error)

	CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*FlowInteractionResponseResource, error)
	EditInteractionResponse(ctx context.Context, applicationID discord.AppID, token string, response api.EditInteractionResponseData) (*discord.Message, error)
	DeleteInteractionResponse(ctx context.Context, applicationID discord.AppID, token string) error
	CreateInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, data api.InteractionResponseData) (*discord.Message, error)
	EditInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID, data api.EditInteractionResponseData) (*discord.Message, error)
	DeleteInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID) error
	CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error)
	EditMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, message api.EditMessageData) (*discord.Message, error)
	DeleteMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, reason api.AuditLogReason) error
	BanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.BanData) error
	UnbanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error
	KickMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error
	EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error
	CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error)
	EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) (*discord.Channel, error)
	DeleteChannel(ctx context.Context, channelID discord.ChannelID) error
	StartThreadWithMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.StartThreadData) (*discord.Channel, error)
	StartThreadWithoutMessage(ctx context.Context, channelID discord.ChannelID, data api.StartThreadData) (*discord.Channel, error)
	CreateRole(ctx context.Context, guildID discord.GuildID, data api.CreateRoleData) (*discord.Role, error)
	EditRole(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID, data api.ModifyRoleData) (*discord.Role, error)
	DeleteRole(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) error

	HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error)
}

type FlowInteractionResponseResource struct {
	Type    api.InteractionResponseType
	Message *discord.Message
}

type FlowHTTPProvider interface {
	HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error)
}

type FlowLogProvider interface {
	CreateLogEntry(ctx context.Context, level LogLevel, message string)
}

type FlowVariableProvider interface {
	SetVariable(ctx context.Context, id string, value FlowValue) error
	Variable(ctx context.Context, id string) (FlowValue, error)
	DeleteVariable(ctx context.Context, id string) error
}

type FlowMessageTemplateProvider interface {
	MessageTemplate(ctx context.Context, id string) (*message.MessageData, error)
	LinkMessageTemplateInstance(ctx context.Context, instance FlowMessageTemplateInstance) error
}

type FlowMessageTemplateInstance struct {
	MessageTemplateID string
	MessageID         discord.MessageID
	ChannelID         discord.ChannelID
	GuildID           discord.GuildID
	Ephemeral         bool
}
