package eval

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func commandTestContext(data *discord.CommandInteraction) Context {
	i := &discord.InteractionEvent{
		User: &discord.User{ID: 10, Username: "invoker"},
		Data: data,
	}
	return Context{Env: Env{"interaction": NewInteractionEnv(i, nil)}}
}

func TestCommandInteractionExposesContextTargets(t *testing.T) {
	tests := []struct {
		name     string
		data     *discord.CommandInteraction
		template string
		want     string
	}{
		{
			name: "user target",
			data: &discord.CommandInteraction{
				ID:       1,
				Name:     "userinfo",
				TargetID: 2,
				Resolved: discord.ResolvedData{
					Users: map[discord.UserID]discord.User{
						2: {ID: 2, Username: "target"},
					},
				},
			},
			template: "{{interaction.command.target_user.username}}",
			want:     "target",
		},
		{
			name: "user target member data",
			data: &discord.CommandInteraction{
				ID:       1,
				Name:     "userinfo",
				TargetID: 2,
				Resolved: discord.ResolvedData{
					Users: map[discord.UserID]discord.User{
						2: {ID: 2, Username: "target"},
					},
					Members: map[discord.UserID]discord.Member{
						2: {Nick: "server nickname"},
					},
				},
			},
			template: "{{interaction.command.target_user.nick}}",
			want:     "server nickname",
		},
		{
			name: "message target",
			data: &discord.CommandInteraction{
				ID:       1,
				Name:     "messageinfo",
				TargetID: 2,
				Resolved: discord.ResolvedData{
					Messages: map[discord.MessageID]discord.Message{
						2: {
							ID:      2,
							Content: "selected message",
							Author:  discord.User{ID: 3, Username: "author"},
						},
					},
				},
			},
			template: "{{interaction.command.target_message.content}} by {{interaction.command.target_message.author.username}}",
			want:     "selected message by author",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := commandTestContext(tt.data)

			got, err := EvalTemplateToString(context.Background(), tt.template, ctx)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCommandInteractionContextTargetsCanBeMissing(t *testing.T) {
	tests := []struct {
		name     string
		data     *discord.CommandInteraction
		template string
	}{
		{
			name:     "slash command user",
			data:     &discord.CommandInteraction{ID: 1, Name: "ping"},
			template: "{{interaction.command.target_user.id}}",
		},
		{
			name:     "slash command message",
			data:     &discord.CommandInteraction{ID: 1, Name: "ping"},
			template: "{{interaction.command.target_message.author.id}}",
		},
		{
			name: "missing user data",
			data: &discord.CommandInteraction{
				ID:       1,
				TargetID: 2,
			},
			template: "{{interaction.command.target_user.id}}",
		},
		{
			name: "missing message data",
			data: &discord.CommandInteraction{
				ID:       1,
				TargetID: 2,
			},
			template: "{{interaction.command.target_message.content}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := commandTestContext(tt.data)

			got, err := EvalTemplateToString(context.Background(), tt.template, ctx)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			if got != "" {
				t.Fatalf("got %q, want empty", got)
			}
		})
	}
}

func TestCommandInteractionContextTargetSurvivesResume(t *testing.T) {
	original := commandTestContext(&discord.CommandInteraction{
		ID:       1,
		Name:     "userinfo",
		TargetID: 2,
		Resolved: discord.ResolvedData{
			Users: map[discord.UserID]discord.User{
				2: {ID: 2, Username: "target"},
			},
		},
	})
	resumed := NewContext(Env{})
	resumed.SetResumeContext([]Context{original})

	got, err := EvalTemplateToString(
		context.Background(),
		"{{origin.interaction.command.target_user.username}}",
		resumed,
	)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got != "target" {
		t.Fatalf("got %q, want %q", got, "target")
	}
}
