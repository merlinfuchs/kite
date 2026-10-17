package eval

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

// ?? dereferences the member, which used to lose its lazy fields, so
// (arg('user') ?? user).role_mentions failed the flow.
func TestLazyFieldsAfterCoalesce(t *testing.T) {
	member := NewMemberEnv(discord.Member{
		User:    discord.User{ID: 2, Username: "other"},
		RoleIDs: []discord.RoleID{5},
	})
	self := NewUserEnv(discord.User{ID: 1, Username: "self"})

	cases := []struct {
		arg  any
		tpl  string
		want string
	}{
		{arg: member, tpl: "{{(arg('user') ?? user).role_mentions}}", want: "<@&5>"},
		{arg: member, tpl: "{{(arg('user') ?? user).username}}", want: "other"},
		{arg: nil, tpl: "{{(arg('user') ?? user).role_mentions}}", want: ""},
		{arg: nil, tpl: "{{(arg('user') ?? user).username}}", want: "self"},
	}

	for _, tt := range cases {
		arg := tt.arg
		c := Context{Env: Env{
			"user": self,
			"arg":  func(name string) any { return arg },
		}}

		got, err := EvalTemplate(context.Background(), tt.tpl, c)
		if err != nil {
			t.Errorf("%s: %v", tt.tpl, err)
			continue
		}
		if got.String() != tt.want {
			t.Errorf("%s = %q, want %q", tt.tpl, got.String(), tt.want)
		}
	}
}
