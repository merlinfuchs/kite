package engine

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

const (
	testRunGuild   = discord.GuildID(100)
	testRunBot     = discord.UserID(1)
	testRunUser    = discord.UserID(2)
	testRunChannel = discord.ChannelID(300)
	// @everyone can't send messages in it.
	testRunLockedChannel = discord.ChannelID(301)
	testRunLowRole       = discord.RoleID(400)
	testRunHighRole      = discord.RoleID(401)
	testRunBotRole       = discord.RoleID(402)
)

// testRunState is a session with everything the tests read cached, so they
// never call Discord.
func testRunState(t *testing.T) *state.State {
	t.Helper()

	s := state.New("Bot test")
	s.Session.Handler.Call(&gateway.ReadyEvent{User: discord.User{ID: testRunBot, Username: "kite", Bot: true}})

	guild := discord.Guild{ID: testRunGuild, Name: "Test", OwnerID: 999, SystemChannelID: testRunChannel}
	roles := []discord.Role{
		{ID: discord.RoleID(testRunGuild), Name: "@everyone", Permissions: discord.PermissionViewChannel | discord.PermissionSendMessages},
		{ID: testRunLowRole, Name: "Verified", Position: 1},
		{ID: testRunBotRole, Name: "Kite", Position: 2, Permissions: discord.PermissionManageRoles},
		{ID: testRunHighRole, Name: "Admin", Position: 3},
	}
	members := []discord.Member{
		{User: discord.User{ID: testRunBot, Username: "kite", Bot: true}, RoleIDs: []discord.RoleID{testRunBotRole}},
		{User: discord.User{ID: testRunUser, Username: "tester"}},
	}
	channels := []discord.Channel{
		{ID: testRunChannel, GuildID: testRunGuild, Type: discord.GuildText, Name: "general"},
		{ID: testRunLockedChannel, GuildID: testRunGuild, Type: discord.GuildText, Name: "announcements", Overwrites: []discord.Overwrite{
			{ID: discord.Snowflake(testRunGuild), Type: discord.OverwriteRole, Deny: discord.PermissionSendMessages},
		}},
	}

	if err := s.Cabinet.GuildSet(&guild, false); err != nil {
		t.Fatal(err)
	}
	for i := range roles {
		if err := s.Cabinet.RoleSet(testRunGuild, &roles[i], false); err != nil {
			t.Fatal(err)
		}
	}
	for i := range members {
		if err := s.Cabinet.MemberSet(testRunGuild, &members[i], false); err != nil {
			t.Fatal(err)
		}
	}
	for i := range channels {
		if err := s.Cabinet.ChannelSet(&channels[i], false); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

type testRunSessions struct {
	session *state.State
}

func (s testRunSessions) AppIDs() []string { return []string{"app"} }

func (s testRunSessions) AppSession(ctx context.Context, appID string) (*state.State, error) {
	return s.session, nil
}

func runTest(t *testing.T, test FlowTest) *FlowTestResult {
	t.Helper()

	e := NewEngine(Env{Config: EngineConfig{ClusterCount: 1}})
	r := e.FlowTestRunner(testRunSessions{session: testRunState(t)})

	if test.UserID == "" {
		test.UserID = testRunUser.String()
	}
	res, err := r.Run(context.Background(), "app", test)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func testFlow(t *testing.T, s string) flow.FlowData {
	t.Helper()

	var data flow.FlowData
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func findEffect(t *testing.T, exec FlowTestExecution, action string) flow.TestEffect {
	t.Helper()

	for _, step := range exec.Steps {
		for _, effect := range step.Effects {
			if effect.Action == action {
				return effect
			}
		}
	}
	t.Fatalf("no %s effect in %+v", action, exec.Steps)
	return flow.TestEffect{}
}

func responseContent(t *testing.T, effect flow.TestEffect) string {
	t.Helper()

	data, _ := effect.Data.(map[string]any)
	msg, ok := data["message"].(*api.InteractionResponseData)
	if !ok || msg.Content == nil {
		t.Fatalf("effect has no message: %+v", effect.Data)
	}
	return msg.Content.Val
}

const greetFlow = `{
	"nodes": [
		{"id": "entry", "type": "entry_command", "data": {"name": "greet", "description": "Greets"}},
		{"id": "arg", "type": "option_command_argument", "data": {"name": "text", "description": "Text", "command_argument_type": "string", "command_argument_required": true}},
		{"id": "reply", "type": "action_response_create", "data": {"message_data": {"content": "Hi {{user.username}}: {{arg('text')}}"}}}
	],
	"edges": [
		{"id": "1", "source": "arg", "target": "entry"},
		{"id": "2", "source": "entry", "target": "reply"}
	]
}`

func TestFlowTestCommand(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind:    FlowTestKindCommand,
		Flow:    testFlow(t, greetFlow),
		Options: map[string]json.RawMessage{"text": json.RawMessage(`"hello"`)},
	})

	if len(res.Executions) != 1 {
		t.Fatalf("got %d executions, want 1", len(res.Executions))
	}
	exec := res.Executions[0]
	if exec.Error != "" || len(exec.Notes) > 0 {
		t.Fatalf("execution failed: %+v", exec)
	}
	if got := responseContent(t, findEffect(t, exec, "respond")); got != "Hi tester: hello" {
		t.Errorf("response = %q", got)
	}
	if res.Credits != 1 {
		t.Errorf("credits = %d, want 1", res.Credits)
	}
}

func TestFlowTestCommandArguments(t *testing.T) {
	e := NewEngine(Env{Config: EngineConfig{ClusterCount: 1}})
	r := e.FlowTestRunner(testRunSessions{session: testRunState(t)})

	for name, options := range map[string]map[string]json.RawMessage{
		"missing": {},
		"unknown": {"text": json.RawMessage(`"hi"`), "other": json.RawMessage(`1`)},
		"wrong":   {"text": json.RawMessage(`5`)},
	} {
		_, err := r.Run(context.Background(), "app", FlowTest{
			Kind:    FlowTestKindCommand,
			Flow:    testFlow(t, greetFlow),
			UserID:  testRunUser.String(),
			Options: options,
		})
		var testErr *FlowTestError
		if !errors.As(err, &testErr) {
			t.Errorf("%s: got %v, want a FlowTestError", name, err)
		}
	}
}

// A click on a button the flow sent continues the flow, and a role the bot
// can't assign is warned about.
func TestFlowTestClickButton(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_command", "data": {"name": "verify", "description": "Verify"}},
				{"id": "ask", "type": "action_response_create", "data": {"message_data": {"content": "Sure?", "components": [
					{"type": 1, "components": [{"id": 1, "type": 2, "style": 1, "label": "Yes"}]}
				]}}},
				{"id": "low", "type": "action_member_role_add", "data": {"user_target": "{{user.id}}", "role_target": "400"}},
				{"id": "high", "type": "action_member_role_add", "data": {"user_target": "{{user.id}}", "role_target": "401"}}
			],
			"edges": [
				{"id": "1", "source": "entry", "target": "ask"},
				{"id": "2", "source": "ask", "target": "low", "sourceHandle": "component_1"},
				{"id": "3", "source": "low", "target": "high"}
			]
		}`),
		Then: []FlowTestAction{{Click: "Yes"}},
	})

	if len(res.Executions) != 2 {
		t.Fatalf("got %d executions, want 2: %+v", len(res.Executions), res.Executions)
	}
	click := res.Executions[1]
	if click.Error != "" {
		t.Fatalf("click failed: %s", click.Error)
	}

	var warnings []string
	var paths []string
	for _, step := range click.Steps {
		for _, effect := range step.Effects {
			if effect.Action == "discord_api_request" {
				paths = append(paths, effect.Data.(map[string]any)["path"].(string))
				warnings = append(warnings, effect.Warnings...)
			}
		}
	}
	if want := []string{"/guilds/100/members/2/roles/400", "/guilds/100/members/2/roles/401"}; strings.Join(paths, " ") != strings.Join(want, " ") {
		t.Errorf("requests = %v, want %v", paths, want)
	}
	if len(warnings) != 1 || !strings.Contains(warnings[0], "Admin isn't below the bot's highest role") {
		t.Errorf("warnings = %q", warnings)
	}
	// Components nothing responds to are acknowledged.
	findEffect(t, click, "respond")
}

func TestFlowTestUnknownButton(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, greetFlow),
		Options: map[string]json.RawMessage{
			"text": json.RawMessage(`"hi"`),
		},
		Then: []FlowTestAction{{Click: "Yes"}},
	})

	last := res.Executions[len(res.Executions)-1]
	if !strings.Contains(last.Error, `no button "Yes"`) {
		t.Errorf("error = %q", last.Error)
	}
}

const weatherFlow = `{
	"nodes": [
		{"id": "entry", "type": "entry_command", "data": {"name": "weather", "description": "Weather"}},
		{"id": "handler", "type": "control_error_handler", "data": {}},
		{"id": "http", "type": "action_http_request", "data": {"http_request_data": {"url": "https://example.com", "method": "GET"}}},
		{"id": "reply", "type": "action_response_create", "data": {"message_data": {"content": "{{result('http').status_code}} {{result('http').data().temp}}"}}},
		{"id": "failed", "type": "action_response_create", "data": {"message_data": {"content": "failed: {{result('handler')}}"}}}
	],
	"edges": [
		{"id": "1", "source": "entry", "target": "handler"},
		{"id": "2", "source": "handler", "target": "http"},
		{"id": "3", "source": "http", "target": "reply"},
		{"id": "4", "source": "handler", "target": "failed", "sourceHandle": "error"}
	]
}`

func TestFlowTestExternalRequests(t *testing.T) {
	res := runTest(t, FlowTest{Kind: FlowTestKindCommand, Flow: testFlow(t, weatherFlow)})
	exec := res.Executions[0]
	for _, step := range exec.Steps {
		if step.NodeID == "http" && (len(step.Notes) == 0 || !strings.HasPrefix(step.Notes[0], "Not run")) {
			t.Errorf("request without a mock isn't marked as not run: %+v", step)
		}
	}

	res = runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, weatherFlow),
		Mocks: map[string]flow.TestMock{
			"http": {Result: json.RawMessage(`{"status_code": 200, "body": {"temp": 21}}`)},
		},
	})
	if got := responseContent(t, findEffect(t, res.Executions[0], "respond")); got != "200 21" {
		t.Errorf("response with mocked result = %q", got)
	}

	res = runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, weatherFlow),
		Mocks: map[string]flow.TestMock{
			"http":    {Error: "connection refused"},
			"missing": {Result: json.RawMessage(`1`)},
		},
	})
	if got := responseContent(t, findEffect(t, res.Executions[0], "respond")); !strings.Contains(got, "connection refused") {
		t.Errorf("response with mocked error = %q", got)
	}
	if len(res.UnusedMocks) != 1 || res.UnusedMocks[0] != "missing" {
		t.Errorf("unused mocks = %v", res.UnusedMocks)
	}
}

func TestFlowTestPermissionWarning(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_command", "data": {"name": "announce", "description": "Announce"}},
				{"id": "send", "type": "action_message_create", "data": {"channel_target": "301", "message_data": {"content": "News"}}}
			],
			"edges": [{"id": "1", "source": "entry", "target": "send"}]
		}`),
	})

	exec := res.Executions[0]
	effect := findEffect(t, exec, "create_message")
	if len(effect.Warnings) != 1 || !strings.Contains(effect.Warnings[0], "lacks Send Messages in channel announcements") {
		t.Errorf("warnings = %q", effect.Warnings)
	}
	if len(exec.Notes) != 1 || !strings.Contains(exec.Notes[0], "Nothing responded") {
		t.Errorf("notes = %q", exec.Notes)
	}
}

