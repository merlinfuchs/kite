package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"gopkg.in/guregu/null.v4"
)

const (
	// testRunTimeout bounds a whole test, with every step of it.
	testRunTimeout = 60 * time.Second
	// maxTestActions bounds the clicks and submits of a test.
	maxTestActions = 10
)

// ErrTestAppOffline means the app has no gateway on this cluster, so there's
// no Discord session to read from.
var ErrTestAppOffline = errors.New("app is offline")

// FlowTestError is a problem with the test itself, like an argument the
// command doesn't have.
type FlowTestError struct {
	Message string
}

func (e *FlowTestError) Error() string {
	return e.Message
}

func testErrorf(format string, args ...any) error {
	return &FlowTestError{Message: fmt.Sprintf(format, args...)}
}

type FlowTestKind string

const (
	FlowTestKindCommand       FlowTestKind = "command"
	FlowTestKindEventListener FlowTestKind = "event_listener"
	FlowTestKindMessage       FlowTestKind = "message"
)

// FlowTest is a test run of a flow: what triggers it, the mocks of its blocks
// and what the user does after, like clicking a button the flow sent.
type FlowTest struct {
	Kind FlowTestKind
	// Flow of the command or event listener.
	Flow flow.FlowData
	// The message template and the ID of the flow of its button or select
	// menu, for messages.
	Message      *model.Message
	FlowSourceID string

	// Who triggers the flow, and where. GuildID and ChannelID default to the
	// bot's only server and its first text channel.
	UserID    string
	GuildID   string
	ChannelID string
	// DM runs a command in a DM with the bot instead of a server.
	DM bool

	// Arguments of a command by name.
	Options map[string]json.RawMessage
	// Values picked in a select menu of a message.
	Values []string
	// Event is merged into the Discord event or interaction, in Discord's
	// format, like {"content": "hi"} for a message.
	Event   json.RawMessage
	Webhook *FlowTestWebhook

	Mocks map[string]flow.TestMock
	Then  []FlowTestAction
}

type FlowTestWebhook struct {
	Body    string
	Headers map[string]string
	Query   map[string]string
}

// FlowTestAction is something the user does after the flow ran: clicking a
// button or using a select menu of a message the flow sent, or submitting the
// modal it showed.
type FlowTestAction struct {
	// Label or custom ID of a button.
	Click string
	// Placeholder or custom ID of a select menu, with the values to pick.
	Select string
	Values []string
	// Values of the modal's inputs by custom ID or label.
	Submit map[string]json.RawMessage
	// Who does it, by default the user who triggered the flow.
	UserID string
}

type FlowTestResult struct {
	Executions []FlowTestExecution
	// What the test would have cost, in credits.
	Credits int
	// Mocked blocks that never ran, likely wrong IDs.
	UnusedMocks []string
}

// FlowTestExecution is one run of a flow in a test, like the command or a
// click on a button it sent.
type FlowTestExecution struct {
	Trigger   string
	Steps     []flow.TestStep
	Truncated bool
	// Problems of the execution as a whole, like an interaction nothing
	// responded to.
	Notes []string
	// The error that ended the execution, if no Error Handler caught it.
	Error string
}

// FlowTestRunner runs test runs of flows for the API, see FlowTest. Sessions
// come from the gateway manager, which is created after the engine.
type FlowTestRunner struct {
	engine   *Engine
	sessions SessionProvider
}

func (e *Engine) FlowTestRunner(sessions SessionProvider) *FlowTestRunner {
	return &FlowTestRunner{
		engine:   e,
		sessions: sessions,
	}
}

// Run runs a test. Problems with the test are returned as FlowTestError,
// problems of the flow are part of the result.
func (r *FlowTestRunner) Run(ctx context.Context, appID string, test FlowTest) (*FlowTestResult, error) {
	session, err := r.sessions.AppSession(ctx, appID)
	if err != nil || session == nil {
		return nil, ErrTestAppOffline
	}

	if len(test.Then) > maxTestActions {
		return nil, testErrorf("a test can have at most %d steps after the trigger", maxTestActions)
	}

	ctx, cancel := context.WithTimeout(ctx, testRunTimeout)
	defer cancel()

	var root *flow.CompiledFlowNode
	switch test.Kind {
	case FlowTestKindCommand:
		root, err = flow.CompileCommand(test.Flow)
	case FlowTestKindEventListener:
		root, err = flow.CompileEventListener(test.Flow)
	case FlowTestKindMessage:
	default:
		return nil, testErrorf("unknown kind %q", test.Kind)
	}
	if err != nil {
		return nil, testErrorf("the flow is invalid: %v", err)
	}

	s := newTestSession(r.engine.env, appID, session, test.Mocks)
	if err := s.setLocation(test, root); err != nil {
		return nil, err
	}
	if err := s.start(ctx, test, root); err != nil {
		return nil, err
	}

	for i, action := range test.Then {
		if ctx.Err() != nil {
			break
		}
		if err := s.act(ctx, action); err != nil {
			var testErr *FlowTestError
			if !errors.As(err, &testErr) {
				return nil, err
			}
			// Later steps can depend on what earlier ones did, so the steps
			// that ran are still worth seeing.
			s.result.Executions = append(s.result.Executions, FlowTestExecution{
				Trigger: fmt.Sprintf("Step %d", i+1),
				Error:   testErr.Message,
			})
			break
		}
	}

	if ctx.Err() != nil {
		s.result.Executions = append(s.result.Executions, FlowTestExecution{
			Error: fmt.Sprintf("The test took longer than %s and was stopped.", testRunTimeout),
		})
	}

	for id := range s.mocks {
		if !s.usedMocks[id] {
			s.result.UnusedMocks = append(s.result.UnusedMocks, id)
		}
	}
	slices.Sort(s.result.UnusedMocks)

	return &s.result, nil
}

