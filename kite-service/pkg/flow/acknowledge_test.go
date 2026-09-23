package flow

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

type recordingDiscord struct {
	provider.MockDiscordProvider
	responded bool
	responses []api.InteractionResponse
}

func (p *recordingDiscord) HasCreatedInteractionResponse(ctx context.Context, interactionID discord.InteractionID) (bool, error) {
	return p.responded, nil
}

func (p *recordingDiscord) CreateInteractionResponse(ctx context.Context, interactionID discord.InteractionID, interactionToken string, response api.InteractionResponse) (*provider.InteractionResponseResource, error) {
	p.responses = append(p.responses, response)
	return nil, nil
}

type interactionData struct {
	FlowContextData
	interaction *discord.InteractionEvent
}

func (d interactionData) Interaction() *discord.InteractionEvent {
	return d.interaction
}

func acknowledge(data discord.InteractionData, responded bool) *recordingDiscord {
	discordProvider := &recordingDiscord{responded: responded}
	AcknowledgeUnansweredComponent(&FlowContext{
		FlowProviders: FlowProviders{Discord: discordProvider},
		Data:          interactionData{interaction: &discord.InteractionEvent{Data: data}},
	})
	return discordProvider
}

// A button whose flow never responds would show "This interaction failed".
func TestUnansweredComponentIsAcknowledged(t *testing.T) {
	got := acknowledge(&discord.ButtonInteraction{CustomID: "x"}, false).responses
	if len(got) != 1 || got[0].Type != api.DeferredMessageUpdate {
		t.Errorf("got %+v, want one DeferredMessageUpdate", got)
	}
}

func TestAnsweredComponentIsLeftAlone(t *testing.T) {
	if got := acknowledge(&discord.ButtonInteraction{CustomID: "x"}, true).responses; len(got) != 0 {
		t.Errorf("got %+v, want no response", got)
	}
}

// Commands can't be acknowledged without a visible response.
func TestUnansweredCommandIsLeftAlone(t *testing.T) {
	if got := acknowledge(&discord.CommandInteraction{}, false).responses; len(got) != 0 {
		t.Errorf("got %+v, want no response", got)
	}
}