func TestFlowTestModal(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_command", "data": {"name": "report", "description": "Report"}},
				{"id": "modal", "type": "suspend_response_modal", "data": {"modal_data": {"title": "Report", "components": [
					{"type": "label", "label": "Reason", "components": [{"type": "text_input", "custom_id": "reason", "style": 1}]}
				]}}},
				{"id": "reply", "type": "action_response_create", "data": {"message_data": {"content": "Reported: {{input('reason')}}"}}}
			],
			"edges": [
				{"id": "1", "source": "entry", "target": "modal"},
				{"id": "2", "source": "modal", "target": "reply"}
			]
		}`),
		Then: []FlowTestAction{{Submit: map[string]json.RawMessage{"Reason": json.RawMessage(`"spam"`)}}},
	})

	if len(res.Executions) != 2 {
		t.Fatalf("got %d executions, want 2: %+v", len(res.Executions), res.Executions)
	}
	if res.Executions[1].Error != "" {
		t.Fatalf("submit failed: %s", res.Executions[1].Error)
	}
	if got := responseContent(t, findEffect(t, res.Executions[1], "respond")); got != "Reported: spam" {
		t.Errorf("response = %q", got)
	}
}

// Variables changed by a test can be read by later blocks, but aren't saved.
func TestFlowTestVariables(t *testing.T) {
	e := NewEngine(Env{
		Config:        EngineConfig{ClusterCount: 1},
		VariableStore: &fakeVariableStore{variables: []*model.Variable{{ID: "count"}}},
	})
	r := e.FlowTestRunner(testRunSessions{session: testRunState(t)})

	res, err := r.Run(context.Background(), "app", FlowTest{
		Kind:   FlowTestKindCommand,
		UserID: testRunUser.String(),
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_command", "data": {"name": "count", "description": "Count"}},
				{"id": "set", "type": "action_variable_set", "data": {"variable_id": "count", "variable_value": "5"}},
				{"id": "get", "type": "action_variable_get", "data": {"variable_id": "count"}},
				{"id": "reply", "type": "action_response_create", "data": {"message_data": {"content": "{{result('get')}}"}}}
			],
			"edges": [
				{"id": "1", "source": "entry", "target": "set"},
				{"id": "2", "source": "set", "target": "get"},
				{"id": "3", "source": "get", "target": "reply"}
			]
		}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	exec := res.Executions[0]
	if exec.Error != "" {
		t.Fatalf("execution failed: %s", exec.Error)
	}
	findEffect(t, exec, "set_variable")
	if got := responseContent(t, findEffect(t, exec, "respond")); got != "5" {
		t.Errorf("response = %q", got)
	}
}

func TestFlowTestEventFilter(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindEventListener,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_event", "data": {"event_type": "message_create", "description": "Messages"}},
				{"id": "send", "type": "action_message_create", "data": {"channel_target": "{{channel.id}}", "message_data": {"content": "You said {{message.content}}"}}}
			],
			"edges": [{"id": "1", "source": "entry", "target": "send"}]
		}`),
		Event: json.RawMessage(`{"content": "hello"}`),
	})

	exec := res.Executions[0]
	if exec.Error != "" {
		t.Fatalf("execution failed: %s", exec.Error)
	}
	effect := findEffect(t, exec, "create_message")
	data := effect.Data.(map[string]any)
	if got := data["message"].(api.SendMessageData).Content; got != "You said hello" {
		t.Errorf("message = %q", got)
	}
	if data["channel_id"] != testRunChannel {
		t.Errorf("channel = %v", data["channel_id"])
	}
}