// testSession is the state of a test, shared by its executions: what the
// flows sent and changed, so later steps see it.
type testSession struct {
	env     Env
	appID   string
	session *state.State
	botID   discord.UserID
	mocks   map[string]flow.TestMock

	// Where the test runs. guildID is 0 in DMs.
	guildID   discord.GuildID
	channelID discord.ChannelID
	userID    discord.UserID

	result    FlowTestResult
	usedMocks map[string]bool

	// The real providers the ones of the test read from.
	real      flow.FlowProviders
	cooldowns *provider.MemoryCooldownProvider

	mu       sync.Mutex
	messages []discord.Message
	channels map[discord.ChannelID]discord.Channel
	// The template each message was sent from, and the templates by ID.
	templates     map[discord.MessageID]string
	templateCache map[string]*model.Message
	// Members and servers read from Discord, nil members for users that
	// aren't members.
	members      map[testMemberKey]*discord.Member
	guilds       map[discord.GuildID]*testGuild
	respondedTo  map[discord.InteractionID]bool
	originals    map[discord.InteractionID]discord.MessageID
	current      *discord.InteractionEvent
	modal        *api.InteractionResponseData
	root         *flow.CompiledFlowNode
	resumePoints map[string]testResumePoint
	timers       []testResumePoint
	variables    map[testVariableKey]testVariableValue
}

type testMemberKey struct {
	guildID discord.GuildID
	userID  discord.UserID
}

type testResumePoint struct {
	rp   flow.ResumePoint
	root *flow.CompiledFlowNode
}

type testVariableKey struct {
	id    string
	scope null.String
}

type testVariableValue struct {
	value   thing.Thing
	deleted bool
}

func newTestSession(env Env, appID string, session *state.State, mocks map[string]flow.TestMock) *testSession {
	return &testSession{
		env:       env,
		appID:     appID,
		session:   session,
		botID:     session.Ready().User.ID,
		mocks:     mocks,
		usedMocks: make(map[string]bool),
		real:      env.flowProviders(appID, session, entityLinks{}),
		cooldowns: provider.NewMemoryCooldownProvider(),

		channels:      make(map[discord.ChannelID]discord.Channel),
		templates:     make(map[discord.MessageID]string),
		templateCache: make(map[string]*model.Message),
		members:       make(map[testMemberKey]*discord.Member),
		guilds:        make(map[discord.GuildID]*testGuild),
		respondedTo:   make(map[discord.InteractionID]bool),
		originals:     make(map[discord.InteractionID]discord.MessageID),
		resumePoints:  make(map[string]testResumePoint),
		variables:     make(map[testVariableKey]testVariableValue),
	}
}

// setLocation picks the user, server and channel of the test. root is nil
// for messages.
func (s *testSession) setLocation(test FlowTest, root *flow.CompiledFlowNode) error {
	userID, err := discord.ParseSnowflake(test.UserID)
	if err != nil || !userID.IsValid() {
		return testErrorf("user_id must be the ID of a Discord user")
	}
	s.userID = discord.UserID(userID)

	dm := test.DM
	// Scheduled and webhook runs have no user or place, but blocks still
	// default to the server and channel of the test.
	nowhere := false
	if test.Kind == FlowTestKindEventListener {
		switch root.EventListenerType() {
		case string(model.EventListenerTypeDiscordDirectMessageCreate):
			dm = true
		case flow.EventTypeScheduleCron, flow.EventTypeWebhook:
			nowhere = true
		}
	}

	if dm {
		s.channelID = discord.ChannelID(flow.TestSnowflake())
		s.channels[s.channelID] = discord.Channel{
			ID:           s.channelID,
			Type:         discord.DirectMessage,
			DMRecipients: []discord.User{{ID: s.userID}},
		}
		if _, err := s.user(s.userID); err != nil {
			return testErrorf("user %s doesn't exist", test.UserID)
		}
		return nil
	}

	if test.GuildID == "" {
		guilds, _ := s.session.Guilds()
		if len(guilds) != 1 {
			if nowhere {
				return nil
			}
			return testErrorf("guild_id is required, the bot is in %d servers", len(guilds))
		}
		s.guildID = guilds[0].ID
	} else {
		guildID, err := discord.ParseSnowflake(test.GuildID)
		if err != nil {
			return testErrorf("guild_id must be the ID of a server")
		}
		s.guildID = discord.GuildID(guildID)
		if _, err := s.session.Guild(s.guildID); err != nil {
			return testErrorf("the bot isn't in server %s", test.GuildID)
		}
	}

	if test.ChannelID != "" {
		channelID, err := discord.ParseSnowflake(test.ChannelID)
		if err != nil {
			return testErrorf("channel_id must be the ID of a channel")
		}
		channel, err := s.session.Channel(discord.ChannelID(channelID))
		if err != nil || channel.GuildID != s.guildID {
			return testErrorf("channel %s isn't in server %s", test.ChannelID, s.guildID)
		}
		s.channelID = channel.ID
	} else {
		s.channelID = s.defaultChannel()
		if !s.channelID.IsValid() {
			return testErrorf("channel_id is required, the server has no text channel the bot can see")
		}
	}

	if nowhere {
		return nil
	}
	if _, err := s.member(s.guildID, s.userID); err != nil {
		return testErrorf("user %s isn't a member of server %s", s.userID, s.guildID)
	}
	return nil
}

// member returns a member of a server, once per test. Without the members
// intent, every lookup would be a request to Discord.
func (s *testSession) member(guildID discord.GuildID, userID discord.UserID) (*discord.Member, error) {
	key := testMemberKey{guildID: guildID, userID: userID}
	s.mu.Lock()
	member, ok := s.members[key]
	s.mu.Unlock()
	if !ok {
		member, _ = s.session.Member(guildID, userID)
		s.mu.Lock()
		s.members[key] = member
		s.mu.Unlock()
	}

	if member == nil {
		return nil, fmt.Errorf("user %s isn't a member of server %s", userID, guildID)
	}
	return member, nil
}

// user returns a user, from their member in the test's server if they are
// one, which saves a request to Discord.
func (s *testSession) user(id discord.UserID) (*discord.User, error) {
	if s.guildID.IsValid() {
		if member, err := s.member(s.guildID, id); err == nil && member.User.ID.IsValid() {
			return &member.User, nil
		}
	}
	return s.session.User(id)
}

// defaultChannel is the server's system channel, or its first text channel.
func (s *testSession) defaultChannel() discord.ChannelID {
	if guild, err := s.session.Guild(s.guildID); err == nil && guild.SystemChannelID.IsValid() {
		return guild.SystemChannelID
	}

	channels, err := s.session.Channels(s.guildID)
	if err != nil {
		return 0
	}
	slices.SortFunc(channels, func(a, b discord.Channel) int { return a.Position - b.Position })
	for _, c := range channels {
		if c.Type == discord.GuildText {
			return c.ID
		}
	}
	return 0
}

