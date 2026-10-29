package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
)

// commandSession is testSession that NewContextFromInteraction can read the
// app user from.
func commandSession(t *testing.T) *state.State {
	t.Helper()

	session := state.New("Bot test")
	session.Cabinet = testSession(t).Cabinet
	return session
}

func TestCommandTargetUser(t *testing.T) {
	data := &discord.CommandInteraction{TargetID: 5}
	data.Resolved.Users = map[discord.UserID]discord.User{5: {ID: 5, Username: "jack"}}

	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: data,
	}, commandSession(t))

	if got := evalString(t, c, "{{command.target.id}}|{{context.target.id}}"); got != "5|5" {
		t.Errorf("got %q", got)
	}
}

func TestCommandTargetMemberUsesGuildRoles(t *testing.T) {
	data := &discord.CommandInteraction{TargetID: 5}
	data.Resolved.Users = map[discord.UserID]discord.User{5: {ID: 5}}
	data.Resolved.Members = map[discord.UserID]discord.Member{5: {RoleIDs: []discord.RoleID{11, 13}}}

	c := NewContextFromInteraction(&discord.InteractionEvent{
		GuildID: testGuildID,
		Member:  &discord.Member{User: discord.User{ID: testUserID}},
		Data:    data,
	}, commandSession(t))

	if got := evalString(t, c, "{{context.target.id}} {{context.target.top_role.name}}"); got != "5 Mod" {
		t.Errorf("got %q", got)
	}
}

func TestCommandTargetMessage(t *testing.T) {
	data := &discord.CommandInteraction{TargetID: 9}
	data.Resolved.Messages = map[discord.MessageID]discord.Message{9: {ID: 9, Content: "hello"}}

	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: data,
	}, commandSession(t))

	if got := evalString(t, c, "{{context.target.id}} {{context.target.content}}"); got != "9 hello" {
		t.Errorf("got %q", got)
	}
}

func TestCommandTargetSlashCommand(t *testing.T) {
	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.CommandInteraction{ID: 3, Name: "ping"},
	}, commandSession(t))

	if got := evalString(t, c, "{{command.id}} {{command.target == nil}} {{context.target == nil}}"); got != "3 true true" {
		t.Errorf("got %q", got)
	}
}

// Buttons and modals have no command, so command is nil like any unknown name
// instead of a CommandEnv that is nil but isn't nil in expressions.
func TestCommandOnlyForCommands(t *testing.T) {
	c := NewContextFromInteraction(&discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.ButtonInteraction{CustomID: "button"},
	}, commandSession(t))

	if _, ok := c.Env["command"]; ok {
		t.Error("command is set for a button")
	}
	if _, ok := c.Env["context"]; ok {
		t.Error("context is set for a button")
	}
	if got := evalString(t, c, "{{command == nil}}"); got != "true" {
		t.Errorf("got %q", got)
	}
}
