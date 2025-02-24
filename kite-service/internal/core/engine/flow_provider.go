package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/sashabaranov/go-openai"
	"gopkg.in/guregu/null.v4"
)

type DiscordProvider struct {
	flow.MockDiscordProvider // TODO: remove this

	appID    string
	appStore store.AppStore
	session  *state.State

	interactionResponseMutex sync.Mutex
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

func (p *DiscordProvider) GuildRoles(ctx context.Context, guildID discord.GuildID) ([]discord.Role, error) {
	roles, err := p.session.Roles(guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return roles, nil
}

func (p *DiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*flow.FlowInteractionResponseResource, error) {
	p.interactionResponseMutex.Lock()
	defer p.interactionResponseMutex.Unlock()

	endpoint := api.EndpointInteractions + interactionID.String() + "/" + interactionToken + "/callback?with_response=true"

	var res struct {
		Resource struct {
			Type    api.InteractionResponseType `json:"type"`
			Message *discord.Message            `json:"message,omitempty"`
		} `json:"resource"`
	}
	if err := sendpart.POST(p.session.Client.Client, response, &res, endpoint); err != nil {
		return nil, fmt.Errorf("failed to respond to interaction: %w", err)
	}

	p.interactionsWithResponse[interactionID] = struct{}{}

	return &flow.FlowInteractionResponseResource{
		Type:    res.Resource.Type,
		Message: res.Resource.Message,
	}, nil
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

func (p *DiscordProvider) AddMemberRole(ctx context.Context, guildID discord.GuildID, userID discord.UserID, roleID discord.RoleID, reason api.AuditLogReason) error {
	err := p.session.AddRole(guildID, userID, roleID, api.AddRoleData{
		AuditLogReason: reason,
	})
	if err != nil {
		return fmt.Errorf("failed to add role: %w", err)
	}

	return nil
}

func (p *DiscordProvider) RemoveMemberRole(ctx context.Context, guildID discord.GuildID, userID discord.UserID, roleID discord.RoleID, reason api.AuditLogReason) error {
	err := p.session.RemoveRole(guildID, userID, roleID, reason)
	if err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	return nil
}

func (p *DiscordProvider) CreatePrivateChannel(ctx context.Context, userID discord.UserID) (*discord.Channel, error) {
	channel, err := p.session.CreatePrivateChannel(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create DM channel: %w", err)
	}

	return channel, nil
}

func (p *DiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	p.interactionResponseMutex.Lock()
	defer p.interactionResponseMutex.Unlock()

	_, ok := p.interactionsWithResponse[interactionID]
	return ok, nil
}

type LogProvider struct {
	appID    string
	logStore store.LogStore

	links entityLinks
}

func NewLogProvider(appID string, logStore store.LogStore, links entityLinks) *LogProvider {
	return &LogProvider{
		appID:    appID,
		logStore: logStore,
		links:    links,
	}
}

func (p *LogProvider) CreateLogEntry(ctx context.Context, level flow.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	err := p.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:           p.appID,
		Level:           model.LogLevel(level),
		Message:         message,
		CommandID:       p.links.CommandID,
		EventListenerID: p.links.EventListenerID,
		MessageID:       p.links.MessageID,
		CreatedAt:       time.Now().UTC(),
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

type AIProvider struct {
	client *openai.Client
}

func NewAIProvider(client *openai.Client) *AIProvider {
	return &AIProvider{
		client: client,
	}
}

func (p *AIProvider) CreateChatCompletion(ctx context.Context, opts flow.CreateChatCompletionOpts) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: opts.Prompt},
	}

	if opts.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: opts.SystemPrompt,
		})
	}

	maxCompletionTokens := 2000
	if opts.MaxCompletionTokens > 0 && opts.MaxCompletionTokens < maxCompletionTokens {
		maxCompletionTokens = opts.MaxCompletionTokens
	}

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:               openai.GPT4oMini,
		Messages:            messages,
		MaxCompletionTokens: maxCompletionTokens,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

