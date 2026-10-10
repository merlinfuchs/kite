package engine

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
)

// Every guild the bot is in arrives as a raw GUILD_CREATE on connect, so only
// arikawa's derived join and leave events may trigger the bot server listeners.
func TestShouldHandleGuildJoinAndLeave(t *testing.T) {
	tests := []struct {
		name  string
		event ws.Event
		want  bool
	}{
		{"raw guild create", &gateway.GuildCreateEvent{}, false},
		{"raw guild delete", &gateway.GuildDeleteEvent{}, false},
		{"guild ready on connect", &state.GuildReadyEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, false},
		{"guild available after outage", &state.GuildAvailableEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, false},
		{"guild unavailable", &state.GuildUnavailableEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{Unavailable: true}}, false},
		{"guild join", &state.GuildJoinEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}}, true},
		{"guild leave", &state.GuildLeaveEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{}}, true},
	}

	l := &EventListener{}
	for _, tt := range tests {
		if got := l.shouldHandleEvent(tt.event, 0); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

// A flow that adds or removes a reaction must not trigger its own reaction
// listeners.
func TestShouldHandleReactionsIgnoresOwnApp(t *testing.T) {
	const botID = discord.UserID(1)
	const userID = discord.UserID(2)

	tests := []struct {
		name  string
		event ws.Event
		want  bool
	}{
		{"reaction add by user", &gateway.MessageReactionAddEvent{UserID: userID}, true},
		{"reaction add by app", &gateway.MessageReactionAddEvent{UserID: botID}, false},
		{"reaction remove by user", &gateway.MessageReactionRemoveEvent{UserID: userID}, true},
		{"reaction remove by app", &gateway.MessageReactionRemoveEvent{UserID: botID}, false},
	}

	l := &EventListener{}
	for _, tt := range tests {
		if got := l.shouldHandleEvent(tt.event, botID); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

// testBoostEvent is the system message Discord posts when a member boosts, as
// it arrives from the gateway. The author is the member who boosted.
func testBoostEvent(msgType discord.MessageType) *gateway.MessageCreateEvent {
	author := discord.User{ID: 100, Username: "booster"}

	return &gateway.MessageCreateEvent{
		Message: discord.Message{
			ID:        900,
			Type:      msgType,
			ChannelID: 300,
			GuildID:   200,
			Author:    author,
		},
		// Discord leaves the user out of the member of a message, arikawa's
		// state copies the author into it before handlers run.
		Member: &discord.Member{
			User:    author,
			Nick:    "Boosty",
			RoleIDs: []discord.RoleID{400},
		},
	}
}

func listenerIDs(listeners []*EventListener) []string {
	ids := make([]string, len(listeners))
	for i, l := range listeners {
		ids[i] = l.listener.ID
	}
	return ids
}

func TestIsBoostMessage(t *testing.T) {
	tests := []struct {
		name    string
		msgType discord.MessageType
		want    bool
	}{
		{"boost", discord.NitroBoostMessage, true},
		{"boost reaching level 1", discord.NitroTier1Message, true},
		{"boost reaching level 2", discord.NitroTier2Message, true},
		{"boost reaching level 3", discord.NitroTier3Message, true},
		{"normal message", discord.DefaultMessage, false},
		{"member join message", discord.GuildMemberJoinMessage, false},
		{"channel follow message", discord.ChannelFollowAddMessage, false},
	}

	for _, tt := range tests {
		if got := isBoostMessage(discord.Message{Type: tt.msgType}); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

// A boost is a MESSAGE_CREATE on the wire, so the boost listeners have to be
// picked up for it without a gateway event type of their own.
func TestBoostMessageReachesBoostListeners(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddEventListener(testListener("message", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))
	app.AddEventListener(testListener("boost", model.EventSourceDiscord, model.EventListenerTypeDiscordGuildBoost))
	app.AddEventListener(testListener("join", model.EventSourceDiscord, model.EventListenerTypeDiscordGuildMemberAdd))

	tests := []struct {
		name  string
		event gateway.Event
		want  []string
	}{
		{"boost", testBoostEvent(discord.NitroBoostMessage), []string{"message", "boost"}},
		{"boost reaching level 2", testBoostEvent(discord.NitroTier2Message), []string{"message", "boost"}},
		{"normal message", testBoostEvent(discord.DefaultMessage), []string{"message"}},
		{"member join", &gateway.GuildMemberAddEvent{}, []string{"join"}},
	}

	for _, tt := range tests {
		got := listenerIDs(app.listenersForEvent(tt.event))
		if len(got) != len(tt.want) {
			t.Errorf("%s: got listeners %v, want %v", tt.name, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("%s: got listeners %v, want %v", tt.name, got, tt.want)
				break
			}
		}
	}
}

// Dispatching a boost must not grow the message create index, or every later
// message would run the boost listeners too.
func TestBoostDispatchDoesNotMutateListenerIndex(t *testing.T) {
	app := NewApp("app-1", Env{})
	app.AddEventListener(testListener("message", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate))
	app.AddEventListener(testListener("boost", model.EventSourceDiscord, model.EventListenerTypeDiscordGuildBoost))

	for i := 0; i < 3; i++ {
		app.listenersForEvent(testBoostEvent(discord.NitroBoostMessage))
	}

	got := listenerIDs(app.listenersForEvent(testBoostEvent(discord.DefaultMessage)))
	if len(got) != 1 || got[0] != "message" {
		t.Errorf("normal message got listeners %v, want [message]", got)
	}
}

func TestShouldHandleBoost(t *testing.T) {
	_, boost := testListener("boost", model.EventSourceDiscord, model.EventListenerTypeDiscordGuildBoost)
	_, message := testListener("message", model.EventSourceDiscord, model.EventListenerTypeDiscordMessageCreate)

	if !boost.shouldHandleEvent(testBoostEvent(discord.NitroBoostMessage), 0) {
		t.Error("boost listener ignored a boost message")
	}
	if boost.shouldHandleEvent(testBoostEvent(discord.DefaultMessage), 0) {
		t.Error("boost listener handled a normal message")
	}
	if !message.shouldHandleEvent(testBoostEvent(discord.DefaultMessage), 0) {
		t.Error("message create listener ignored a normal message")
	}
}

// The placeholders of a boost listener describe the member who boosted and
// the server and channel the boost message was posted in.
func TestBoostEventPlaceholders(t *testing.T) {
	event := testBoostEvent(discord.NitroBoostMessage)

	data := &EventData{event: event}
	if got := data.UserID(); got != 100 {
		t.Errorf("UserID() = %d, want 100", got)
	}
	if got := data.GuildID(); got != 200 {
		t.Errorf("GuildID() = %d, want 200", got)
	}
	if got := data.ChannelID(); got != 300 {
		t.Errorf("ChannelID() = %d, want 300", got)
	}

	evalCtx := eval.NewContext(eval.Env{})
	env := eval.NewEventEnv(event, nil)
	evalCtx.Env["user"] = env.User
	evalCtx.Env["member"] = env.Member
	evalCtx.Env["guild"] = env.Guild
	evalCtx.Env["channel"] = env.Channel
	evalCtx.Env["message"] = env.Message

	tests := []struct {
		template string
		want     string
	}{
		{"{{user}}", "<@100>"},
		{"{{user.id}}", "100"},
		{"{{user.mention}}", "<@100>"},
		{"{{user.username}}", "booster"},
		{"{{user.display_name}}", "booster"},
		{"{{user.nick}}", "Boosty"},
		{"{{member.id}}", "100"},
		{"{{guild.id}}", "200"},
		{"{{channel.id}}", "300"},
		{"{{message.id}}", "900"},
	}

	for _, tt := range tests {
		got, err := eval.EvalTemplateToString(context.Background(), tt.template, evalCtx)
		if err != nil {
			t.Errorf("%s: %v", tt.template, err)
			continue
		}
		if got != tt.want {
			t.Errorf("%s = %q, want %q", tt.template, got, tt.want)
		}
	}
}