// Long waits are skipped, the flow continues in a second execution like it
// would after the wait.
func TestFlowTestDurableSleep(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_command", "data": {"name": "remind", "description": "Remind"}},
				{"id": "wait", "type": "control_sleep", "data": {"sleep_duration_seconds": "3600"}},
				{"id": "reply", "type": "action_response_create", "data": {"message_data": {"content": "Reminder"}}}
			],
			"edges": [
				{"id": "1", "source": "entry", "target": "wait"},
				{"id": "2", "source": "wait", "target": "reply"}
			]
		}`),
	})

	if len(res.Executions) != 2 {
		t.Fatalf("got %d executions, want 2: %+v", len(res.Executions), res.Executions)
	}
	if !strings.HasPrefix(res.Executions[1].Trigger, "After waiting") {
		t.Errorf("trigger = %q", res.Executions[1].Trigger)
	}
	// The interaction was deferred before the wait, so the reply is a
	// follow-up.
	findEffect(t, res.Executions[0], "respond")
	findEffect(t, res.Executions[1], "create_followup")
}

func TestFlowTestWebhook(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindEventListener,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_event", "data": {"event_type": "webhook", "description": "Webhook"}},
				{"id": "verify", "type": "control_webhook_verify", "data": {"webhook_verify_preset": "github", "webhook_verify_key": "{{secrets.github}}"}},
				{"id": "ok", "type": "action_webhook_response_create", "data": {"webhook_response_status": "200", "webhook_response_body": "{{webhook.data.action}}"}},
				{"id": "denied", "type": "action_webhook_response_create", "data": {"webhook_response_status": "401", "webhook_response_body": "no"}}
			],
			"edges": [
				{"id": "1", "source": "entry", "target": "verify"},
				{"id": "2", "source": "verify", "target": "ok", "sourceHandle": "valid"},
				{"id": "3", "source": "verify", "target": "denied", "sourceHandle": "invalid"}
			]
		}`),
		Webhook: &FlowTestWebhook{Body: `{"action": "opened"}`},
	})

	effect := findEffect(t, res.Executions[0], "respond_webhook")
	if got := effect.Data.(map[string]any); got["status"] != 200 || got["body"] != "opened" {
		t.Errorf("response = %+v", got)
	}
}

