package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

// errNotInTestRuns is returned by providers of services test runs don't call.
// Blocks that use them don't run in test runs, so it's only a safeguard.
var errNotInTestRuns = errors.New("not available in test runs")

// testProviders returns the providers of one execution of a test run. Reads
// from Discord and the app's data are real, everything else is recorded in run
// instead of done.
func (s *testSession) testProviders(run *flow.TestRun) flow.FlowProviders {
	real := s.real

	return flow.FlowProviders{
		Discord:         &testDiscordProvider{real: real.Discord, s: s, run: run},
		Roblox:          testRobloxProvider{},
		Log:             testLogProvider{run: run},
		HTTP:            testHTTPProvider{},
		AI:              testAIProvider{},
		Decision:        testDecisionProvider{run: run},
		MessageTemplate: &testMessageTemplateProvider{real: real.MessageTemplate, s: s},
		Variable:        &testVariableProvider{real: real.Variable, s: s, run: run},
		Cooldown:        s.cooldowns,
		ResumePoint:     &testResumePointProvider{s: s},
		Secret:          testSecretProvider{},
		Integration:     testIntegrationProvider{real: real.Integration},
	}
}

type testDiscordProvider struct {
	real provider.DiscordProvider
	s    *testSession
	run  *flow.TestRun
}

func (p *testDiscordProvider) record(action string, data any, warnings ...string) {
	p.run.AddEffect(flow.TestEffect{Action: action, Data: data, Warnings: warnings})
}

func (p *testDiscordProvider) Guild(ctx context.Context, guildID discord.GuildID) (*discord.Guild, error) {
	return p.real.Guild(ctx, guildID)
}

func (p *testDiscordProvider) GuildRoles(ctx context.Context, guildID discord.GuildID) ([]discord.Role, error) {
	return p.real.GuildRoles(ctx, guildID)
}

func (p *testDiscordProvider) Channel(ctx context.Context, channelID discord.ChannelID) (*discord.Channel, error) {
	if channel, ok := p.s.fakeChannel(channelID); ok {
		return &channel, nil
	}
	return p.real.Channel(ctx, channelID)
}

func (p *testDiscordProvider) User(ctx context.Context, userID discord.UserID) (*discord.User, error) {
	return p.real.User(ctx, userID)
}

func (p *testDiscordProvider) Role(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) (*discord.Role, error) {
	return p.real.Role(ctx, guildID, roleID)
}

func (p *testDiscordProvider) Member(ctx context.Context, guildID discord.GuildID, userID discord.UserID) (*discord.Member, error) {
	return p.real.Member(ctx, guildID, userID)
}

func (p *testDiscordProvider) Message(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error) {
	if msg, ok := p.s.fakeMessage(messageID); ok {
		return &msg, nil
	}
	return p.real.Message(ctx, channelID, messageID)
}

func (p *testDiscordProvider) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*provider.InteractionResponseResource, error) {
	if !p.s.markResponded(interactionID) {
		return nil, testDiscordError(http.StatusBadRequest, 40060, "Interaction has already been acknowledged.")
	}

	interaction := p.s.currentInteraction()
	res := &provider.InteractionResponseResource{Type: response.Type}

	switch response.Type {
	case api.MessageInteractionWithSource:
		msg := p.s.addMessage(p.responseMessage(interaction, response.Data))
		p.s.setOriginal(interactionID, msg.ID)
		res.Message = &msg
		p.record("respond", map[string]any{"type": "message", "message": response.Data})
	case api.UpdateMessage:
		if interaction != nil && interaction.Message != nil {
			msg := p.s.editMessage(*interaction.Message, response.Data.Content, response.Data.Embeds, response.Data.Components)
			res.Message = &msg
		}
		p.record("respond", map[string]any{"type": "update_message", "message": response.Data})
	case api.ModalResponse:
		p.s.setModal(response.Data)
		p.record("respond", map[string]any{"type": "modal", "modal": response.Data})
	case api.DeferredMessageInteractionWithSource:
		ephemeral := response.Data != nil && response.Data.Flags&discord.EphemeralMessage != 0
		p.record("respond", map[string]any{"type": "defer", "ephemeral": ephemeral})
	case api.DeferredMessageUpdate:
		p.record("respond", map[string]any{"type": "defer_update"})
	default:
		p.record("respond", map[string]any{"type": response.Type, "data": response.Data})
	}

	return res, nil
}

