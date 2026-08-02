package engine

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// The command and listener lookup indexes replaced linear scans over every
// command/listener on the dispatch path. These tests pin the behaviour that
// dispatch depends on, in particular that the indexes survive renames, type
// changes, and deletions -- a stale index silently routes an interaction to
// the wrong flow or drops it.

func testCommand(id, name string) (string, *Command) {
	return id, &Command{cmd: &model.Command{ID: id, Name: name}}
}

func testListener(id string, source model.EventSource, eventType model.EventListenerType) (string, *EventListener) {
	return id, &EventListener{
		listener: &model.EventListener{ID: id, Source: source, Type: eventType},
	}
}

func TestCommandIndexLookupByName(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddCommand(testCommand("cmd-1", "ping"))
	app.AddCommand(testCommand("cmd-2", "config set"))

	if got := app.commandsByName["ping"]; got == nil || got.cmd.ID != "cmd-1" {
		t.Errorf("lookup of %q did not resolve to cmd-1", "ping")
	}
	if got := app.commandsByName["config set"]; got == nil || got.cmd.ID != "cmd-2" {
		t.Errorf("lookup of %q did not resolve to cmd-2", "config set")
	}
	if got := app.commandsByName["missing"]; got != nil {
		t.Errorf("lookup of unknown name resolved to %v, want nil", got.cmd.ID)
	}
}

// Renaming a command must not leave the old name pointing at it, or the
// command would keep firing under a name Discord no longer knows about.
func TestCommandIndexDropsOldNameOnRename(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddCommand(testCommand("cmd-1", "oldname"))
	app.AddCommand(testCommand("cmd-1", "newname"))

	if got := app.commandsByName["oldname"]; got != nil {
		t.Error("old name still resolves after rename")
	}
	if got := app.commandsByName["newname"]; got == nil {
		t.Fatal("new name does not resolve after rename")
	}
	if len(app.commands) != 1 {
		t.Errorf("commands = %d entries, want 1 after rename of same ID", len(app.commands))
	}
}

func TestCommandIndexRebuiltOnRemoval(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddCommand(testCommand("cmd-1", "keep"))
	app.AddCommand(testCommand("cmd-2", "drop"))

	app.RemoveDanglingCommands(util.IDSet([]string{"cmd-1"}))

	if got := app.commandsByName["drop"]; got != nil {
		t.Error("removed command still resolves by name")
	}
	if got := app.commandsByName["keep"]; got == nil {
		t.Error("surviving command no longer resolves by name")
	}
}

func TestListenerIndexGroupsByEventType(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddEventListener(testListener("l-1", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))
	app.AddEventListener(testListener("l-2", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))
	app.AddEventListener(testListener("l-3", model.EventSourceDiscord, model.EventListenerTypeDiscordGuildMemberAdd))

	if got := len(app.listenersByType[model.EventListenerTypeDiscordMessageCreate]); got != 2 {
		t.Errorf("message_create listeners = %d, want 2", got)
	}
	if got := len(app.listenersByType[model.EventListenerTypeDiscordGuildMemberAdd]); got != 1 {
		t.Errorf("guild_member_add listeners = %d, want 1", got)
	}
	if got := len(app.listenersByType[model.EventListenerTypeDiscordMessageDelete]); got != 0 {
		t.Errorf("message_delete listeners = %d, want 0", got)
	}
}

// Dispatch previously filtered on source at event time. The index does it at
// insert time instead, so a non-Discord listener must never end up in it.
func TestListenerIndexExcludesNonDiscordSources(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddEventListener(testListener("l-1", model.EventSource("webhook"), model.EventListenerTypeDiscordMessageCreate))

	if got := len(app.listenersByType[model.EventListenerTypeDiscordMessageCreate]); got != 0 {
		t.Errorf("non-Discord listener was indexed: got %d entries, want 0", got)
	}
	if len(app.listeners) != 1 {
		t.Errorf("listeners = %d, want the listener to still be registered", len(app.listeners))
	}
}