// start runs the flow with the trigger of the test. root is nil for messages.
func (s *testSession) start(ctx context.Context, test FlowTest, root *flow.CompiledFlowNode) error {
	switch test.Kind {
	case FlowTestKindCommand:
		event, err := s.commandEvent(root, test)
		if err != nil {
			return err
		}
		s.execute(ctx, "/"+root.CommandName(), root, root, event, nil, nil)
		return nil
	case FlowTestKindEventListener:
		event, err := s.listenerEvent(root, test)
		if err != nil {
			return err
		}
		t := model.EventListenerType(root.EventListenerType())
		listener := &EventListener{listener: &model.EventListener{Type: t, Source: model.EventSourceForType(t)}}
		s.execute(ctx, "Event "+root.EventListenerType(), root, root, event, nil, listener)
		return nil
	}

	msg := test.Message
	if _, ok := msg.FlowSources[test.FlowSourceID]; !ok {
		return testErrorf("message %s has no button or select menu with flow %s", msg.ID, test.FlowSourceID)
	}

	// The message as it was sent, so the click runs like any other.
	data := msg.Data.Copy()
	send := data.ToSendMessageData(message.ConvertOptions{
		ComponentIDFactory: func(c *message.ComponentData) discord.ComponentID {
			return discord.ComponentID(c.FlowSourceID)
		},
	})
	sent := s.addMessage(discord.Message{
		ChannelID:  s.channelID,
		GuildID:    s.guildID,
		Content:    send.Content,
		Embeds:     send.Embeds,
		Components: send.Components,
	})
	s.mu.Lock()
	s.templateCache[msg.ID] = msg
	s.mu.Unlock()
	s.linkTemplate(sent.ID, msg.ID)

	action := FlowTestAction{Click: test.FlowSourceID}
	if len(test.Values) > 0 {
		action = FlowTestAction{Select: test.FlowSourceID, Values: test.Values}
	}
	return s.act(ctx, action)
}

// execute runs node with event and continues the flow after durable sleeps
// right away. listener is set for the first run of an event listener, whose
// filters decide if it runs.
func (s *testSession) execute(
	ctx context.Context,
	trigger string,
	root *flow.CompiledFlowNode,
	node *flow.CompiledFlowNode,
	event gateway.Event,
	state *flow.FlowContextState,
	listener *EventListener,
) {
	s.runOnce(ctx, trigger, root, node, event, state, listener, false)

	// The flow limits how often it can wait, this only guards the loop.
	for range 20 {
		s.mu.Lock()
		if len(s.timers) == 0 || ctx.Err() != nil {
			s.mu.Unlock()
			return
		}
		timer := s.timers[0]
		s.timers = s.timers[1:]
		s.mu.Unlock()

		node := timer.root.FindChildWithID(timer.rp.NodeID, true)
		event, state, err := resumeTriggerEvent(timer.rp.State, timer.rp.InteractionToken)
		if node == nil || err != nil {
			continue
		}
		s.runOnce(ctx, "After waiting until "+timer.rp.ResumeAt.Format(time.RFC3339), timer.root, node, event, &state, nil, true)
	}
}

func (s *testSession) runOnce(
	ctx context.Context,
	trigger string,
	root *flow.CompiledFlowNode,
	node *flow.CompiledFlowNode,
	event gateway.Event,
	state *flow.FlowContextState,
	listener *EventListener,
	afterSleep bool,
) {
	exec := FlowTestExecution{Trigger: trigger}
	defer func() {
		s.result.Executions = append(s.result.Executions, exec)
	}()

	data := flowContextData(event)
	if listener != nil {
		if listener.listener.Source == model.EventSourceDiscord && !listener.shouldHandleEvent(event, s.botID) {
			exec.Notes = append(exec.Notes, "Event listeners ignore this event, like messages of bots, so nothing ran.")
			return
		}
		if !node.FilterEvent(data, s.botID) {
			exec.Notes = append(exec.Notes, "The event filter of the flow doesn't match the event, so nothing ran.")
			return
		}
	}

	interaction := data.Interaction()
	s.mu.Lock()
	s.root = root
	s.current = interaction
	s.mu.Unlock()

	run := flow.NewTestRun(s.mocks)
	fCtx := s.env.flowContextWithProviders(ctx, s.appID, s.session, data, s.testProviders(run), state, node)
	defer fCtx.Cancel()
	fCtx.Test = run

	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error(
					"Recovered from panic in flow test run",
					slog.String("app_id", s.appID),
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				err = fmt.Errorf("the flow crashed: %v", r)
			}
		}()

		if afterSleep {
			if interaction != nil {
				fCtx.Discord.MarkInteractionResponded(interaction.ID)
			}
			return node.ResumeAfterSleep(fCtx)
		}
		return node.Execute(fCtx)
	}()

	s.result.Credits += fCtx.CreditsUsed()
	exec.Steps = run.Steps()
	exec.Truncated = run.Truncated()
	for _, id := range run.UsedMocks() {
		s.usedMocks[id] = true
	}
	if err != nil {
		exec.Error = err.Error()
	}

	if interaction != nil && !s.responded(interaction.ID) {
		exec.Notes = append(exec.Notes, "Nothing responded to the interaction, so Discord shows that it failed.")
	}
}

// act does what the user does after the flow ran, see FlowTestAction.
func (s *testSession) act(ctx context.Context, action FlowTestAction) error {
	userID := s.userID
	if action.UserID != "" {
		id, err := discord.ParseSnowflake(action.UserID)
		if err != nil {
			return testErrorf("user_id must be the ID of a Discord user")
		}
		userID = discord.UserID(id)
	}

	switch {
	case action.Submit != nil:
		return s.submitModal(ctx, userID, action.Submit)
	case action.Click != "":
		return s.useComponent(ctx, userID, action.Click, false, nil)
	case action.Select != "":
		if len(action.Values) == 0 {
			return testErrorf("values are required to use select menu %q", action.Select)
		}
		return s.useComponent(ctx, userID, action.Select, true, action.Values)
	}
	return testErrorf("a step needs click, select or submit")
}

