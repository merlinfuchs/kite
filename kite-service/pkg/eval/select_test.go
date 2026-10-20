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

// Without a session nothing can be looked up, so the picks only have their IDs,
// which still mention them.
func TestInteractionEnvExposesPickedOfEntitySelects(t *testing.T) {
	tests := []struct {
		name string
		data discord.InteractionData
		want string
	}{
		{"user", &discord.UserSelectInteraction{CustomID: "x", Values: []discord.UserID{11, 22}}, "11 <@11> <@22> 2"},
		{"role", &discord.RoleSelectInteraction{CustomID: "x", Values: []discord.RoleID{11, 22}}, "11 <@&11> <@&22> 2"},
		{"mentionable", &discord.MentionableSelectInteraction{CustomID: "x", Values: []discord.Snowflake{11, 22}}, "11 <@11> <@22> 2"},
		{"channel", &discord.ChannelSelectInteraction{CustomID: "x", Values: []discord.ChannelID{11, 22}}, "11 <#11> <#22> 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &discord.InteractionEvent{
				User: &discord.User{ID: 1},
				Data: tt.data,
			}

			c := Context{Env: Env{"interaction": NewInteractionEnv(i, nil)}}

			got, err := EvalTemplateToString(context.Background(), "{{interaction.picked.id}} {{interaction.picked.mention}} {{interaction.picked_all[1].mention}} {{len(interaction.picked_all)}}", c)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// A string select has options instead, so there's nothing picked, and reading
// it doesn't fail the flow.
func TestInteractionEnvPickedOfStringSelectIsEmpty(t *testing.T) {
	i := &discord.InteractionEvent{
		User: &discord.User{ID: 1},
		Data: &discord.StringSelectInteraction{CustomID: "x", Values: []string{"red"}},
	}

	c := Context{Env: Env{"interaction": NewInteractionEnv(i, nil)}}

	got, err := EvalTemplateToString(context.Background(), "[{{interaction.picked}}] {{len(interaction.picked_all)}}", c)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got != "[] 0" {
		t.Fatalf("got %q", got)
	}
}
