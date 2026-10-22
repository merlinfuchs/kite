package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
)

func TestAppPlaceholders(t *testing.T) {
	session := testSession(t)
	other := discord.Guild{ID: 200, Name: "Other", ApproximateMembers: 66}
	if err := session.Cabinet.GuildSet(&other, true); err != nil {
		t.Fatal(err)
	}

	c := Context{Env: Env{"app": newAppEnv(session, discord.User{ID: 900, Username: "kite"})}}

	if got := evalString(t, c, "{{app.guild_count}}|{{app.member_count}}|{{app.user.id}}"); got != "2|1300|900" {
		t.Errorf("got %q", got)
	}
	if got := evalString(t, c, "{{app.guild_count + 1}}"); got != "3" {
		t.Errorf("guild_count isn't a number: got %q", got)
	}
	// The test session never connected to the gateway.
	if got := evalString(t, c, "{{app.latency}}"); got != "0" {
		t.Errorf("latency: got %q", got)
	}
}

func TestAppPlaceholdersWithoutServers(t *testing.T) {
	session := &state.State{Cabinet: defaultstore.New()}
	c := Context{Env: Env{"app": newAppEnv(session, discord.User{ID: 901})}}

	got := evalString(t, c, "{{app.guild_count}}|{{app.member_count}}|{{app.latency}}")
	if want := "0|0|0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
