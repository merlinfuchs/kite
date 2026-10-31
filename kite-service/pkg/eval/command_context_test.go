package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func TestContextCommandTarget(t *testing.T) {
	tests := map[string]struct {
		data     *discord.CommandInteraction
		template string
		want     string
	}{
		"user": {
			data: &discord.CommandInteraction{
				TargetID: 200,
				Resolved: discord.ResolvedData{
					Users: map[discord.UserID]discord.User{
						200: {ID: 200, Username: "target-user"},
					},
				},
			},
			template: "{{context.target.id}}|{{context.target.username}}",
			want:     "200|target-user",
		},
		"message": {
			data: &discord.CommandInteraction{
				TargetID: 300,
				Resolved: discord.ResolvedData{
					Messages: map[discord.MessageID]discord.Message{
						300: {ID: 300, ChannelID: 400, Content: "target message"},
					},
				},
			},
			template: "{{context.target.id}}|{{context.target.content}}",
			want:     "300|target message",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			i := &discord.InteractionEvent{
				GuildID: testGuildID,
				Member:  &discord.Member{User: discord.User{ID: testUserID}},
				Data:    test.data,
			}
			interaction := NewInteractionEnv(i, testSession(t))
			context := NewContext(Env{"context": interaction.Command.Context})
			if got := evalString(t, context, test.template); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