func (s *testSession) useComponent(ctx context.Context, userID discord.UserID, target string, isSelect bool, values []string) error {
	msg, component, ok := s.findComponent(target, isSelect)
	if !ok {
		kind := "button"
		if isSelect {
			kind = "select menu"
		}
		available := s.availableComponents()
		if len(available) == 0 {
			return testErrorf("no %s %q, the flow sent no message with buttons or select menus", kind, target)
		}
		return testErrorf("no %s %q, the messages the flow sent have: %s", kind, target, strings.Join(available, ", "))
	}

	data := map[string]any{
		"custom_id":      component.CustomID,
		"component_type": component.Type,
	}
	label := fmt.Sprintf("Click %q", target)
	if isSelect {
		label = fmt.Sprintf("Select %s in %q", strings.Join(values, ", "), target)
		data["values"] = values
		if component.Type != int(discord.StringSelectComponentType) {
			resolved := map[string]any{}
			if err := s.resolveSelect(resolved, component.Type, values); err != nil {
				return testErrorf("%v", err)
			}
			data["resolved"] = resolved
		}
	}

	guildID, channelID := msg.GuildID, msg.ChannelID
	if !channelID.IsValid() {
		guildID, channelID = s.guildID, s.channelID
	}
	event, err := s.interactionEvent(userID, guildID, channelID, discord.ComponentInteractionType, data, &msg, nil)
	if err != nil {
		return err
	}

	if rpID, _, ok := message.DecodeCustomIDMessageComponentResumePoint(component.CustomID); ok {
		return s.resume(ctx, label, rpID, event)
	}

	tmpl, err := s.messageTemplate(ctx, msg.ID)
	if err != nil {
		return err
	}
	if tmpl == nil {
		return testErrorf("%q is on a message no flow handles", target)
	}
	source, ok := tmpl.FlowSources[component.CustomID]
	if !ok {
		return testErrorf("%q has no flow in message %s", target, tmpl.ID)
	}
	root, err := flow.CompileComponentButton(source)
	if err != nil {
		return testErrorf("the flow of %q is invalid: %v", target, err)
	}
	s.execute(ctx, label, root, root, event, nil, nil)
	return nil
}

func (s *testSession) submitModal(ctx context.Context, userID discord.UserID, values map[string]json.RawMessage) error {
	s.mu.Lock()
	modal := s.modal
	s.modal = nil
	s.mu.Unlock()
	if modal == nil || modal.CustomID == nil {
		return testErrorf("no modal is open to submit")
	}

	components, resolved, err := s.modalSubmission(modal, values)
	if err != nil {
		return err
	}

	data := map[string]any{
		"custom_id":  modal.CustomID.Val,
		"components": components,
	}
	if len(resolved) > 0 {
		data["resolved"] = resolved
	}
	event, err := s.interactionEvent(userID, s.guildID, s.channelID, discord.ModalInteractionType, data, nil, nil)
	if err != nil {
		return err
	}

	rpID, _ := message.DecodeCustomIDModalResumePoint(modal.CustomID.Val)
	return s.resume(ctx, "Submit the modal", rpID, event)
}

// resume continues the flow that waits at a resume point, like clicks and
// submits do.
func (s *testSession) resume(ctx context.Context, label string, rpID string, event gateway.Event) error {
	s.mu.Lock()
	rp, ok := s.resumePoints[rpID]
	s.mu.Unlock()
	if !ok {
		return testErrorf("%s: no flow is waiting for it", label)
	}
	node := rp.root.FindChildWithID(rp.rp.NodeID, true)
	if node == nil {
		return testErrorf("%s: the block that waits for it doesn't exist", label)
	}

	// A click or submit starts a new execution, like in resumeFlow.
	state := rp.rp.State.Copy()
	state.DurableSleeps = 0
	s.execute(ctx, label, rp.root, node, event, &state, nil)
	return nil
}

// messageTemplate returns the template a message was sent from, nil if none.
func (s *testSession) messageTemplate(ctx context.Context, messageID discord.MessageID) (*model.Message, error) {
	s.mu.Lock()
	id := s.templates[messageID]
	tmpl := s.templateCache[id]
	s.mu.Unlock()
	if id == "" || tmpl != nil {
		return tmpl, nil
	}

	tmpl, err := s.env.MessageStore.Message(ctx, s.appID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message template: %w", err)
	}
	s.mu.Lock()
	s.templateCache[id] = tmpl
	s.mu.Unlock()
	return tmpl, nil
}

// modalSubmission fills in the inputs of the modal with values, in the format
// Discord sends a submitted modal in.
func (s *testSession) modalSubmission(modal *api.InteractionResponseData, values map[string]json.RawMessage) ([]any, map[string]any, error) {
	var sent []map[string]any
	if modal.Components != nil {
		b, err := json.Marshal(modal.Components)
		if err != nil {
			return nil, nil, err
		}
		if err := json.Unmarshal(b, &sent); err != nil {
			return nil, nil, err
		}
	}

	used := make(map[string]bool)
	resolved := make(map[string]any)
	var inputs []string

	fill := func(input map[string]any, label string) (map[string]any, error) {
		customID, _ := input["custom_id"].(string)
		if label == "" {
			label, _ = input["label"].(string)
		}
		inputs = append(inputs, fmt.Sprintf("%s (%s)", customID, label))

		raw, ok := values[customID]
		key := customID
		if !ok {
			raw, ok = values[label]
			key = label
		}
		if ok {
			used[key] = true
		}

		typ, _ := input["type"].(float64)
		res := map[string]any{"type": typ, "custom_id": customID}
		switch discord.ComponentType(typ) {
		case discord.TextInputComponentType, discord.RadioGroupComponentType:
			value := ""
			if ok {
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("the value of %s must be text", key)
				}
			}
			res["value"] = value
		case discord.CheckboxComponentType:
			value := false
			if ok {
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("the value of %s must be true or false", key)
				}
			}
			res["value"] = value
		default:
			picked := []string{}
			if ok {
				if err := json.Unmarshal(raw, &picked); err != nil {
					var one string
					if json.Unmarshal(raw, &one) != nil {
						return nil, fmt.Errorf("the value of %s must be a list of values", key)
					}
					picked = []string{one}
				}
			}
			res["values"] = picked
			if discord.ComponentType(typ) != discord.StringSelectComponentType && discord.ComponentType(typ) != discord.CheckboxGroupComponentType && len(picked) > 0 {
				if err := s.resolveSelect(resolved, int(typ), picked); err != nil {
					return nil, err
				}
			}
		}
		return res, nil
	}

	var res []any
	for _, c := range sent {
		switch typ, _ := c["type"].(float64); discord.ComponentType(typ) {
		case discord.LabelComponentType:
			input, _ := c["component"].(map[string]any)
			if input == nil {
				continue
			}
			label, _ := c["label"].(string)
			filled, err := fill(input, label)
			if err != nil {
				return nil, nil, testErrorf("%v", err)
			}
			res = append(res, map[string]any{"type": typ, "component": filled})
		case discord.ActionRowComponentType:
			items, _ := c["components"].([]any)
			var row []any
			for _, item := range items {
				input, _ := item.(map[string]any)
				filled, err := fill(input, "")
				if err != nil {
					return nil, nil, testErrorf("%v", err)
				}
				row = append(row, filled)
			}
			res = append(res, map[string]any{"type": typ, "components": row})
		}
	}

	for key := range values {
		if !used[key] {
			return nil, nil, testErrorf("the modal has no input %q, it has: %s", key, strings.Join(inputs, ", "))
		}
	}
	return res, resolved, nil
}