func (p *testDiscordProvider) EditInteractionResponse(ctx context.Context, applicationID discord.AppID, token string, response api.EditInteractionResponseData) (*discord.Message, error) {
	p.record("edit_response", map[string]any{"message": response})
	msg := p.s.editOriginal(response.Content, response.Embeds, response.Components)
	return &msg, nil
}

func (p *testDiscordProvider) DeleteInteractionResponse(ctx context.Context, applicationID discord.AppID, token string) error {
	p.record("delete_response", nil)
	p.s.deleteOriginal()
	return nil
}

func (p *testDiscordProvider) CreateInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, data api.InteractionResponseData) (*discord.Message, error) {
	p.record("create_followup", map[string]any{"message": data})
	msg := p.s.addMessage(p.responseMessage(p.s.currentInteraction(), &data))
	return &msg, nil
}

func (p *testDiscordProvider) EditInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID, data api.EditInteractionResponseData) (*discord.Message, error) {
	p.record("edit_followup", map[string]any{"message_id": messageID, "message": data})
	msg, ok := p.s.fakeMessage(messageID)
	if !ok {
		msg = discord.Message{ID: messageID}
	}
	msg = p.s.editMessage(msg, data.Content, data.Embeds, data.Components)
	return &msg, nil
}

func (p *testDiscordProvider) DeleteInteractionFollowup(ctx context.Context, applicationID discord.AppID, token string, messageID discord.MessageID) error {
	p.record("delete_followup", map[string]any{"message_id": messageID})
	p.s.deleteMessage(messageID)
	return nil
}

func (p *testDiscordProvider) CreateMessage(ctx context.Context, channelID discord.ChannelID, data api.SendMessageData) (*discord.Message, error) {
	effect := map[string]any{"channel_id": channelID, "message": data}
	var warnings []string
	if channel, ok := p.s.fakeChannel(channelID); ok && channel.Type == discord.DirectMessage {
		if len(channel.DMRecipients) > 0 {
			effect["dm_user_id"] = channel.DMRecipients[0].ID
		}
	} else {
		perms := discord.PermissionViewChannel | discord.PermissionSendMessages
		if len(data.Embeds) > 0 {
			perms |= discord.PermissionEmbedLinks
		}
		if data.Reference != nil {
			perms |= discord.PermissionReadMessageHistory
		}
		warnings = p.s.checkChannel(channelID, perms)
	}
	p.record("create_message", effect, warnings...)

	msg := p.s.addMessage(discord.Message{
		ChannelID:  channelID,
		GuildID:    p.s.guildOfChannel(channelID),
		Content:    data.Content,
		Embeds:     data.Embeds,
		Components: data.Components,
		Flags:      data.Flags,
	})
	return &msg, nil
}

func (p *testDiscordProvider) EditMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.EditMessageData) (*discord.Message, error) {
	p.record("edit_message", map[string]any{"channel_id": channelID, "message_id": messageID, "message": data}, p.s.checkOwnMessage(channelID, messageID)...)

	msg, ok := p.s.fakeMessage(messageID)
	if !ok {
		msg = discord.Message{ID: messageID, ChannelID: channelID}
	}
	msg = p.s.editMessage(msg, data.Content, data.Embeds, data.Components)
	return &msg, nil
}

