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

	c := Context{Env: Env{"interaction": NewInteractionEnv(i)}}

	got, err := EvalTemplateToString(context.Background(), "{{interaction.value}} {{interaction.values[1]}} {{len(interaction.values)}}", c)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got != "red blue 2" {
		t.Fatalf("got %q", got)
	}
}