// interactionEvent builds an interaction like Discord sends it, from the user
// and the channel. override is merged into it.
func (s *testSession) interactionEvent(
	userID discord.UserID,
	guildID discord.GuildID,
	channelID discord.ChannelID,
	t discord.InteractionDataType,
	data map[string]any,
	msg *discord.Message,
	override json.RawMessage,
) (gateway.Event, error) {
	user, err := s.user(userID)
	if err != nil {
		return nil, testErrorf("user %s doesn't exist", userID)
	}

	appID := s.session.Ready().Application.ID
	if !appID.IsValid() {
		appID = discord.AppID(s.botID)
	}

	i := map[string]any{
		"id":             flow.TestSnowflake().String(),
		"application_id": appID.String(),
		"type":           t,
		"token":          "test-run",
		"version":        1,
		"channel_id":     channelID.String(),
		"locale":         "en-US",
		"data":           data,
		"entitlements":   []any{},
	}
	if msg != nil {
		i["message"] = msg
	}

	channel, ok := s.fakeChannel(channelID)
	if !ok {
		if c, err := s.session.Channel(channelID); err == nil {
			channel, ok = *c, true
		}
	}
	if ok {
		i["channel"] = channel
	}

	if guildID.IsValid() {
		member, err := s.member(guildID, userID)
		if err != nil {
			return nil, testErrorf("user %s isn't a member of server %s", userID, guildID)
		}
		m := *member
		m.User = *user
		i["guild_id"] = guildID.String()
		i["member"] = m
		i["context"] = discord.InteractionContextGuild
		i["authorizing_integration_owners"] = map[string]string{"0": guildID.String()}
		if g := s.testGuild(guildID); g.problem == "" {
			i["guild_locale"] = g.guild.PreferredLocale
			if ok {
				i["app_permissions"] = strconv.FormatUint(uint64(g.permissions(channel)), 10)
			}
		}
	} else {
		i["user"] = user
		i["context"] = discord.InteractionContextBotDM
		i["authorizing_integration_owners"] = map[string]string{"1": userID.String()}
	}

	var event discord.InteractionEvent
	if err := mergeJSON(i, override, &event); err != nil {
		return nil, testErrorf("invalid interaction: %v", err)
	}
	return &gateway.InteractionCreateEvent{InteractionEvent: event}, nil
}

// commandEvent builds the interaction of the command with the arguments of
// the test.
func (s *testSession) commandEvent(root *flow.CompiledFlowNode, test FlowTest) (gateway.Event, error) {
	var args []map[string]any
	for _, arg := range root.CommandArguments() {
		b, err := json.Marshal(arg)
		if err != nil {
			return nil, err
		}
		var a map[string]any
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		args = append(args, a)
	}

	names := make([]string, len(args))
	for i, arg := range args {
		names[i], _ = arg["name"].(string)
	}
	for name := range test.Options {
		if !slices.Contains(names, name) {
			if len(names) == 0 {
				return nil, testErrorf("the command has no arguments, got %q", name)
			}
			return nil, testErrorf("the command has no argument %q, it has: %s", name, strings.Join(names, ", "))
		}
	}

	resolved := make(map[string]any)
	options := []map[string]any{}
	for _, arg := range args {
		name, _ := arg["name"].(string)
		raw, ok := test.Options[name]
		if !ok || string(raw) == "null" {
			if required, _ := arg["required"].(bool); required {
				return nil, testErrorf("argument %q is required", name)
			}
			continue
		}

		typ, _ := arg["type"].(float64)
		value, err := s.commandOptionValue(arg, discord.CommandOptionType(typ), raw, resolved)
		if err != nil {
			return nil, testErrorf("argument %q: %v", name, err)
		}
		options = append(options, map[string]any{"name": name, "type": typ, "value": value})
	}

	// Subcommands are the words after the first in the name.
	parts := strings.Split(root.CommandName(), " ")
	var top any = options
	switch len(parts) {
	case 2:
		top = []any{map[string]any{"type": discord.SubcommandOptionType, "name": parts[1], "options": options}}
	case 3:
		top = []any{map[string]any{"type": discord.SubcommandGroupOptionType, "name": parts[1], "options": []any{
			map[string]any{"type": discord.SubcommandOptionType, "name": parts[2], "options": options},
		}}}
	}

	data := map[string]any{
		"id":       flow.TestSnowflake().String(),
		"name":     parts[0],
		"type":     discord.ChatInputCommand,
		"options":  top,
		"resolved": resolved,
	}
	if s.guildID.IsValid() {
		data["guild_id"] = s.guildID.String()
	}
	return s.interactionEvent(s.userID, s.guildID, s.channelID, discord.CommandInteractionType, data, nil, test.Event)
}