func (p *testDiscordProvider) CreateMessageReaction(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, emoji discord.APIEmoji) error {
	warnings := p.s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionReadMessageHistory|discord.PermissionAddReactions)
	p.record("add_reaction", map[string]any{"channel_id": channelID, "message_id": messageID, "emoji": emoji}, warnings...)
	return nil
}

func (p *testDiscordProvider) CreatePoll(ctx context.Context, channelID discord.ChannelID, data provider.CreatePollData) (*discord.Message, error) {
	warnings := p.s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionSendMessages)
	p.record("create_poll", map[string]any{"channel_id": channelID, "poll": data}, warnings...)
	msg := p.s.addMessage(discord.Message{ChannelID: channelID, GuildID: p.s.guildOfChannel(channelID)})
	return &msg, nil
}

func (p *testDiscordProvider) ForwardMessage(ctx context.Context, channelID discord.ChannelID, messageChannelID discord.ChannelID, messageID discord.MessageID) (*discord.Message, error) {
	warnings := p.s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionSendMessages)
	warnings = append(warnings, p.s.checkChannel(messageChannelID, discord.PermissionViewChannel|discord.PermissionReadMessageHistory)...)
	p.record("forward_message", map[string]any{"channel_id": channelID, "message_channel_id": messageChannelID, "message_id": messageID}, warnings...)
	msg := p.s.addMessage(discord.Message{ChannelID: channelID, GuildID: p.s.guildOfChannel(channelID)})
	return &msg, nil
}

func (p *testDiscordProvider) EditMember(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) error {
	p.record("edit_member", map[string]any{"guild_id": guildID, "user_id": userID, "changes": data}, p.s.checkMemberEdit(guildID, userID, data)...)
	return nil
}

func (p *testDiscordProvider) CreateChannel(ctx context.Context, guildID discord.GuildID, data api.CreateChannelData) (*discord.Channel, error) {
	p.record("create_channel", map[string]any{"guild_id": guildID, "channel": data}, p.s.checkGuild(guildID, discord.PermissionManageChannels)...)
	channel := p.s.addChannel(discord.Channel{
		GuildID:  guildID,
		Name:     data.Name,
		Type:     data.Type,
		Topic:    data.Topic,
		ParentID: data.CategoryID,
	})
	return &channel, nil
}

func (p *testDiscordProvider) EditChannel(ctx context.Context, channelID discord.ChannelID, data api.ModifyChannelData) error {
	p.record("edit_channel", map[string]any{"channel_id": channelID, "changes": data}, p.s.checkChannelManage(channelID)...)
	return nil
}

func (p *testDiscordProvider) CreatePrivateChannel(ctx context.Context, userID discord.UserID) (*discord.Channel, error) {
	// Opening a DM changes nothing, so it isn't recorded. Messages to it show
	// who they'd go to.
	channel := p.s.addChannel(discord.Channel{
		Type:         discord.DirectMessage,
		DMRecipients: []discord.User{{ID: userID}},
	})
	return &channel, nil
}

func (p *testDiscordProvider) StartThreadWithMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID, data api.StartThreadData) (*discord.Channel, error) {
	warnings := p.s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionCreatePublicThreads)
	p.record("create_thread", map[string]any{"channel_id": channelID, "message_id": messageID, "thread": data}, warnings...)
	return p.thread(channelID, data.Name, discord.GuildPublicThread), nil
}

func (p *testDiscordProvider) StartThreadWithoutMessage(ctx context.Context, channelID discord.ChannelID, data api.StartThreadData) (*discord.Channel, error) {
	perm := discord.PermissionCreatePublicThreads
	if data.Type == discord.GuildPrivateThread {
		perm = discord.PermissionCreatePrivateThreads
	}
	warnings := p.s.checkChannel(channelID, discord.PermissionViewChannel|perm)
	p.record("create_thread", map[string]any{"channel_id": channelID, "thread": data}, warnings...)
	return p.thread(channelID, data.Name, data.Type), nil
}

