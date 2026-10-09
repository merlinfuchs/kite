package gateway

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store/defaultstore"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// Protocol frames report an empty event type. The dispatch path can never
// match one, so they are dropped at the gateway handler -- but only because
// nothing downstream uses an empty type as a real key. These tests pin that
// assumption, since it is what makes the filter safe.

func TestProtocolEventsReportEmptyEventType(t *testing.T) {
	// A representative sample of arikawa's protocol frames. If a future
	// version gives one of these a real event type the filter simply stops
	// dropping it, which is harmless. The dangerous direction is the reverse,
	// covered below.
	protocol := []ws.Event{
		&gateway.HeartbeatAckEvent{},
		&gateway.HelloEvent{},
		&gateway.ReconnectEvent{},
		func() *gateway.InvalidSessionEvent { e := gateway.InvalidSessionEvent(false); return &e }(),
	}

	for _, e := range protocol {
		if got := e.EventType(); got != "" {
			t.Errorf("%T reports event type %q, expected empty", e, got)
		}
	}
}

func TestForumPostEventOnlyWrapsForumThreads(t *testing.T) {
	cabinet := defaultstore.New()
	forum := discord.Channel{ID: 200, GuildID: 100, Type: discord.GuildForum, Name: "ideas"}
	text := discord.Channel{ID: 201, GuildID: 100, Type: discord.GuildText, Name: "chat"}
	for _, channel := range []*discord.Channel{&forum, &text} {
		if err := cabinet.ChannelSet(channel, false); err != nil {
			t.Fatal(err)
		}
	}
	session := &state.State{Cabinet: cabinet}

	post := discord.Channel{
		ID:          300,
		GuildID:     100,
		Type:        discord.GuildPublicThread,
		Name:        "A suggestion",
		OwnerID:     400,
		ParentID:    forum.ID,
		AppliedTags: []discord.TagID{500},
	}
	tests := []struct {
		name      string
		event     gateway.Event
		eventType ws.EventType
	}{
		{"create", &gateway.ThreadCreateEvent{Channel: post}, eventTypeForumPostCreate},
		{"update", &gateway.ThreadUpdateEvent{Channel: post}, eventTypeForumPostUpdate},
		{"delete", &gateway.ThreadDeleteEvent{
			ID: post.ID, GuildID: post.GuildID, Type: post.Type, ParentID: post.ParentID,
		}, eventTypeForumPostDelete},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped, ok := forumPostEvent(session, tt.event)
			if !ok {
				t.Fatal("forum thread was not wrapped")
			}
			if got := wrapped.EventType(); got != tt.eventType {
				t.Errorf("event type = %q, want %q", got, tt.eventType)
			}
			custom := wrapped.(*forumPostGatewayEvent)
			if custom.OriginalEvent() != tt.event {
				t.Error("original Discord event was not preserved")
			}
			if custom.ForumPost().ID != post.ID {
				t.Errorf("post id = %s, want %s", custom.ForumPost().ID, post.ID)
			}
		})
	}

	nonForumPost := post
	nonForumPost.ParentID = text.ID
	if _, ok := forumPostEvent(session, &gateway.ThreadCreateEvent{Channel: nonForumPost}); ok {
		t.Error("text-channel thread was wrapped as a forum post")
	}

	privatePost := post
	privatePost.Type = discord.GuildPrivateThread
	if _, ok := forumPostEvent(session, &gateway.ThreadCreateEvent{Channel: privatePost}); ok {
		t.Error("private thread was wrapped as a forum post")
	}
}

// The filter is only safe if no event the engine acts on carries an empty
// type. If one ever does, it would be silently swallowed at the gateway.
func TestDispatchedEventsHaveNonEmptyEventType(t *testing.T) {
	dispatched := []ws.Event{
		&gateway.MessageCreateEvent{},
		&gateway.MessageUpdateEvent{},
		&gateway.MessageDeleteEvent{},
		&gateway.GuildMemberAddEvent{},
		&gateway.GuildMemberRemoveEvent{},
		&gateway.MessageReactionAddEvent{},
		&gateway.MessageReactionRemoveEvent{},
		&gateway.ThreadCreateEvent{},
		&gateway.ThreadUpdateEvent{},
		&gateway.ThreadDeleteEvent{},
		&gateway.InteractionCreateEvent{},
		&gateway.GuildCreateEvent{},
		&gateway.GuildDeleteEvent{},
		&state.GuildJoinEvent{GuildCreateEvent: &gateway.GuildCreateEvent{}},
		&state.GuildLeaveEvent{GuildDeleteEvent: &gateway.GuildDeleteEvent{}},
		&gateway.ReadyEvent{},
	}

	for _, e := range dispatched {
		if e.EventType() == "" {
			t.Errorf("%T reports an empty event type and would be filtered out", e)
		}
	}
}

// No event listener type may be empty, or the filter would drop events that
// listener was registered for.
func TestNoEventListenerTypeIsEmpty(t *testing.T) {
	types := []model.EventListenerType{
		model.EventListenerTypeDiscordMessageCreate,
		model.EventListenerTypeDiscordMessageUpdate,
		model.EventListenerTypeDiscordMessageDelete,
		model.EventListenerTypeDiscordGuildMemberAdd,
		model.EventListenerTypeDiscordGuildMemberRemove,
		model.EventListenerTypeDiscordMessageReactionAdd,
		model.EventListenerTypeDiscordMessageReactionRemove,
		model.EventListenerTypeDiscordForumPostCreate,
		model.EventListenerTypeDiscordForumPostUpdate,
		model.EventListenerTypeDiscordForumPostDelete,
		model.EventListenerTypeDiscordGuildCreate,
		model.EventListenerTypeDiscordGuildDelete,
	}

	for _, tp := range types {
		if tp == "" {
			t.Error("an event listener type is empty, which the protocol filter would swallow")
		}
	}
}