// The request is answered when the flow waits, so responding after a long
// wait fails like it would.
func TestFlowTestWebhookResponseAfterWait(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindEventListener,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_event", "data": {"event_type": "webhook", "description": "Webhook"}},
				{"id": "wait", "type": "control_sleep", "data": {"sleep_duration_seconds": "60"}},
				{"id": "ok", "type": "action_webhook_response_create", "data": {"webhook_response_status": "200", "webhook_response_body": "ok"}}
			],
			"edges": [
				{"id": "1", "source": "entry", "target": "wait"},
				{"id": "2", "source": "wait", "target": "ok"}
			]
		}`),
		Webhook: &FlowTestWebhook{Body: "{}"},
	})

	if len(res.Executions) != 2 {
		t.Fatalf("got %d executions, want 2: %+v", len(res.Executions), res.Executions)
	}
	for _, step := range res.Executions[1].Steps {
		for _, effect := range step.Effects {
			if effect.Action == "respond_webhook" {
				t.Errorf("responded after the wait: %+v", effect)
			}
		}
	}
}

// Endpoints only flows can use, like the webhooks of a server with their
// tokens, aren't read in test runs, whose token may not read them.
func TestFlowTestFlowOnlyRead(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_command", "data": {"name": "hooks", "description": "Hooks"}},
				{"id": "get", "type": "action_discord_api_request", "data": {"discord_api_request_data": {"operation": "get_guild_webhooks", "path_params": [{"key": "guild_id", "value": "100"}]}}}
			],
			"edges": [{"id": "1", "source": "entry", "target": "get"}]
		}`),
	})

	exec := res.Executions[0]
	if exec.Error != "" {
		t.Fatalf("execution failed: %s", exec.Error)
	}
	effect := findEffect(t, exec, "discord_api_request")
	if len(effect.Warnings) != 1 || !strings.HasPrefix(effect.Warnings[0], "Not sent") {
		t.Errorf("warnings = %q", effect.Warnings)
	}
}

