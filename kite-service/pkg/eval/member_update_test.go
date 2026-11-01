package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

func memberUpdateContext(e *guildevent.MemberUpdate) Context {
	env := NewEventEnv(e, nil)
	return NewContext(Env{
		"event":         env,
		"user":          env.User,
		"member":        env.Member,
		"member_update": env.MemberUpdate,
	})
}

func TestMemberUpdatePlaceholders(t *testing.T) {
	tests := map[string]struct {
		event    guildevent.MemberUpdate
		template string
		want     string
	}{
		"nickname": {
			guildevent.MemberUpdate{
				GuildMemberUpdateEvent: gateway.GuildMemberUpdateEvent{GuildID: testGuildID, User: discord.User{ID: 7}, Nick: "new"},
				Old:                    discord.Member{User: discord.User{ID: 7}, Nick: "old"},
			},
			"{{member_update}}|{{member_update.old_nickname}} → {{member_update.new_nickname}}|{{user.id}}|{{member.nick}}|{{member_update.screening_passed}}",
			"nickname|old → new|7|new|false",
		},
		"screening passed": {
			guildevent.MemberUpdate{
				GuildMemberUpdateEvent: gateway.GuildMemberUpdateEvent{GuildID: testGuildID, User: discord.User{ID: 7}},
				Old:                    discord.Member{User: discord.User{ID: 7}, IsPending: true},
			},
			"{{member_update.changes}}|{{member_update.screening_passed}}|{{member_update.nickname_changed}}",
			"screening_passed|true|false",
		},
		"server avatar": {
			guildevent.MemberUpdate{
				GuildMemberUpdateEvent: gateway.GuildMemberUpdateEvent{GuildID: testGuildID, User: discord.User{ID: 7}, Avatar: "abc"},
				Old:                    discord.Member{User: discord.User{ID: 7}},
			},
			"{{member_update.avatar_changed}}|{{member_update.old_avatar_url}}|{{member_update.new_avatar_url != ''}}",
			"true||true",
		},
	}

	for name, tt := range tests {
		if got := evalString(t, memberUpdateContext(&tt.event), tt.template); got != tt.want {
			t.Errorf("%s: got %q, want %q", name, got, tt.want)
		}
	}
}