type VariableProvider struct {
	variableValueStore store.VariableValueStore
}

func NewVariableProvider(variableValueStore store.VariableValueStore) *VariableProvider {
	return &VariableProvider{
		variableValueStore: variableValueStore,
	}
}

func (p *VariableProvider) UpdateVariable(ctx context.Context, id string, scope null.String, operation flow.VariableOperation, value thing.Any) (thing.Any, error) {
	v := model.VariableValue{
		VariableID: id,
		Scope:      scope,
		Data:       value,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	newValue, err := p.variableValueStore.UpdateVariableValue(ctx, operation, v)
	if err != nil {
		return thing.Null, fmt.Errorf("failed to %s variable value: %w", operation, err)
	}

	return newValue.Data, nil
}

func (p *VariableProvider) Variable(ctx context.Context, id string, scope null.String) (thing.Any, error) {
	row, err := p.variableValueStore.VariableValue(ctx, id, scope)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return thing.Null, flow.ErrNotFound
		}
		return thing.Null, fmt.Errorf("failed to get variable value: %w", err)
	}

	return row.Data, nil
}

func (p *VariableProvider) DeleteVariable(ctx context.Context, id string, scope null.String) error {
	err := p.variableValueStore.DeleteVariableValue(ctx, id, scope)
	if err != nil {
		return fmt.Errorf("failed to delete variable value: %w", err)
	}

	return nil
}

type MessageTemplateProvider struct {
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
}

func NewMessageTemplateProvider(messageStore store.MessageStore, messageInstanceStore store.MessageInstanceStore) *MessageTemplateProvider {
	return &MessageTemplateProvider{
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
	}
}

func (p *MessageTemplateProvider) MessageTemplate(ctx context.Context, id string) (*message.MessageData, error) {
	message, err := p.messageStore.Message(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message.Data, nil
}

func (p *MessageTemplateProvider) LinkMessageTemplateInstance(ctx context.Context, instance flow.FlowMessageTemplateInstance) error {
	message, err := p.messageStore.Message(ctx, instance.MessageTemplateID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	_, err = p.messageInstanceStore.CreateMessageInstance(ctx, &model.MessageInstance{
		MessageID:        message.ID,
		DiscordMessageID: instance.MessageID.String(),
		DiscordChannelID: instance.ChannelID.String(),
		DiscordGuildID:   instance.GuildID.String(),
		Ephemeral:        instance.Ephemeral,
		Hidden:           true,
		FlowSources:      message.FlowSources,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to create message instance: %w", err)
	}

	return nil
}

type ResumePointProvider struct {
	resumePointStore store.ResumePointStore

	appID string
	links entityLinks
}

func NewResumePointProvider(
	resumePointStore store.ResumePointStore,
	appID string,
	links entityLinks,
) *ResumePointProvider {
	return &ResumePointProvider{
		resumePointStore: resumePointStore,
		appID:            appID,
		links:            links,
	}
}

func (p *ResumePointProvider) CreateResumePoint(ctx context.Context, s flow.FlowResumePoint) (flow.FlowResumePoint, error) {
	s.ID = util.UniqueID()

	var expiresAt null.Time
	if s.Type == flow.FlowResumePointTypeModal {
		expiresAt = null.NewTime(time.Now().UTC().Add(time.Hour*1), true)
	}

	_, err := p.resumePointStore.CreateResumePoint(ctx, &model.ResumePoint{
		ID:                s.ID,
		Type:              model.ResumePointType(s.Type),
		AppID:             p.appID,
		CommandID:         p.links.CommandID,
		EventListenerID:   p.links.EventListenerID,
		MessageID:         p.links.MessageID,
		MessageInstanceID: p.links.MessageInstanceID,
		FlowSourceID:      p.links.FlowSourceID,
		FlowNodeID:        s.NodeID,
		FlowState:         s.State,
		CreatedAt:         time.Now().UTC(),
		ExpiresAt:         expiresAt,
	})

	return s, err
}
