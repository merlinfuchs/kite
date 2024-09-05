package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

type DiscordProvider struct {
	flow.MockDiscordProvider // TODO: remove this

	appID    string
	appStore store.AppStore
	session  *state.State

	interactionsWithResponse map[discord.InteractionID]struct{}
}

func NewDiscordProvider(
	appID string,
	appStore store.AppStore,
	session *state.State,
) *DiscordProvider {
	return &DiscordProvider{
		appID:    appID,
		appStore: appStore,
		session:  session,

		interactionsWithResponse: make(map[discord.InteractionID]struct{}),
	}
}

func (p *DiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) error {
	err := p.session.RespondInteraction(interactionID, interactionToken, response)
	if err != nil {
		return fmt.Errorf("failed to respond to interaction: %w", err)
	}

	p.interactionsWithResponse[interactionID] = struct{}{}
	return nil
}

func (p *DiscordProvider) EditInteractionResponse(ctx context.Context, applicationID discord.AppID, token string, response api.EditInteractionResponseData) (*discord.Message, error) {
	msg, err := p.session.EditInteractionResponse(applicationID, token, response)
	if err != nil {
		return nil, fmt.Errorf("failed to edit interaction response: %w", err)
	}

	return msg, err
}

func (p *DiscordProvider) DeleteInteractionResponse(ctx context.Context, applicationID discord.AppID, token string) error {
	err := p.session.DeleteInteractionResponse(applicationID, token)
	if err != nil {
		return fmt.Errorf("failed to delete interaction response: %w", err)
	}

	return nil
}

func (p *DiscordProvider) CreateInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, data api.InteractionResponseData) (*discord.Message, error) {
	msg, err := p.session.FollowUpInteraction(applicationID, token, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create interaction followup: %w", err)
	}

	return msg, nil
}

func (p *DiscordProvider) EditInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID, data api.EditInteractionResponseData) (*discord.Message, error) {
	msg, err := p.session.EditInteractionFollowup(applicationID, messageID, token, data)
	if err != nil {
		return nil, fmt.Errorf("failed to edit interaction followup: %w", err)
	}

	return msg, nil
}

func (p *DiscordProvider) DeleteInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID) error {
	err := p.session.DeleteInteractionFollowup(applicationID, messageID, token)
	if err != nil {
		return fmt.Errorf("failed to delete interaction followup: %w", err)
	}

	return nil
}

func (p *DiscordProvider) CreateMessage(ctx context.Context, channelID discord.ChannelID, message api.SendMessageData) (*discord.Message, error) {
	msg, err := p.session.SendMessageComplex(channelID, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return msg, nil
}

func (p *DiscordProvider) EditMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, message api.EditMessageData) (*discord.Message, error) {
	msg, err := p.session.EditMessageComplex(channelID, messageID, message)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	return msg, nil
}

func (p *DiscordProvider) DeleteMessage(
	ctx context.Context,
	channelID discord.ChannelID,
	messageID discord.MessageID,
	reason api.AuditLogReason,
) error {
	err := p.session.DeleteMessage(channelID, messageID, reason)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (p *DiscordProvider) BanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.BanData) error {
	err := p.session.Ban(guildID, userID, data)
	if err != nil {
		return fmt.Errorf("failed to ban member: %w", err)
	}

	return nil
}

func (p *DiscordProvider) UnbanMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error {
	err := p.session.Unban(guildID, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to unban member: %w", err)
	}

	return nil
}

func (p *DiscordProvider) KickMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, reason api.AuditLogReason) error {
	err := p.session.Kick(guildID, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to kick member: %w", err)
	}

	return nil
}

func (p *DiscordProvider) EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error {
	err := p.session.ModifyMember(guildID, userID, data)
	if err != nil {
		return fmt.Errorf("failed to edit member: %w", err)
	}

	return nil
}

func (p *DiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	_, ok := p.interactionsWithResponse[interactionID]
	return ok, nil
}

type LogProvider struct {
	appID    string
	logStore store.LogStore
}

func NewLogProvider(appID string, logStore store.LogStore) *LogProvider {
	return &LogProvider{
		appID:    appID,
		logStore: logStore,
	}
}

func (p *LogProvider) CreateLogEntry(ctx context.Context, level flow.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := p.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     p.appID,
		Level:     model.LogLevel(level),
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", p.appID).Error("Failed to create log entry")
	}
}

type HTTPProvider struct {
	client *http.Client
}

func NewHTTPProvider(client *http.Client) *HTTPProvider {
	return &HTTPProvider{
		client: client,
	}
}

func (p *HTTPProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return p.client.Do(req)
}

type VariableProvider struct {
	scope              model.VariableValueScope
	variableValueStore store.VariableValueStore
}

func NewVariableProvider(scope model.VariableValueScope, variableValueStore store.VariableValueStore) *VariableProvider {
	return &VariableProvider{
		scope:              scope,
		variableValueStore: variableValueStore,
	}
}

func (p *VariableProvider) SetVariable(ctx context.Context, id string, value flow.FlowValue) error {
	rawValue, err := json.Marshal(value.Value)
	if err != nil {
		return fmt.Errorf("failed to marshal variable value: %w", err)
	}

	err = p.variableValueStore.SetVariableValue(ctx, model.VariableValue{
		VariableID: id,
		Value:      rawValue,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}, p.scope)
	if err != nil {
		return fmt.Errorf("failed to set variable value: %w", err)
	}

	return nil
}

func (p *VariableProvider) Variable(ctx context.Context, id string) (flow.FlowValue, error) {
	row, err := p.variableValueStore.VariableValue(ctx, id, p.scope)
	if err != nil {
		return flow.FlowValue{}, fmt.Errorf("failed to get variable value: %w", err)
	}

	var value flow.FlowValue
	err = json.Unmarshal(row.Value, &value.Value)
	if err != nil {
		return flow.FlowValue{}, fmt.Errorf("failed to unmarshal variable value: %w", err)
	}

	return value, nil
}

func (p *VariableProvider) DeleteVariable(ctx context.Context, id string) error {
	err := p.variableValueStore.DeleteVariableValue(ctx, id, p.scope)
	if err != nil {
		return fmt.Errorf("failed to delete variable value: %w", err)
	}

	return nil
}

type MessageTemplateProvider struct {
	messageStore store.MessageStore
}

func NewMessageTemplateProvider(messageStore store.MessageStore) *MessageTemplateProvider {
	return &MessageTemplateProvider{
		messageStore: messageStore,
	}
}

func (p *MessageTemplateProvider) MessageTemplate(ctx context.Context, id string) (*message.MessageData, error) {
	message, err := p.messageStore.Message(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message.Data, nil
}