func TestFlowTestDeletedMessage(t *testing.T) {
	res := runTest(t, FlowTest{
		Kind: FlowTestKindCommand,
		Flow: testFlow(t, `{
			"nodes": [
				{"id": "entry", "type": "entry_command", "data": {"name": "confirm", "description": "Confirm"}},
				{"id": "ask", "type": "action_response_create", "data": {"message_data": {"content": "Sure?", "components": [
					{"type": 1, "components": [{"id": 1, "type": 2, "style": 1, "label": "Yes"}]}
				]}}},
				{"id": "delete", "type": "action_response_delete", "data": {}}
			],
			"edges": [
				{"id": "1", "source": "entry", "target": "ask"},
				{"id": "2", "source": "ask", "target": "delete", "sourceHandle": "component_1"}
			]
		}`),
		Then: []FlowTestAction{{Click: "Yes"}, {Click: "Yes"}},
	})

	last := res.Executions[len(res.Executions)-1]
	if !strings.Contains(last.Error, `no button "Yes"`) {
		t.Errorf("clicked a deleted message: %+v", res.Executions)
	}
}

func TestFlowTestMessageButton(t *testing.T) {
	panel := &model.Message{
		ID:    "panel",
		AppID: "app",
		Data: message.MessageData{
			Content: "Pick",
			Components: []message.ComponentData{{Type: 1, Components: []message.ComponentData{
				{ID: 1, Type: 2, Style: 1, Label: "Ping", FlowSourceID: "ping"},
			}}},
		},
		FlowSources: map[string]flow.FlowData{"ping": testFlow(t, `{
				"nodes": [
					{"id": "entry", "type": "entry_component_button", "data": {}},
					{"id": "reply", "type": "action_response_create", "data": {"message_data": {"content": "Pong"}}}
				],
				"edges": [{"id": "1", "source": "entry", "target": "reply"}]
			}`)},
	}

	res := runTest(t, FlowTest{
		Kind:         FlowTestKindMessage,
		Message:      panel,
		FlowSourceID: "ping",
	})
	if got := responseContent(t, findEffect(t, res.Executions[0], "respond")); got != "Pong" {
		t.Errorf("response = %q", got)
	}
}

