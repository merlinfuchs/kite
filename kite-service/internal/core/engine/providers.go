package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	disstore "github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/responses"
	"github.com/openai/openai-go/v2/shared"
	"gopkg.in/guregu/null.v4"
)

// FeatureProvider resolves the premium features an app has access to.
type FeatureProvider interface {
	AppFeatures(ctx context.Context, appID string) model.Features
	AppFeaturesForApps(ctx context.Context, appIDs []string) (map[string]model.Features, error)
}

type DiscordProvider struct {
	provider.MockDiscordProvider // TODO: remove this

	appID           string
	appStore        store.AppStore
	featureProvider FeatureProvider
	rateLimiter     *BlockRateLimiter
	session         *state.State

	interactionResponseMutex sync.Mutex
	interactionsWithResponse map[discord.InteractionID]struct{}
}

func NewDiscordProvider(
	appID string,
	appStore store.AppStore,
	featureProvider FeatureProvider,
	rateLimiter *BlockRateLimiter,
	session *state.State,
) *DiscordProvider {
	return &DiscordProvider{
		appID:           appID,
		appStore:        appStore,
		featureProvider: featureProvider,
		rateLimiter:     rateLimiter,
		session:         session,

		interactionsWithResponse: make(map[discord.InteractionID]struct{}),
	}
}

func (p *DiscordProvider) Member(ctx context.Context, guildID discord.GuildID, userID discord.UserID) (*discord.Member, error) {
	member, err := p.session.Member(guildID, userID)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusNotFound) || errors.Is(err, disstore.ErrNotFound) {
			return nil, provider.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return member, nil
}

func (p *DiscordProvider) User(ctx context.Context, userID discord.UserID) (*discord.User, error) {
	user, err := p.session.User(userID)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusNotFound) || errors.Is(err, disstore.ErrNotFound) {
			return nil, provider.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (p *DiscordProvider) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	channel, err := p.session.Channel(channelID)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusNotFound) || errors.Is(err, disstore.ErrNotFound) {
			return nil, provider.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get channel: %w", err)
	}

	return channel, nil
}

func (p *DiscordProvider) Role(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) (*discord.Role, error) {
	role, err := p.session.Role(guildID, roleID)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusNotFound) || errors.Is(err, disstore.ErrNotFound) {
			return nil, provider.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

func (p *DiscordProvider) Guild(ctx context.Context, guildID discord.GuildID) (*discord.Guild, error) {
	guild, err := p.session.Guild(guildID)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusNotFound) || errors.Is(err, disstore.ErrNotFound) {
			return nil, provider.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get guild: %w", err)
	}

	return guild, nil
}

func (p *DiscordProvider) Message(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error) {
	msg, err := p.session.Message(channelID, messageID)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusNotFound) || errors.Is(err, disstore.ErrNotFound) {
			return nil, provider.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return msg, nil
}