func (s *testSession) commandOptionValue(arg map[string]any, t discord.CommandOptionType, raw json.RawMessage, resolved map[string]any) (any, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}

	switch t {
	case discord.StringOptionType:
		str, ok := v.(string)
		if !ok {
			return nil, errors.New("must be text")
		}
		if max, ok := arg["max_length"].(float64); ok && len([]rune(str)) > int(max) {
			return nil, fmt.Errorf("must be at most %d characters", int(max))
		}
		if err := checkChoice(arg, str); err != nil {
			return nil, err
		}
		return str, nil
	case discord.IntegerOptionType, discord.NumberOptionType:
		f, ok := v.(float64)
		if str, isStr := v.(string); isStr {
			parsed, err := strconv.ParseFloat(strings.TrimSpace(str), 64)
			f, ok = parsed, err == nil
		}
		if !ok {
			return nil, errors.New("must be a number")
		}
		if t == discord.IntegerOptionType && f != float64(int64(f)) {
			return nil, errors.New("must be a whole number")
		}
		if min, ok := arg["min_value"].(float64); ok && f < min {
			return nil, fmt.Errorf("must be at least %v", min)
		}
		if max, ok := arg["max_value"].(float64); ok && f > max {
			return nil, fmt.Errorf("must be at most %v", max)
		}
		if err := checkChoice(arg, strconv.FormatFloat(f, 'f', -1, 64)); err != nil {
			return nil, err
		}
		return f, nil
	case discord.BooleanOptionType:
		switch v := v.(type) {
		case bool:
			return v, nil
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b, nil
			}
		}
		return nil, errors.New("must be true or false")
	case discord.UserOptionType, discord.ChannelOptionType, discord.RoleOptionType, discord.MentionableOptionType:
		id, ok := v.(string)
		if !ok {
			return nil, errors.New("must be an ID")
		}
		componentType := map[discord.CommandOptionType]int{
			discord.UserOptionType:        int(discord.UserSelectComponentType),
			discord.ChannelOptionType:     int(discord.ChannelSelectComponentType),
			discord.RoleOptionType:        int(discord.RoleSelectComponentType),
			discord.MentionableOptionType: int(discord.MentionableSelectComponentType),
		}[t]
		if err := s.resolveSelect(resolved, componentType, []string{id}); err != nil {
			return nil, err
		}
		return id, nil
	case discord.AttachmentOptionType:
		var att struct {
			URL         string `json:"url"`
			Filename    string `json:"filename"`
			ContentType string `json:"content_type"`
		}
		if url, ok := v.(string); ok {
			att.URL = url
		} else if err := json.Unmarshal(raw, &att); err != nil || att.URL == "" {
			return nil, errors.New(`must be the URL of a file or {"url": ..., "filename": ...}`)
		}
		if att.Filename == "" {
			att.Filename = att.URL[strings.LastIndex(att.URL, "/")+1:]
			att.Filename, _, _ = strings.Cut(att.Filename, "?")
		}
		id := flow.TestSnowflake().String()
		addResolved(resolved, "attachments", id, map[string]any{
			"id":           id,
			"filename":     att.Filename,
			"url":          att.URL,
			"proxy_url":    att.URL,
			"content_type": att.ContentType,
			"size":         0,
		})
		return id, nil
	}
	return v, nil
}

func checkChoice(arg map[string]any, value string) error {
	choices, _ := arg["choices"].([]any)
	if len(choices) == 0 {
		return nil
	}
	var values []string
	for _, c := range choices {
		choice, _ := c.(map[string]any)
		v := fmt.Sprint(choice["value"])
		if v == value {
			return nil
		}
		values = append(values, v)
	}
	return fmt.Errorf("must be one of the choices: %s", strings.Join(values, ", "))
}

// resolveSelect adds the resolved data Discord sends for the users, roles or
// channels picked in a select menu or command argument of the type to dst.
func (s *testSession) resolveSelect(dst map[string]any, componentType int, values []string) error {
	addUser := func(id discord.Snowflake) bool {
		user, err := s.user(discord.UserID(id))
		if err != nil {
			return false
		}
		addResolved(dst, "users", id.String(), user)
		if s.guildID.IsValid() {
			if member, err := s.member(s.guildID, discord.UserID(id)); err == nil {
				addResolved(dst, "members", id.String(), member)
			}
		}
		return true
	}
	addRole := func(id discord.Snowflake) bool {
		role, err := s.session.Role(s.guildID, discord.RoleID(id))
		if err != nil {
			return false
		}
		addResolved(dst, "roles", id.String(), role)
		return true
	}

	for _, value := range values {
		id, err := discord.ParseSnowflake(value)
		if err != nil {
			return fmt.Errorf("%q isn't an ID", value)
		}

		switch discord.ComponentType(componentType) {
		case discord.UserSelectComponentType:
			if !addUser(id) {
				return fmt.Errorf("user %s doesn't exist", value)
			}
		case discord.RoleSelectComponentType:
			if !addRole(id) {
				return fmt.Errorf("role %s doesn't exist in the server", value)
			}
		case discord.MentionableSelectComponentType:
			if !addRole(id) && !addUser(id) {
				return fmt.Errorf("%s is neither a user nor a role of the server", value)
			}
		case discord.ChannelSelectComponentType:
			channel, err := s.session.Channel(discord.ChannelID(id))
			if err != nil || channel.GuildID != s.guildID {
				return fmt.Errorf("channel %s isn't in the server", value)
			}
			addResolved(dst, "channels", value, channel)
		}
	}
	return nil
}

// addResolved adds v to the resolved data of the kind, like "users".
func addResolved(dst map[string]any, kind string, id string, v any) {
	m, _ := dst[kind].(map[string]any)
	if m == nil {
		m = map[string]any{}
		dst[kind] = m
	}
	m[id] = v
}

