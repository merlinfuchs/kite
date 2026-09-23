package flow

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

func acknowledge(data discord.InteractionData, responded bool) *TestDiscordProvider {
	discordProvider := &TestDiscordProvider{responded: responded}
	acknowledgeUnansweredComponent(&FlowContext{
		Context:       context.Background(),
		FlowProviders: FlowProviders{Discord: discordProvider},
		Data:          &TestContextData{interaction: &discord.InteractionEvent{Data: data}},
	})
	return discordProvider
}

// A button whose flow never responds would show "This interaction failed".
func TestUnansweredComponentIsAcknowledged(t *testing.T) {
	if got := acknowledge(&discord.ButtonInteraction{CustomID: "x"}, false).response; got.Type != api.DeferredMessageUpdate {
		t.Errorf("got %+v, want a DeferredMessageUpdate", got)
	}
}

func TestAnsweredComponentIsLeftAlone(t *testing.T) {
	if got := acknowledge(&discord.ButtonInteraction{CustomID: "x"}, true).response; got.Type != 0 {
		t.Errorf("got %+v, want no response", got)
	}
}

// Commands can't be acknowledged without a visible response.
func TestUnansweredCommandIsLeftAlone(t *testing.T) {
	if got := acknowledge(&discord.CommandInteraction{}, false).response; got.Type != 0 {
		t.Errorf("got %+v, want no response", got)
	}
}