func TestFlowTestAICheck(t *testing.T) {
	flowData := testFlow(t, `{
		"nodes": [
			{"id": "entry", "type": "entry_command", "data": {"name": "check", "description": "Check"}},
			{"id": "ai", "type": "control_ai_check", "data": {"ai_decision_data": {"question": "Is it spam?", "input": "buy now"}}},
			{"id": "yes", "type": "action_response_create", "data": {"message_data": {"content": "spam"}}},
			{"id": "no", "type": "action_response_create", "data": {"message_data": {"content": "fine"}}}
		],
		"edges": [
			{"id": "1", "source": "entry", "target": "ai"},
			{"id": "2", "source": "ai", "target": "yes", "sourceHandle": "yes"},
			{"id": "3", "source": "ai", "target": "no", "sourceHandle": "no"}
		]
	}`)

	res := runTest(t, FlowTest{Kind: FlowTestKindCommand, Flow: flowData})
	if got := responseContent(t, findEffect(t, res.Executions[0], "respond")); got != "fine" {
		t.Errorf("response without a mock = %q", got)
	}

	res = runTest(t, FlowTest{
		Kind:  FlowTestKindCommand,
		Flow:  flowData,
		Mocks: map[string]flow.TestMock{"ai": {Result: json.RawMessage(`"yes"`)}},
	})
	if got := responseContent(t, findEffect(t, res.Executions[0], "respond")); got != "spam" {
		t.Errorf("response with a mock = %q", got)
	}
}