// listenerEvent builds the event of the listener's type, with the user and
// channel of the test, and test.Event merged into it.
func (s *testSession) listenerEvent(root *flow.CompiledFlowNode, test FlowTest) (gateway.Event, error) {
	eventType := root.EventListenerType()
	now := time.Now().UTC()

	switch eventType {
	case flow.EventTypeScheduleCron:
		return &schedule.Event{Time: now}, nil
	case flow.EventTypeWebhook:
		w := test.Webhook
		if w == nil {
			w = &FlowTestWebhook{}
		}
		if len(w.Body) > webhook.MaxBodySize {
			return nil, testErrorf("the body can be at most %d bytes", webhook.MaxBodySize)
		}
		headers := make(map[string]string, len(w.Headers))
		for k, v := range w.Headers {
			headers[strings.ToLower(k)] = v
		}
		query := w.Query
		if query == nil {
			query = map[string]string{}
		}
		return &webhook.Event{
			Headers:    headers,
			Query:      query,
			Body:       w.Body,
			ReceivedAt: now,
			Responder:  webhook.NewResponder(),
		}, nil
	}

	user, err := s.user(s.userID)
	if err != nil {
		return nil, testErrorf("user %s doesn't exist", s.userID)
	}
	var member *discord.Member
	if s.guildID.IsValid() {
		member, _ = s.member(s.guildID, s.userID)
	}

	msg := map[string]any{
		"id":         flow.TestSnowflake().String(),
		"channel_id": s.channelID.String(),
		"author":     user,
		"content":    "",
		"timestamp":  now.Format(time.RFC3339),
		"type":       discord.DefaultMessage,
	}
	if s.guildID.IsValid() {
		msg["guild_id"] = s.guildID.String()
		if member != nil {
			msg["member"] = member
		}
	}

	reaction := func() map[string]any {
		return map[string]any{
			"user_id":    s.userID.String(),
			"channel_id": s.channelID.String(),
			"message_id": flow.TestSnowflake().String(),
			"guild_id":   s.guildID.String(),
			"emoji":      map[string]any{"name": "👍"},
		}
	}

	switch model.EventListenerType(eventType) {
	case model.EventListenerTypeDiscordMessageCreate, model.EventListenerTypeDiscordDirectMessageCreate:
		return decodeEvent[gateway.MessageCreateEvent](msg, test.Event)
	case model.EventListenerTypeDiscordGuildBoost:
		msg["type"] = discord.NitroBoostMessage
		return decodeEvent[gateway.MessageCreateEvent](msg, test.Event)
	case model.EventListenerTypeDiscordMessageUpdate:
		msg["edited_timestamp"] = now.Format(time.RFC3339)
		return decodeEvent[gateway.MessageUpdateEvent](msg, test.Event)
	case model.EventListenerTypeDiscordMessageDelete:
		deleted, err := decodeEvent[discord.Message](msg, test.Event)
		if err != nil {
			return nil, err
		}
		return &guildevent.MessageDelete{
			MessageDeleteEvent: gateway.MessageDeleteEvent{ID: deleted.ID, ChannelID: deleted.ChannelID, GuildID: deleted.GuildID},
			Message:            deleted,
		}, nil
	case model.EventListenerTypeDiscordGuildMemberAdd:
		base := map[string]any{"guild_id": s.guildID.String(), "user": user, "joined_at": now.Format(time.RFC3339), "roles": []string{}}
		return decodeEvent[gateway.GuildMemberAddEvent](base, test.Event)
	case model.EventListenerTypeDiscordGuildMemberRemove:
		base := map[string]any{"guild_id": s.guildID.String(), "user": user}
		return decodeEvent[gateway.GuildMemberRemoveEvent](base, test.Event)
	case model.EventListenerTypeDiscordMessageReactionAdd:
		base := reaction()
		if member != nil {
			base["member"] = member
		}
		return decodeEvent[gateway.MessageReactionAddEvent](base, test.Event)
	case model.EventListenerTypeDiscordMessageReactionRemove:
		return decodeEvent[gateway.MessageReactionRemoveEvent](reaction(), test.Event)
	case model.EventListenerTypeDiscordGuildCreate:
		guild, err := s.session.Guild(s.guildID)
		if err != nil {
			return nil, testErrorf("the bot isn't in server %s", s.guildID)
		}
		e, err := decodeEvent[gateway.GuildCreateEvent](map[string]any{"joined_at": now.Format(time.RFC3339)}, test.Event)
		if err != nil {
			return nil, err
		}
		if !e.ID.IsValid() {
			e.Guild = *guild
		}
		return &state.GuildJoinEvent{GuildCreateEvent: e}, nil
	case model.EventListenerTypeDiscordGuildDelete:
		e, err := decodeEvent[gateway.GuildDeleteEvent](map[string]any{"id": s.guildID.String()}, test.Event)
		if err != nil {
			return nil, err
		}
		return &state.GuildLeaveEvent{GuildDeleteEvent: e}, nil
	case model.EventListenerTypeDiscordGuildAuditLogEntryCreate:
		base := map[string]any{
			"id":          flow.TestSnowflake().String(),
			"guild_id":    s.guildID.String(),
			"user_id":     s.userID.String(),
			"action_type": discord.GuildUpdate,
		}
		return decodeEvent[gateway.GuildAuditLogEntryCreateEvent](base, test.Event)
	}
	return nil, testErrorf("test runs don't support %s event listeners yet", eventType)
}

// decodeEvent decodes base with override merged into it as a T.
func decodeEvent[T any](base map[string]any, override json.RawMessage) (*T, error) {
	var e T
	if err := mergeJSON(base, override, &e); err != nil {
		return nil, testErrorf("invalid event: %v", err)
	}
	return &e, nil
}

// mergeJSON merges override into base, both JSON objects, and decodes the
// result into target. Objects are merged, other values replaced.
func mergeJSON(base map[string]any, override json.RawMessage, target any) error {
	b, err := json.Marshal(base)
	if err != nil {
		return err
	}

	if len(override) > 0 && string(override) != "null" {
		var merged map[string]any
		if err := json.Unmarshal(b, &merged); err != nil {
			return err
		}
		var o map[string]any
		if err := json.Unmarshal(override, &o); err != nil {
			return errors.New("must be a JSON object")
		}
		mergeMaps(merged, o)
		if b, err = json.Marshal(merged); err != nil {
			return err
		}
	}

	return json.Unmarshal(b, target)
}

func mergeMaps(dst map[string]any, src map[string]any) {
	for k, v := range src {
		if vm, ok := v.(map[string]any); ok {
			if dm, ok := dst[k].(map[string]any); ok {
				mergeMaps(dm, vm)
				continue
			}
		}
		dst[k] = v
	}
}

// testComponent is a button or select menu of a sent message.
type testComponent struct {
	CustomID    string
	Type        int
	Label       string
	Placeholder string
}

