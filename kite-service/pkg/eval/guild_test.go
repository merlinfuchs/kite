package eval

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func TestGuildJoinExposesSystemChannel(t *testing.T) {
	event := &state.GuildJoinEvent{GuildCreateEvent: &gateway.GuildCreateEvent{
		Guild: discord.Guild{ID: 1, Name: "Test", SystemChannelID: 2},
	}}

	env := NewEventEnv(event)
	c := Context{Env: Env{"guild": env.Guild, "server": env.Guild}}

	got, err := EvalTemplateToString(context.Background(), "{{guild.name}} {{guild.system_channel_id}} {{server.system_channel_id}}", c)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got != "Test 2 2" {
		t.Fatalf("got %q", got)
	}
}

func TestGuildWithoutSystemChannelIsEmpty(t *testing.T) {
	c := Context{Env: Env{"guild": NewGuildEnv(discord.Guild{ID: 1})}}

	got, err := EvalTemplateToString(context.Background(), "{{guild.system_channel_id}}", c)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got != "" {
		t.Fatalf("got %q", got)
	}
}
