package eval

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func TestMessageMentions(t *testing.T) {
	msg := discord.Message{
		ID:     1,
		Author: discord.User{ID: 10, Username: "author"},
		Mentions: []discord.GuildUser{
			{User: discord.User{ID: 10, Username: "author"}},
			{User: discord.User{ID: 20, Username: "pinged"}},
		},
	}
	c := NewContext(Env{"message": NewMessageEnv(msg)})

	tests := map[string]string{
		"{{len(message.mentions)}}":                                         "2",
		"{{message.mentions[1].username}}":                                  "pinged",
		"{{first(filter(message.mentions, .id != message.author.id))?.id}}": "20",
	}
	for template, want := range tests {
		res, err := EvalTemplate(context.Background(), template, c)
		if err != nil {
			t.Fatalf("%s: %v", template, err)
		}
		if got := res.String(); got != want {
			t.Errorf("%s = %q, want %q", template, got, want)
		}
	}
}
