package eval

import (
	"context"
	"errors"
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

func TestAppUserInstallCount(t *testing.T) {
	session := &state.State{Cabinet: defaultstore.New()}
	app := newAppEnv(session, discord.User{ID: 900})
	c := Context{Env: Env{"app": app}}

	if got := evalString(t, c, "{{app.user_install_count}}"); got != "0" {
		t.Errorf("without loader: got %q", got)
	}

	var loads int
	app.SetUserInstallCount(func() (int, error) {
		loads++
		return 42, nil
	})
	if got := evalString(t, c, "{{app.user_install_count}}|{{app.user_install_count + 1}}"); got != "42|43" {
		t.Errorf("got %q", got)
	}
	if loads != 1 {
		t.Errorf("loaded %d times, want 1", loads)
	}

	app.SetUserInstallCount(func() (int, error) { return 0, errors.New("db down") })
	if _, err := EvalTemplateToString(context.Background(), "{{app.user_install_count}}", c); err == nil {
		t.Error("expected the error of the loader")
	}
}