func (p *DiscordProvider) GuildRoles(ctx context.Context, guildID discord.GuildID) ([]discord.Role, error) {
	roles, err := p.session.Roles(guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return roles, nil
}

func (p *DiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*provider.InteractionResponseResource, error) {
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

	return &provider.InteractionResponseResource{
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

func (p *DiscordProvider) CreateMessageReaction(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, emoji discord.APIEmoji) error {
	err := p.session.React(channelID, messageID, emoji)
	if err != nil {
		return fmt.Errorf("failed to create message reaction: %w", err)
	}

	return nil
}

func (p *DiscordProvider) DeleteMessageReaction(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, emoji discord.APIEmoji) error {
	err := p.session.Unreact(channelID, messageID, emoji)
	if err != nil {
		return fmt.Errorf("failed to delete message reaction: %w", err)
	}

	return nil
}

func (p *DiscordProvider) PinMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, reason api.AuditLogReason) error {
	err := p.session.PinMessage(channelID, messageID, reason)
	if err != nil {
		return fmt.Errorf("failed to pin message: %w", err)
	}

	return nil
}

func (p *DiscordProvider) UnpinMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, reason api.AuditLogReason) error {
	err := p.session.UnpinMessage(channelID, messageID, reason)
	if err != nil {
		return fmt.Errorf("failed to unpin message: %w", err)
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

func (p *DiscordProvider) CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error) {
	channel, err := p.session.CreateChannel(guildID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	return channel, nil
}

func (p *DiscordProvider) EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) error {
	err := p.session.ModifyChannel(channelID, data)
	if err != nil {
		return fmt.Errorf("failed to edit channel: %w", err)
	}

	return nil
}

func (p *DiscordProvider) DeleteChannel(ctx context.Context, channelID discord.ChannelID, reason api.AuditLogReason) error {
	err := p.session.DeleteChannel(channelID, reason)
	if err != nil {
		return fmt.Errorf("failed to delete channel: %w", err)
	}

	return nil
}

func (p *DiscordProvider) CreateInvite(ctx context.Context, channelID discord.ChannelID, data api.CreateInviteData) (*discord.Invite, error) {
	invite, err := p.session.CreateInvite(channelID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	return invite, nil
}

func (p *DiscordProvider) StartThreadWithMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.StartThreadData) (*discord.Channel, error) {
	thread, err := p.session.StartThreadWithMessage(channelID, messageID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to start thread with message: %w", err)
	}

	return thread, nil
}

func (p *DiscordProvider) StartThreadWithoutMessage(ctx context.Context, channelID discord.ChannelID, data api.StartThreadData) (*discord.Channel, error) {
	thread, err := p.session.StartThreadWithoutMessage(channelID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to start thread without message: %w", err)
	}

	return thread, nil
}

func (p *DiscordProvider) AddThreadMember(ctx context.Context, channelID discord.ChannelID, userID discord.UserID) error {
	err := p.session.AddThreadMember(channelID, userID)
	if err != nil {
		return fmt.Errorf("failed to add thread member: %w", err)
	}

	return nil
}

func (p *DiscordProvider) RemoveThreadMember(ctx context.Context, channelID discord.ChannelID, userID discord.UserID) error {
	err := p.session.RemoveThreadMember(channelID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove thread member: %w", err)
	}

	return nil
}

func (p *DiscordProvider) UpdateVoiceState(ctx context.Context, guildID discord.GuildID, channelID discord.ChannelID, selfMute bool, selfDeaf bool) error {
	if err := p.allowGatewayCommand(); err != nil {
		return err
	}

	err := p.session.SendGateway(ctx, &gateway.UpdateVoiceStateCommand{
		GuildID:   guildID,
		ChannelID: channelID,
		SelfMute:  selfMute,
		SelfDeaf:  selfDeaf,
	})
	if err != nil {
		return fmt.Errorf("failed to update voice state: %w", err)
	}

	return nil
}

func (p *DiscordProvider) UpdatePresence(ctx context.Context, status discord.Status, activity discord.Activity) error {
	// Part of the same premium feature as rotating statuses in the app settings
	if !p.featureProvider.AppFeatures(ctx, p.appID).RotatingStatus {
		return fmt.Errorf("setting the status from a flow requires premium")
	}
	if err := p.allowGatewayCommand(); err != nil {
		return err
	}

	err := p.session.SendGateway(ctx, &gateway.UpdatePresenceCommand{
		Status:     status,
		Activities: []discord.Activity{activity},
	})
	if err != nil {
		return fmt.Errorf("failed to update presence: %w", err)
	}

	return nil
}

func (p *DiscordProvider) allowGatewayCommand() error {
	if !p.rateLimiter.Allow(p.appID, gatewayCommandRateLimit) {
		return fmt.Errorf("blocks that change the status or voice state are rate limited, try again in a few seconds")
	}
	return nil
}

func (p *DiscordProvider) MarkInteractionResponded(interactionID discord.InteractionID) {
	p.interactionResponseMutex.Lock()
	defer p.interactionResponseMutex.Unlock()
	p.interactionsWithResponse[interactionID] = struct{}{}
}

func (p *DiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	p.interactionResponseMutex.Lock()
	defer p.interactionResponseMutex.Unlock()

	_, ok := p.interactionsWithResponse[interactionID]
	return ok, nil
}

func (p *DiscordProvider) AutoDeferInteraction(
	ctx context.Context,
	interactionID discord.InteractionID,
	interactionToken string,
	response api.InteractionResponse,
) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(1500 * time.Millisecond):
		hasCreatedResponse, err := p.HasCreatedInteractionResponse(ctx, interactionID)
		if err != nil {
			return
		}

		if !hasCreatedResponse {
			_, err := p.CreateInteractionResponse(ctx, interactionID, interactionToken, response)
			if err != nil {
				slog.Error(
					"Failed to auto-defer interaction",
					slog.String("interaction_id", interactionID.String()),
					slog.String("error", err.Error()),
				)
			}
		}
	}
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

func (p *LogProvider) CreateLogEntry(ctx context.Context, level provider.LogLevel, message string) {
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
		slog.Error(
			"Failed to create log entry",
			slog.String("app_id", p.appID),
			slog.String("error", err.Error()),
		)
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

func (p *AIProvider) CreateResponse(ctx context.Context, opts provider.CreateResponseOpts) (string, error) {
	tools := []responses.ToolUnionParam{}
	for _, tool := range opts.Tools {
		switch tool {
		case provider.AIToolTypeWebSearchPreview:
			tools = append(tools, responses.ToolUnionParam{
				OfWebSearchPreview: &responses.WebSearchPreviewToolParam{
					Type: responses.WebSearchPreviewToolTypeWebSearchPreview,
				},
			})
		}
	}

	inputs := responses.ResponseInputParam{
		{
			OfMessage: &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleUser,
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: openai.String(opts.Prompt),
				},
			},
		},
	}
	if opts.SystemPrompt != "" {
		inputs = append(inputs, responses.ResponseInputItemUnionParam{
			OfMessage: &responses.EasyInputMessageParam{
				Role: responses.EasyInputMessageRoleSystem,
				Content: responses.EasyInputMessageContentUnionParam{
					OfString: openai.String(opts.SystemPrompt),
				},
			},
		})
	}

	model := opts.Model
	if model == "" {
		model = openai.ChatModelGPT4oMini
	}

	maxOutputTokens := 500
	if opts.MaxOutputTokens > 0 && opts.MaxOutputTokens < maxOutputTokens {
		maxOutputTokens = opts.MaxOutputTokens
	}

	params := responses.ResponseNewParams{
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: inputs,
		},
		MaxOutputTokens: openai.Int(int64(maxOutputTokens)),
		Tools:           tools,
	}
	// GPT-5 models reason before answering, and the reasoning counts towards
	// the output tokens, so less of it leaves more room for the answer.
	if strings.HasPrefix(model, "gpt-5") {
		params.Reasoning = shared.ReasoningParam{Effort: shared.ReasoningEffortLow}
	}

	resp, err := p.client.Responses.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to create response: %w", err)
	}

	return resp.OutputText(), nil
}

