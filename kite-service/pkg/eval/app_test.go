package eval

import (
	"strconv"
	"testing"
	"time"

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

	bot := discord.User{ID: 900, Username: "kite"}
	AppConnected(bot.ID)
	c := Context{Env: Env{"app": newAppEnv(session, bot)}}

	if got := evalString(t, c, "{{app.guild_count}}|{{app.member_count}}|{{app.user.id}}"); got != "2|1300|900" {
		t.Errorf("got %q", got)
	}
	if got := evalString(t, c, "{{app.guild_count + 1}}"); got != "3" {
		t.Errorf("guild_count isn't a number: got %q", got)
	}

	connectedAt, err := strconv.ParseInt(evalString(t, c, "{{app.connected_at}}"), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	if now := time.Now().Unix(); connectedAt < now-5 || connectedAt > now {
		t.Errorf("connected_at: got %d, now is %d", connectedAt, now)
	}
	uptime, err := strconv.ParseInt(evalString(t, c, "{{app.uptime}}"), 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	if uptime < 0 || uptime > 5 {
		t.Errorf("uptime: got %d", uptime)
	}
	// The test session never connected to the gateway.
	if got := evalString(t, c, "{{app.latency}}"); got != "0" {
		t.Errorf("latency: got %q", got)
	}
}

func TestAppPlaceholdersBeforeConnecting(t *testing.T) {
	session := &state.State{Cabinet: defaultstore.New()}
	c := Context{Env: Env{"app": newAppEnv(session, discord.User{ID: 901})}}

	got := evalString(t, c, "{{app.guild_count}}|{{app.member_count}}|{{app.uptime}}|{{app.connected_at}}|{{app.latency}}")
	if want := "0|0|0|0|0"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
