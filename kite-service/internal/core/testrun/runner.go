package testrun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"runtime/debug"
	"slices"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
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

// ErrAppOffline means the app has no gateway on this cluster, so there's
// no Discord session to read from.
var ErrAppOffline = errors.New("app is offline")

// TestError is a problem with the test itself, like an argument the
// command doesn't have.
type TestError struct {
	Message string
}

func (e *TestError) Error() string {
	return e.Message
}

func testErrorf(format string, args ...any) error {
	return &TestError{Message: fmt.Sprintf(format, args...)}
}

type Kind string

const (
	KindCommand       Kind = "command"
	KindEventListener Kind = "event_listener"
	KindMessage       Kind = "message"
)

// Test is a test run of a flow: what triggers it, the mocks of its blocks
// and what the user does after, like clicking a button the flow sent.
type Test struct {
	Kind Kind
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
	Webhook *Webhook

	Mocks map[string]flow.TestMock
	// Seed makes the test pick the same at random each time it runs, see
	// flow.TestRandom. 0 picks a seed.
	Seed uint64
	// Values of stored variables before the test, in place of the stored
	// ones.
	Variables []Variable
	Then      []Action
}

// Variable is the value of a stored variable when the test starts.
type Variable struct {
	VariableID string
	// The scope, for variables with a value per scope.
	Scope null.String
	// The value, or thing.Null for no value.
	Value thing.Thing
}

type Webhook struct {
	Body    string
	Headers map[string]string
	Query   map[string]string
}

// Action is something the user does after the flow ran: clicking a
// button or using a select menu of a message the flow sent, or submitting the
// modal it showed.
type Action struct {
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

type Result struct {
	Executions []Execution
	// What the test would have cost, in credits.
	Credits int
	// Mocked blocks that never ran, likely wrong IDs.
	UnusedMocks []string
}

// Execution is one run of a flow in a test, like the command or a
// click on a button it sent.
type Execution struct {
	Trigger   string
	Steps     []flow.TestStep
	Truncated bool
	// Problems of the execution as a whole, like an interaction nothing
	// responded to.
	Notes []string
	// The error that ended the execution, if no Error Handler caught it.
	Error string
}

// Runner runs test runs of flows for the API, see Test. Flows run with the
// engine's env, sessions come from the gateway manager.
type Runner struct {
	env      engine.Env
	sessions engine.SessionProvider
}

func NewRunner(env engine.Env, sessions engine.SessionProvider) *Runner {
	return &Runner{
		env:      env,
		sessions: sessions,
	}
}

// Run runs a test. Problems with the test are returned as TestError,
// problems of the flow are part of the result.
func (r *Runner) Run(ctx context.Context, appID string, test Test) (*Result, error) {
	session, err := r.sessions.AppSession(ctx, appID)
	if err != nil || session == nil {
		return nil, ErrAppOffline
	}

	if len(test.Then) > maxTestActions {
		return nil, testErrorf("a test can have at most %d steps after the trigger", maxTestActions)
	}

	ctx, cancel := context.WithTimeout(ctx, testRunTimeout)
	defer cancel()

	var root *flow.CompiledFlowNode
	switch test.Kind {
	case KindCommand:
		root, err = flow.CompileCommand(test.Flow)
	case KindEventListener:
		root, err = flow.CompileEventListener(test.Flow)
	case KindMessage:
	default:
		return nil, testErrorf("unknown kind %q", test.Kind)
	}
	if err != nil {
		return nil, testErrorf("the flow is invalid: %v", err)
	}

	seed := test.Seed
	if seed == 0 {
		seed = rand.Uint64()
	}
	s := newTestSession(r.env, appID, session, test.Mocks, flow.NewTestRandom(seed))
	if err := s.setVariables(ctx, test.Variables); err != nil {
		return nil, err
	}
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
			var testErr *TestError
			if !errors.As(err, &testErr) {
				return nil, err
			}
			// Later steps can depend on what earlier ones did, so the steps
			// that ran are still worth seeing.
			s.result.Executions = append(s.result.Executions, Execution{
				Trigger: fmt.Sprintf("Step %d", i+1),
				Error:   testErr.Message,
			})
			break
		}
	}

	if ctx.Err() != nil {
		s.result.Executions = append(s.result.Executions, Execution{
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
	random  *flow.TestRandom
	env     engine.Env
	appID   string
	session *state.State
	botID   discord.UserID
	mocks   map[string]flow.TestMock

	// Where the test runs. guildID is 0 in DMs.
	guildID   discord.GuildID
	channelID discord.ChannelID
	userID    discord.UserID

	result    Result
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
	value     thing.Thing
	deleted   bool
	updatedAt time.Time
}

func newTestSession(env engine.Env, appID string, session *state.State, mocks map[string]flow.TestMock, random *flow.TestRandom) *testSession {
	return &testSession{
		random:    random,
		env:       env,
		appID:     appID,
		session:   session,
		botID:     session.Ready().User.ID,
		mocks:     mocks,
		usedMocks: make(map[string]bool),
		real:      env.FlowProviders(appID, session),
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
func (s *testSession) setLocation(test Test, root *flow.CompiledFlowNode) error {
	userID, err := discord.ParseSnowflake(test.UserID)
	if err != nil || !userID.IsValid() {
		return testErrorf("user_id must be the ID of a Discord user")
	}
	s.userID = discord.UserID(userID)

	dm := test.DM
	// Scheduled and webhook runs have no user or place, but blocks still
	// default to the server and channel of the test.
	nowhere := false
	if test.Kind == KindEventListener {
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
func (s *testSession) start(ctx context.Context, test Test, root *flow.CompiledFlowNode) error {
	switch test.Kind {
	case KindCommand:
		event, err := s.commandEvent(root, test)
		if err != nil {
			return err
		}
		trigger := root.CommandName()
		if root.CommandType() == discord.ChatInputCommand {
			trigger = "/" + trigger
		}
		s.execute(ctx, trigger, root, root, event, nil, nil)
		return nil
	case KindEventListener:
		event, err := s.listenerEvent(root, test)
		if err != nil {
			return err
		}
		t := model.EventListenerType(root.EventListenerType())
		listener := &model.EventListener{Type: t, Source: model.EventSourceForType(t)}
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

	action := Action{Click: test.FlowSourceID}
	if len(test.Values) > 0 {
		action = Action{Select: test.FlowSourceID, Values: test.Values}
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
	listener *model.EventListener,
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
		event, state, err := engine.ResumeTriggerEvent(timer.rp.State, timer.rp.InteractionToken)
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
	listener *model.EventListener,
	afterSleep bool,
) {
	exec := Execution{Trigger: trigger}
	defer func() {
		s.result.Executions = append(s.result.Executions, exec)
	}()

	data := engine.FlowContextData(event)
	if listener != nil {
		if !engine.ShouldHandleEvent(listener, event, s.botID) {
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

	run := flow.NewTestRun(s.mocks, s.random)
	fCtx := s.env.FlowContextWithProviders(ctx, s.appID, s.session, data, s.testProviders(run), state, node)
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

	// The request is answered when the flow ends or waits, like in
	// HandleWebhookRun, so responses after a wait fail like they would.
	if e, ok := event.(*webhook.Event); ok {
		e.Responder.Close()
	}

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