// Variable IDs come from user-authored flow data, so lookups are scoped to the app.
type VariableProvider struct {
	appID              string
	variableValueStore store.VariableValueStore
}

func NewVariableProvider(appID string, variableValueStore store.VariableValueStore) *VariableProvider {
	return &VariableProvider{
		appID:              appID,
		variableValueStore: variableValueStore,
	}
}

func (p *VariableProvider) UpdateVariable(ctx context.Context, id string, scope null.String, operation provider.VariableOperation, value thing.Thing) (thing.Thing, error) {
	v := model.VariableValue{
		VariableID: id,
		Scope:      scope,
		Data:       value,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	newValue, err := p.variableValueStore.UpdateVariableValue(ctx, p.appID, operation, v)
	if err != nil {
		return thing.Null, fmt.Errorf("failed to %s variable value: %w", operation, err)
	}

	return newValue.Data, nil
}

func (p *VariableProvider) Variable(ctx context.Context, id string, scope null.String) (thing.Thing, error) {
	row, err := p.variableValueStore.VariableValue(ctx, p.appID, id, scope)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return thing.Null, provider.ErrNotFound
		}
		return thing.Null, fmt.Errorf("failed to get variable value: %w", err)
	}

	return row.Data, nil
}

func (p *VariableProvider) DeleteVariable(ctx context.Context, id string, scope null.String) error {
	err := p.variableValueStore.DeleteVariableValue(ctx, p.appID, id, scope)
	if err != nil {
		return fmt.Errorf("failed to delete variable value: %w", err)
	}

	return nil
}

