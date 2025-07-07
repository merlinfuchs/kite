package flow

import (
	"context"
	"net/http"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

type MockDiscordProvider struct{}

func (p *MockDiscordProvider) Guild(ctx context.Context, guildID discord.GuildID) (*discord.Guild, error) {
	return nil, nil
}

func (p *MockDiscordProvider) GuildChannels(ctx context.Context, guildID discord.GuildID) ([]discord.Channel, error) {
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

func (p *MockDiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*FlowInteractionResponseResource, error) {
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

func (p *MockDiscordProvider) DeleteMessage(
	ctx context.Context,
	channelID discord.ChannelID,
	messageID discord.MessageID,
	reason api.AuditLogReason,
) error {
	return nil
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

func (p *MockDiscordProvider) EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error {
	return nil
}

func (p *MockDiscordProvider) AddMemberRole(ctx context.Context, guildID discord.GuildID, userID discord.UserID, roleID discord.RoleID, reason api.AuditLogReason) error {
	return nil
}

func (p *MockDiscordProvider) RemoveMemberRole(ctx context.Context, guildID discord.GuildID, userID discord.UserID, roleID discord.RoleID, reason api.AuditLogReason) error {
	return nil
}

func (p *MockDiscordProvider) CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) (*discord.Channel, error) {
	return nil, nil
}

func (p *MockDiscordProvider) DeleteChannel(ctx context.Context, channelID discord.ChannelID) error {
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

func (p *MockDiscordProvider) CreateRole(ctx context.Context, guildID discord.GuildID, data api.CreateRoleData) (*discord.Role, error) {
	return nil, nil
}

func (p *MockDiscordProvider) EditRole(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID, data api.ModifyRoleData) (*discord.Role, error) {
	return nil, nil
}

func (p *MockDiscordProvider) DeleteRole(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) error {
	return nil
}

func (p *MockDiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	return false, nil
}

func (p *MockDiscordProvider) AutoDeferInteraction(ctx context.Context, interactionID discord.InteractionID, interactionToken string, flags discord.MessageFlags) {
}

type MockLogProvider struct{}

func (p *MockLogProvider) CreateLogEntry(ctx context.Context, level LogLevel, message string) {}

type MockHTTPprovider struct{}

func (p *MockHTTPprovider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, nil
}

type MockAIProvider struct{}

func (p *MockAIProvider) CreateChatCompletion(ctx context.Context, opts CreateChatCompletionOpts) (string, error) {
	return "", nil
}

type MockResumePointProvider struct{}

func (p *MockResumePointProvider) CreateResumePoint(ctx context.Context, _ FlowResumePoint) error {
	return nil
}