func (p *testDiscordProvider) StartThreadInForum(ctx context.Context, channelID discord.ChannelID, data api.StartForumThreadData) (*discord.Channel, error) {
	warnings := p.s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionSendMessages)
	p.record("create_forum_post", map[string]any{"channel_id": channelID, "post": data}, warnings...)
	return p.thread(channelID, data.Name, discord.GuildPublicThread), nil
}

func (p *testDiscordProvider) thread(parentID discord.ChannelID, name string, t discord.ChannelType) *discord.Channel {
	if t == 0 {
		t = discord.GuildPublicThread
	}
	channel := p.s.addChannel(discord.Channel{
		GuildID:  p.s.guildOfChannel(parentID),
		ParentID: parentID,
		Name:     name,
		Type:     t,
	})
	return &channel
}

// APIRequest sends reads to Discord and records everything else. Requests
// with the interaction's token aren't sent either, test runs have none. Nor
// are reads of endpoints only flows can use, like the webhooks of a server
// with their tokens, which the token of a test may not read.
func (p *testDiscordProvider) APIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error) {
	usesToken := strings.HasPrefix(req.Path, "/webhooks/") || strings.HasPrefix(req.Path, "/interactions/")
	if req.Method == http.MethodGet && !usesToken && !flow.DiscordAPIFlowOnly(req.Method, req.Path) {
		return p.real.APIRequest(ctx, req)
	}

	var body any
	if len(req.Body) > 0 {
		body = json.RawMessage(req.Body)
	}
	warnings := p.s.checkAPIRequest(req)
	if req.Method == http.MethodGet {
		warnings = append(warnings, "Not sent, test runs don't read this endpoint. The result is empty.")
	}
	p.record("discord_api_request", map[string]any{"method": req.Method, "path": req.Path, "body": body}, warnings...)

	if req.Method == http.MethodDelete {
		if id, ok := deletedMessageID(req.Path); ok {
			p.s.deleteMessage(id)
		}
		return nil, nil
	}
	return []byte(fmt.Sprintf(`{"id":%q}`, flow.TestSnowflake().String())), nil
}

func (p *testDiscordProvider) BotToken() string {
	return "<bot token>"
}

func (p *testDiscordProvider) UpdateVoiceState(ctx context.Context, guildID discord.GuildID, channelID discord.ChannelID, selfMute bool, selfDeaf bool) error {
	var warnings []string
	if channelID.IsValid() {
		warnings = p.s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionConnect)
	}
	p.record("update_voice_state", map[string]any{"guild_id": guildID, "channel_id": channelID, "self_mute": selfMute, "self_deaf": selfDeaf}, warnings...)
	return nil
}

func (p *testDiscordProvider) UpdatePresence(ctx context.Context, status discord.Status, activity discord.Activity) error {
	p.record("update_status", map[string]any{"status": status, "activity": activity})
	return nil
}

func (p *testDiscordProvider) LeaveGuild(ctx context.Context, guildID discord.GuildID) error {
	p.record("leave_server", map[string]any{"guild_id": guildID})
	return nil
}

func (p *testDiscordProvider) AppVerifyKey(ctx context.Context) (string, error) {
	return p.real.AppVerifyKey(ctx)
}

func (p *testDiscordProvider) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	return p.s.responded(interactionID), nil
}

func (p *testDiscordProvider) MarkInteractionResponded(interactionID discord.InteractionID) {
	p.s.markResponded(interactionID)
}

// AutoDeferInteraction does nothing, test runs don't take long enough for it.
func (p *testDiscordProvider) AutoDeferInteraction(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) {
}