// findComponent finds the button or select menu with the label, placeholder
// or custom ID in the messages the flows sent, newest first.
func (s *testSession) findComponent(target string, isSelect bool) (discord.Message, testComponent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := len(s.messages) - 1; i >= 0; i-- {
		for _, c := range messageComponents(s.messages[i]) {
			if isSelect != (c.Type != int(discord.ButtonComponentType)) {
				continue
			}
			if c.CustomID == target || (c.Label != "" && c.Label == target) || (c.Placeholder != "" && c.Placeholder == target) {
				return s.messages[i], c, true
			}
		}
	}
	return discord.Message{}, testComponent{}, false
}

func (s *testSession) availableComponents() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res []string
	for i := len(s.messages) - 1; i >= 0; i-- {
		for _, c := range messageComponents(s.messages[i]) {
			if c.Type == int(discord.ButtonComponentType) {
				res = append(res, fmt.Sprintf("button %q", c.Label))
			} else {
				res = append(res, fmt.Sprintf("select menu %q (%s)", c.Placeholder, c.CustomID))
			}
		}
	}
	return res
}

// messageComponents returns the buttons and select menus of a message that
// run flows. Link buttons don't.
func messageComponents(msg discord.Message) []testComponent {
	b, err := json.Marshal(msg.Components)
	if err != nil {
		return nil
	}
	var tree any
	if err := json.Unmarshal(b, &tree); err != nil {
		return nil
	}

	var res []testComponent
	var walk func(v any)
	walk = func(v any) {
		switch v := v.(type) {
		case []any:
			for _, item := range v {
				walk(item)
			}
		case map[string]any:
			typ, _ := v["type"].(float64)
			customID, _ := v["custom_id"].(string)
			if customID != "" && discord.ComponentType(typ) != discord.TextInputComponentType {
				label, _ := v["label"].(string)
				placeholder, _ := v["placeholder"].(string)
				res = append(res, testComponent{CustomID: customID, Type: int(typ), Label: label, Placeholder: placeholder})
			}
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(tree)
	return res
}

func (s *testSession) addMessage(msg discord.Message) discord.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg.ID = discord.MessageID(flow.TestSnowflake())
	msg.Author = s.session.Ready().User
	msg.Timestamp = discord.NewTimestamp(time.Now())
	s.messages = append(s.messages, msg)
	return msg
}

// setOriginal records the message an interaction responded with, which
// edits of its response change.
func (s *testSession) setOriginal(interactionID discord.InteractionID, messageID discord.MessageID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.originals[interactionID] = messageID
}

// editMessage applies an edit to a message and stores it, so its new
// components can be used.
func (s *testSession) editMessage(msg discord.Message, content option.NullableString, embeds *[]discord.Embed, components *discord.TopLevelComponents) discord.Message {
	if content != nil {
		msg.Content = content.Val
	}
	if embeds != nil {
		msg.Embeds = *embeds
	}
	if components != nil {
		msg.Components = *components
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if i := s.messageIndex(msg.ID); i != -1 {
		s.messages[i] = msg
	} else {
		s.messages = append(s.messages, msg)
	}
	return msg
}

// editOriginal edits the response of the current interaction: the message it
// created, the message of a clicked component, or the message a defer was
// waiting for.
func (s *testSession) editOriginal(content option.NullableString, embeds *[]discord.Embed, components *discord.TopLevelComponents) discord.Message {
	s.mu.Lock()
	interaction := s.current
	var original discord.Message
	found := false
	if interaction != nil {
		if id, ok := s.originals[interaction.ID]; ok {
			i := s.messageIndex(id)
			if i != -1 {
				original, found = s.messages[i], true
			}
		}
		if !found && interaction.Message != nil {
			original, found = *interaction.Message, true
		}
	}
	s.mu.Unlock()

	if !found {
		var msg discord.Message
		if interaction != nil {
			msg.ChannelID, msg.GuildID = interaction.ChannelID, interaction.GuildID
		}
		original = s.addMessage(msg)
		if interaction != nil {
			s.setOriginal(interaction.ID, original.ID)
		}
	}
	return s.editMessage(original, content, embeds, components)
}

// messageIndex is the index of a sent message, -1 if there's none. The lock
// has to be held.
func (s *testSession) messageIndex(id discord.MessageID) int {
	return slices.IndexFunc(s.messages, func(m discord.Message) bool { return m.ID == id })
}

func (s *testSession) fakeMessage(id discord.MessageID) (discord.Message, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.messageIndex(id)
	if i == -1 {
		return discord.Message{}, false
	}
	return s.messages[i], true
}

func (s *testSession) addChannel(channel discord.Channel) discord.Channel {
	s.mu.Lock()
	defer s.mu.Unlock()
	channel.ID = discord.ChannelID(flow.TestSnowflake())
	s.channels[channel.ID] = channel
	return channel
}

func (s *testSession) fakeChannel(id discord.ChannelID) (discord.Channel, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	channel, ok := s.channels[id]
	return channel, ok
}

func (s *testSession) guildOfChannel(id discord.ChannelID) discord.GuildID {
	if channel, ok := s.fakeChannel(id); ok {
		return channel.GuildID
	}
	if channel, err := s.session.Channel(id); err == nil {
		return channel.GuildID
	}
	return 0
}

func (s *testSession) linkTemplate(messageID discord.MessageID, templateID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[messageID] = templateID
}

func (s *testSession) setModal(data *api.InteractionResponseData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modal = data
}

// markResponded records a response to an interaction, false if it had one.
func (s *testSession) markResponded(id discord.InteractionID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.respondedTo[id] {
		return false
	}
	s.respondedTo[id] = true
	return true
}

func (s *testSession) responded(id discord.InteractionID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.respondedTo[id]
}

func (s *testSession) currentInteraction() *discord.InteractionEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

func (s *testSession) addResumePoint(rp flow.ResumePoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	point := testResumePoint{rp: rp, root: s.root}
	if rp.Type == flow.ResumePointTypeTimer {
		s.timers = append(s.timers, point)
		return
	}
	s.resumePoints[rp.ID] = point
}

func (s *testSession) variable(id string, scope null.String) (thing.Thing, bool, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.variables[testVariableKey{id: id, scope: scope}]
	return v.value, v.deleted, ok
}

func (s *testSession) setVariable(id string, scope null.String, value thing.Thing, deleted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variables[testVariableKey{id: id, scope: scope}] = testVariableValue{value: value, deleted: deleted}
}