func TestListenerIndexRebuiltOnTypeChange(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddEventListener(testListener("l-1", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))
	app.AddEventListener(testListener("l-1", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageDelete))

	if got := len(app.listenersByType[model.EventListenerTypeDiscordMessageCreate]); got != 0 {
		t.Errorf("listener still indexed under old type: got %d entries, want 0", got)
	}
	if got := len(app.listenersByType[model.EventListenerTypeDiscordMessageDelete]); got != 1 {
		t.Errorf("listener not indexed under new type: got %d entries, want 1", got)
	}
}

func TestListenerIndexRebuiltOnRemoval(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddEventListener(testListener("l-1", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))
	app.AddEventListener(testListener("l-2", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))

	app.RemoveDanglingEventListeners(util.IDSet([]string{"l-1"}))

	if got := len(app.listenersByType[model.EventListenerTypeDiscordMessageCreate]); got != 1 {
		t.Errorf("message_create listeners = %d, want 1 after removal", got)
	}
}

// The dispatch path reads the index under a read lock, then releases it before
// spawning goroutines. That is only safe because rebuilds replace the slice
// rather than appending in place.
func TestListenerIndexRebuildDoesNotMutateExistingSlice(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddEventListener(testListener("l-1", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))

	snapshot := app.listenersByType[model.EventListenerTypeDiscordMessageCreate]

	app.AddEventListener(testListener("l-2", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))

	if len(snapshot) != 1 {
		t.Errorf("snapshot taken before rebuild changed length to %d, want 1", len(snapshot))
	}
	if snapshot[0].listener.ID != "l-1" {
		t.Errorf("snapshot contents changed: got %q, want l-1", snapshot[0].listener.ID)
	}
}

// Resume points are created by commands, event listeners and message
// instances alike, but dispatch only ever handled the command case. A button
// or modal inside an event listener flow resolved to nothing, so the
// interaction got no response and Discord showed "This interaction failed".

func TestResumeFlowTargetResolvesCommand(t *testing.T) {
	app := NewApp("app-1", Env{})
	id, command := testCommand("cmd-1", "ping")
	command.flow = &flow.CompiledFlowNode{}
	app.AddCommand(id, command)

	targetFlow := app.resumeFlowTarget(&model.ResumePoint{
		CommandID: null.NewString("cmd-1", true),
	})
	if targetFlow != command.flow {
		t.Error("command-owned resume point did not resolve to its flow")
	}
}

func TestResumeFlowTargetResolvesEventListener(t *testing.T) {
	app := NewApp("app-1", Env{})
	id, listener := testListener("lst-1", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate)
	listener.flow = &flow.CompiledFlowNode{}
	app.AddEventListener(id, listener)

	targetFlow := app.resumeFlowTarget(&model.ResumePoint{
		EventListenerID: null.NewString("lst-1", true),
	})
	if targetFlow != listener.flow {
		t.Error("event listener resume point did not resolve; buttons in listener flows are dead")
	}
}

// A resume point whose owner is gone (deleted or disabled) must resolve to
// nothing rather than panic.
func TestResumeFlowTargetUnknownOwner(t *testing.T) {
	app := NewApp("app-1", Env{})

	for name, resumePoint := range map[string]model.ResumePoint{
		"missing command":  {CommandID: null.NewString("nope", true)},
		"missing listener": {EventListenerID: null.NewString("nope", true)},
		"no owner":         {},
	} {
		if got := app.resumeFlowTarget(&resumePoint); got != nil {
			t.Errorf("%s: resolved to a flow, want nil", name)
		}
	}
}

// Links are stored on the resume point when the flow suspends, so attribution
// of logs and usage has to survive the round trip rather than be rebuilt.
func TestEntityLinksFromResumePoint(t *testing.T) {
	links := entityLinksFromResumePoint(&model.ResumePoint{
		MessageID:         null.NewString("msg-1", true),
		MessageInstanceID: null.NewInt(7, true),
		FlowSourceID:      null.NewString("src-1", true),
	})

	if links.MessageID.String != "msg-1" ||
		links.MessageInstanceID.Int64 != 7 ||
		links.FlowSourceID.String != "src-1" {
		t.Errorf("message instance attribution lost: %+v", links)
	}
}
