package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

// messageDeleteContext builds the context of a Message Delete event like
// NewContextFromEvent does, without the app user, which needs a Ready event.
func messageDeleteContext(id discord.MessageID, msg *discord.Message, session *state.State) Context {
	env := NewEventEnv(&guildevent.MessageDelete{
		MessageDeleteEvent: gateway.MessageDeleteEvent{ID: id, ChannelID: 21, GuildID: testGuildID},
		Message:            msg,
	}, session)
	return NewContext(Env{
		"event":   env,
		"user":    env.User,
		"member":  env.Member,
		"message": env.Message,
	})
}

func TestMessageDeletePlaceholders(t *testing.T) {
	session := testSession(t)
	author := discord.User{ID: testUserID, Username: "jack"}
	// The state caches the member of the author when a message is created.
	member := discord.Member{User: author, Nick: "Jacky", RoleIDs: []discord.RoleID{12}}
	if err := session.Cabinet.MemberSet(testGuildID, &member, false); err != nil {
		t.Fatal(err)
	}

	const template = "{{message.id}}|{{message.content}}|{{user.id}}|{{user.username}}"
	tests := map[string]struct {
		msg  discord.Message
		want string
	}{
		// Sent in a server by a member that is cached.
		"with member": {
			discord.Message{ID: 1, ChannelID: 21, GuildID: testGuildID, Author: author, Content: "deleted text"},
			"1|deleted text|175928847299117063|jack",
		},
		// Sent by someone who isn't cached as a member.
		"without member": {
			discord.Message{ID: 2, ChannelID: 21, GuildID: testGuildID, Author: discord.User{ID: 42, Username: "someone"}, Content: "no member"},
			"2|no member|42|someone",
		},
	}
	for name, tt := range tests {
		c := messageDeleteContext(tt.msg.ID, &tt.msg, session)
		if got := evalString(t, c, template); got != tt.want {
			t.Errorf("%s: got %q, want %q", name, got, tt.want)
		}
	}

	// The bot never saw the message, so only its ID is known.
	env := NewEventEnv(&guildevent.MessageDelete{
		MessageDeleteEvent: gateway.MessageDeleteEvent{ID: 3, ChannelID: 21, GuildID: testGuildID},
	}, session)
	if env.User != nil {
		t.Errorf("user is set for an unknown message: %#v", env.User)
	}
	if env.Message.ID != "3" || env.Message.Content != "" {
		t.Errorf("unknown message: got ID %q and content %q", env.Message.ID, env.Message.Content)
	}

	// The nickname of the member is available, as for Message Create.
	msg := tests["with member"].msg
	c := messageDeleteContext(msg.ID, &msg, session)
	if got := evalString(t, c, "{{member.nick}}"); got != "Jacky" {
		t.Fatalf("member nick: got %q, want %q", got, "Jacky")
	}
}

func TestMemberRemovePlaceholders(t *testing.T) {
	session := testSession(t)
	user := discord.User{ID: testUserID, Username: "jack"}

	withMember := NewEventEnv(&guildevent.MemberRemove{
		GuildMemberRemoveEvent: gateway.GuildMemberRemoveEvent{GuildID: testGuildID, User: user},
		Member:                 &discord.Member{Nick: "Jacky", RoleIDs: []discord.RoleID{11, 12}},
	}, session)
	c := NewContext(Env{"user": withMember.User, "member": withMember.Member})
	if got := evalString(t, c, "{{user.username}}|{{user.nick}}|{{user.role_ids}}|{{user.role_names}}|{{member.role_count}}"); got != "jack|Jacky|[11 12]|Admin, Member|2" {
		t.Errorf("with member: got %q", got)
	}

	// Without a cached member, user fields work and member fields are empty.
	withoutMember := NewEventEnv(&guildevent.MemberRemove{
		GuildMemberRemoveEvent: gateway.GuildMemberRemoveEvent{GuildID: testGuildID, User: user},
	}, session)
	c = NewContext(Env{"user": withoutMember.User, "member": withoutMember.Member})
	if got := evalString(t, c, "{{user.username}}|{{user.nick}}|{{user.role_ids}}|{{user.role_names}}|{{member.role_count}}"); got != "jack||[]||0" {
		t.Errorf("without member: got %q", got)
	}
}
