package eval

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
)

func TestInteractionEnvExposesSelectValues(t *testing.T) {
	i := &discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.StringSelectInteraction{CustomID: "x", Values: []string{"red", "blue"}},
	}

	c := Context{Env: Env{"interaction": NewInteractionEnv(i, nil)}}

	got, err := EvalTemplateToString(context.Background(), "{{interaction.value}} {{interaction.values[1]}} {{len(interaction.values)}}", c)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got != "red blue 2" {
		t.Fatalf("got %q", got)
	}
}

func TestInteractionEnvExposesEntitySelectValues(t *testing.T) {
	tests := []struct {
		name string
		data discord.InteractionData
	}{
		{"user", &discord.UserSelectInteraction{CustomID: "x", Values: []discord.UserID{11, 22}}},
		{"role", &discord.RoleSelectInteraction{CustomID: "x", Values: []discord.RoleID{11, 22}}},
		{"mentionable", &discord.MentionableSelectInteraction{CustomID: "x", Values: []discord.Snowflake{11, 22}}},
		{"channel", &discord.ChannelSelectInteraction{CustomID: "x", Values: []discord.ChannelID{11, 22}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &discord.InteractionEvent{
				User: &discord.User{ID: 1},
				Data: tt.data,
			}

			c := Context{Env: Env{"interaction": NewInteractionEnv(i, nil)}}

			got, err := EvalTemplateToString(context.Background(), "{{interaction.value}} {{interaction.values[1]}} {{len(interaction.values)}}", c)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			if got != "11 22 2" {
				t.Fatalf("got %q", got)
			}
		})
	}
}