type MessageTemplateProvider struct {
	// Template IDs come from user-authored flow data, so lookups are scoped to
	// the app to keep a flow from using another app's templates.
	appID                string
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
}

func NewMessageTemplateProvider(appID string, messageStore store.MessageStore, messageInstanceStore store.MessageInstanceStore) *MessageTemplateProvider {
	return &MessageTemplateProvider{
		appID:                appID,
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
	}
}

func (p *MessageTemplateProvider) MessageTemplate(ctx context.Context, id string) (*message.MessageData, error) {
	message, err := p.messageStore.Message(ctx, p.appID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return &message.Data, nil
}

func (p *MessageTemplateProvider) LinkMessageTemplateInstance(ctx context.Context, instance provider.MessageTemplateInstance) error {
	message, err := p.messageStore.Message(ctx, p.appID, instance.MessageTemplateID)
	if err != nil {
		return fmt.Errorf("failed to get message: %w", err)
	}

	_, err = p.messageInstanceStore.CreateMessageInstance(ctx, p.appID, &model.MessageInstance{
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

// maxPendingTimers bounds the durable sleeps an app can have waiting, a busy
// listener could otherwise create one for every message.
const maxPendingTimers = 1000

// timerExpiry is how long after resume_at a timer that couldn't resume, e.g.
// because the app's gateway was down, is kept before it's deleted.
const timerExpiry = time.Hour

type ResumePointProvider struct {
	resumePointStore store.ResumePointStore
	tokenCrypt       *util.SymmetricCrypt

	appID string
	links entityLinks
}

func NewResumePointProvider(
	resumePointStore store.ResumePointStore,
	tokenCrypt *util.SymmetricCrypt,
	appID string,
	links entityLinks,
) *ResumePointProvider {
	return &ResumePointProvider{
		resumePointStore: resumePointStore,
		tokenCrypt:       tokenCrypt,
		appID:            appID,
		links:            links,
	}
}

// timerFields checks the app's limit of pending timers and prepares the
// columns only timers have.
func (p *ResumePointProvider) timerFields(ctx context.Context, s flow.ResumePoint) (resumeAt, expiresAt null.Time, interactionToken null.String, err error) {
	pending, err := p.resumePointStore.CountPendingTimerResumePoints(ctx, p.appID)
	if err != nil {
		return resumeAt, expiresAt, interactionToken, fmt.Errorf("failed to count pending timers: %w", err)
	}
	if pending >= maxPendingTimers {
		return resumeAt, expiresAt, interactionToken, fmt.Errorf("too many sleeping flows, at most %d can wait at the same time", maxPendingTimers)
	}

	if s.InteractionToken != "" {
		token, err := p.tokenCrypt.EncryptString(s.InteractionToken)
		if err != nil {
			return resumeAt, expiresAt, interactionToken, fmt.Errorf("failed to encrypt interaction token: %w", err)
		}
		interactionToken = null.StringFrom(token)
	}

	return null.TimeFrom(s.ResumeAt), null.TimeFrom(s.ResumeAt.Add(timerExpiry)), interactionToken, nil
}

func (p *ResumePointProvider) CreateResumePoint(ctx context.Context, s flow.ResumePoint) (flow.ResumePoint, error) {
	if s.ID == "" {
		s.ID = util.UniqueID()
	}

	var expiresAt, resumeAt null.Time
	var interactionToken null.String
	switch s.Type {
	case flow.ResumePointTypeModal:
		expiresAt = null.NewTime(time.Now().UTC().Add(time.Hour*1), true)
	case flow.ResumePointTypeTimer:
		var err error
		resumeAt, expiresAt, interactionToken, err = p.timerFields(ctx, s)
		if err != nil {
			return s, err
		}
	}

	// TODO: Implement some kind of expiration for other resume point types
	// Maybe based on last usage?

	err := p.resumePointStore.CreateResumePoint(ctx, &model.ResumePoint{
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
		ResumeAt:          resumeAt,
		InteractionToken:  interactionToken,
	})

	return s, err
}

type ValueProvider struct {
	pluginInstanceID string
	pluginValueStore store.PluginValueStore
}

func NewValueProvider(pluginInstanceID string, pluginValueStore store.PluginValueStore) *ValueProvider {
	return &ValueProvider{
		pluginInstanceID: pluginInstanceID,
		pluginValueStore: pluginValueStore,
	}
}

func (p *ValueProvider) UpdateValue(ctx context.Context, key string, op provider.VariableOperation, value thing.Thing) (thing.Thing, error) {
	v := model.PluginValue{
		PluginInstanceID: p.pluginInstanceID,
		Key:              key,
		Value:            value,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	newValue, err := p.pluginValueStore.UpdatePluginValue(ctx, op, v)
	if err != nil {
		return thing.Null, fmt.Errorf("failed to %s plugin value: %w", op, err)
	}

	return newValue.Value, nil
}

func (p *ValueProvider) GetValue(ctx context.Context, key string) (thing.Thing, error) {
	v, err := p.pluginValueStore.GetPluginValue(ctx, p.pluginInstanceID, key)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return thing.Null, nil
		}
		return thing.Null, fmt.Errorf("failed to get plugin value: %w", err)
	}

	return v.Value, nil
}

func (p *ValueProvider) DeleteValue(ctx context.Context, key string) error {
	err := p.pluginValueStore.DeletePluginValue(ctx, p.pluginInstanceID, key)
	if err != nil {
		return fmt.Errorf("failed to delete plugin value: %w", err)
	}

	return nil
}

type RobloxProvider struct {
	client *http.Client
}

func NewRobloxProvider(client *http.Client) *RobloxProvider {
	return &RobloxProvider{
		client: client,
	}
}

func (p *RobloxProvider) UserByID(ctx context.Context, id int64) (*thing.RobloxUserValue, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://users.roblox.com/v1/users/%d", id), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create roblox user request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get roblox user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// User-not-found is a normal outcome here. Drain first: closing an
		// unread body makes net/http drop the connection instead of pooling it,
		// so every miss would otherwise cost a fresh TCP + TLS handshake.
		io.Copy(io.Discard, io.LimitReader(resp.Body, 4*1024))
		return nil, provider.ErrNotFound
	}

	var v thing.RobloxUserValue
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode roblox user: %w", err)
	}

	return &v, nil
}

func (p *RobloxProvider) UsersByUsername(ctx context.Context, username string) ([]thing.RobloxUserValue, error) {
	data := struct {
		Usernames []string `json:"usernames"`
	}{
		Usernames: []string{username},
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal roblox users request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://users.roblox.com/v1/usernames/users", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create roblox users request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get roblox users: %w", err)
	}
	defer resp.Body.Close()

	var v struct {
		Data []thing.RobloxUserValue `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode roblox users: %w", err)
	}

	return v.Data, nil
}