func (p *testDiscordProvider) responseMessage(interaction *discord.InteractionEvent, data *api.InteractionResponseData) discord.Message {
	var msg discord.Message
	if interaction != nil {
		msg.ChannelID = interaction.ChannelID
		msg.GuildID = interaction.GuildID
	}
	if data != nil {
		msg.Content = data.Content.Val
		if data.Embeds != nil {
			msg.Embeds = *data.Embeds
		}
		if data.Components != nil {
			msg.Components = *data.Components
		}
		msg.Flags = data.Flags
	}
	return msg
}

var deletedMessagePath = regexp.MustCompile(`^/channels/\d+/messages/(\d+)$`)

// deletedMessageID returns the message a DELETE request to the path deletes.
func deletedMessageID(path string) (discord.MessageID, bool) {
	match := deletedMessagePath.FindStringSubmatch(path)
	if match == nil {
		return 0, false
	}
	id, err := discord.ParseSnowflake(match[1])
	return discord.MessageID(id), err == nil
}

func testDiscordError(status int, code int, msg string) error {
	return &httputil.HTTPError{Status: status, Code: httputil.ErrorCode(code), Message: msg}
}

type testMessageTemplateProvider struct {
	real provider.MessageTemplateProvider
	s    *testSession
}

func (p *testMessageTemplateProvider) MessageTemplate(ctx context.Context, id string) (*message.MessageData, error) {
	return p.real.MessageTemplate(ctx, id)
}

// LinkMessageTemplateInstance remembers which template a message was sent
// from, so clicks on its buttons run the template's flows.
func (p *testMessageTemplateProvider) LinkMessageTemplateInstance(ctx context.Context, instance provider.MessageTemplateInstance) error {
	p.s.linkTemplate(instance.MessageID, instance.MessageTemplateID)
	return nil
}

// testVariableProvider reads the app's variables, but keeps changes to them
// for the test run only.
type testVariableProvider struct {
	real provider.VariableProvider
	s    *testSession
	run  *flow.TestRun
}

func (p *testVariableProvider) VariableScoped(ctx context.Context, id string) (bool, error) {
	return p.real.VariableScoped(ctx, id)
}

func (p *testVariableProvider) Variable(ctx context.Context, id string, scope null.String) (thing.Thing, error) {
	if value, deleted, ok := p.s.variable(id, scope); ok {
		if deleted {
			return thing.Null, provider.ErrNotFound
		}
		return value, nil
	}
	return p.real.Variable(ctx, id, scope)
}

func (p *testVariableProvider) UpdateVariable(ctx context.Context, id string, scope null.String, operation provider.VariableOperation, value thing.Thing) (thing.Thing, error) {
	newValue := value
	if !operation.IsOverwrite() {
		current, err := p.Variable(ctx, id, scope)
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return thing.Null, err
		}
		if err == nil {
			switch operation {
			case provider.VariableOperationAppend:
				newValue = current.Append(value)
			case provider.VariableOperationPrepend:
				newValue = value.Append(current)
			case provider.VariableOperationIncrement:
				newValue = current.Add(value)
			case provider.VariableOperationDecrement:
				newValue = current.Sub(value)
			}
		}
	}

	p.s.setVariable(id, scope, newValue, false)
	p.run.AddEffect(flow.TestEffect{
		Action: "set_variable",
		Data:   map[string]any{"variable_id": id, "scope": scope, "operation": operation, "value": newValue.Value},
	})
	return newValue, nil
}

func (p *testVariableProvider) DeleteVariable(ctx context.Context, id string, scope null.String) error {
	p.s.setVariable(id, scope, thing.Null, true)
	p.run.AddEffect(flow.TestEffect{
		Action: "delete_variable",
		Data:   map[string]any{"variable_id": id, "scope": scope},
	})
	return nil
}

type testResumePointProvider struct {
	s *testSession
}

func (p *testResumePointProvider) CreateResumePoint(ctx context.Context, rp flow.ResumePoint) (flow.ResumePoint, error) {
	if rp.ID == "" {
		rp.ID = flow.TestSnowflake().String()
	}
	p.s.addResumePoint(rp)
	return rp, nil
}

type testLogProvider struct {
	run *flow.TestRun
}

func (p testLogProvider) CreateLogEntry(ctx context.Context, level provider.LogLevel, message string) {
	p.run.AddEffect(flow.TestEffect{
		Action: "log",
		Data:   map[string]any{"level": level, "message": message},
	})
}

// testSecretProvider stands in for the app's secrets, which test runs don't
// reveal. Their values only matter for requests that aren't sent anyway.
type testSecretProvider struct{}

func (testSecretProvider) Secrets(ctx context.Context, names []string) (map[string]string, error) {
	res := make(map[string]string, len(names))
	for _, name := range names {
		res[name] = "<secret " + name + ">"
	}
	return res, nil
}

// testIntegrationProvider tells which integrations the app enabled, but not
// their credentials.
type testIntegrationProvider struct {
	real provider.IntegrationProvider
}

func (p testIntegrationProvider) Credential(ctx context.Context, integrationID string) (string, error) {
	if _, err := p.real.Credential(ctx, integrationID); err != nil {
		return "", err
	}
	return "<credential>", nil
}

func (p testIntegrationProvider) Choice(ctx context.Context, integrationID string) (null.Bool, error) {
	return p.real.Choice(ctx, integrationID)
}

type testHTTPProvider struct{}

func (testHTTPProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, errNotInTestRuns
}

func (testHTTPProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, errNotInTestRuns
}

type testAIProvider struct{}

func (testAIProvider) CreateResponse(ctx context.Context, opts provider.CreateResponseOpts) (string, error) {
	return "", errNotInTestRuns
}

type testRobloxProvider struct{}

func (testRobloxProvider) UserByID(ctx context.Context, id int64) (*thing.RobloxUserValue, error) {
	return nil, errNotInTestRuns
}

func (testRobloxProvider) UsersByUsername(ctx context.Context, username string) ([]thing.RobloxUserValue, error) {
	return nil, errNotInTestRuns
}

// testDecisionProvider answers AI decisions with the mock of the block, AI
// isn't asked in test runs.
type testDecisionProvider struct {
	run *flow.TestRun
}

func (p testDecisionProvider) Decide(ctx context.Context, opts provider.DecideOpts) (provider.Decision, error) {
	mock, ok := p.run.CurrentMock()
	if !ok {
		switch opts.Type {
		case provider.DecisionTypeChoice:
			p.run.AddNote("AI isn't asked in test runs, so no option was picked. Mock the block with the name of the option to pick.")
		case provider.DecisionTypeNoul:
			p.run.AddNote(`AI isn't asked in test runs, so the answer was no. Mock the block with "yes", "no" or the probability of yes from 0 to 1.`)
		}
		return provider.Decision{}, nil
	}

	if err := mock.Err(); err != nil {
		return provider.Decision{}, err
	}

	var v any
	if err := json.Unmarshal(mock.Result, &v); err != nil {
		return provider.Decision{}, fmt.Errorf("invalid mock result: %w", err)
	}

	decision := provider.Decision{Confidence: 1}
	switch opts.Type {
	case provider.DecisionTypeChoice:
		choice, ok := v.(string)
		if !ok {
			return provider.Decision{}, errors.New("invalid mock result: must be the name of an option")
		}
		decision.Choice = choice
	case provider.DecisionTypeNoul:
		switch v := v.(type) {
		case string:
			switch strings.ToLower(v) {
			case "yes":
				decision.Noul = 1
			case "no":
				decision.Noul = 0
			default:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return provider.Decision{}, errors.New(`invalid mock result: must be "yes", "no" or a probability`)
				}
				decision.Noul = f
			}
		case float64:
			decision.Noul = v
		case bool:
			if v {
				decision.Noul = 1
			}
		default:
			return provider.Decision{}, errors.New(`invalid mock result: must be "yes", "no" or a probability`)
		}
	}
	return decision, nil
}
