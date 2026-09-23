package flow

import (
	"encoding/json"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	arikawajson "github.com/diamondburned/arikawa/v3/utils/json"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type triggerTestData struct {
	TestContextData
	event ws.Event
}

func (d *triggerTestData) Interaction() *discord.InteractionEvent {
	return d.interaction
}

func (d *triggerTestData) Event() ws.Event {
	return d.event
}

func commandInteraction() *discord.InteractionEvent {
	return &discord.InteractionEvent{
		ID:      1,
		Token:   "secret",
		GuildID: 2,
		Member:  &discord.Member{User: discord.User{ID: 3, Username: "alice"}},
		Message: &discord.Message{ID: 4},
		Data: &discord.CommandInteraction{
			ID:   5,
			Name: "ban",
			Options: discord.CommandInteractionOptions{
				{Type: discord.StringOptionType, Name: "reason", Value: arikawajson.Raw(`"spam"`)},
			},
		},
	}
}

func modalInteraction(customID, value string) *discord.InteractionEvent {
	return &discord.InteractionEvent{
		ID:     6,
		Member: &discord.Member{User: discord.User{ID: 7}},
		Data: &discord.ModalInteraction{
			CustomID: "modal",
			Components: discord.TopLevelComponents{
				&discord.ActionRowComponent{
					&discord.TextInputComponent{CustomID: discord.ComponentID(customID), Value: value},
				},
			},
		},
	}
}

func roundTrip(t *testing.T, state FlowContextState) FlowContextState {
	t.Helper()

	data, err := json.Marshal(state)
	require.NoError(t, err)

	var res FlowContextState
	require.NoError(t, json.Unmarshal(data, &res))
	return res
}

func TestRecordTriggerStripsInteraction(t *testing.T) {
	state := NewFlowContextState()
	state.recordTrigger(&triggerTestData{TestContextData: TestContextData{interaction: commandInteraction()}})

	res := roundTrip(t, *state)
	require.NotNil(t, res.Origin())
	require.NotNil(t, res.Origin().Interaction)

	i := res.Origin().Interaction
	assert.Empty(t, i.Token)
	assert.Nil(t, i.Message)
	assert.Equal(t, discord.UserID(3), i.Member.User.ID)

	cmd, ok := i.Data.(*discord.CommandInteraction)
	require.True(t, ok)
	assert.Equal(t, "spam", cmd.Options[0].String())
}

func TestRecordTriggerRoundTripsEvent(t *testing.T) {
	state := NewFlowContextState()
	state.recordTrigger(&triggerTestData{event: &gateway.MessageCreateEvent{
		Message: discord.Message{ID: 1, Content: "hello"},
	}})

	res := roundTrip(t, *state)
	require.NotNil(t, res.Origin())

	event, ok := res.Origin().Event.(*gateway.MessageCreateEvent)
	require.True(t, ok)
	assert.Equal(t, "hello", event.Content)
}

func TestRecordTriggerKeepsOriginAndAppendsPrevious(t *testing.T) {
	state := NewFlowContextState()
	state.recordTrigger(&triggerTestData{TestContextData: TestContextData{interaction: commandInteraction()}})
	assert.Same(t, state.Origin(), state.Previous())

	resumed := state.Copy()
	resumed.recordTrigger(&triggerTestData{TestContextData: TestContextData{interaction: modalInteraction("name", "bob")}})

	assert.Equal(t, discord.InteractionID(1), resumed.Origin().Interaction.ID)
	assert.Equal(t, discord.InteractionID(6), resumed.Previous().Interaction.ID)
	// The state the first resume point was created from is untouched.
	assert.Len(t, state.Triggers, 1)
}

func TestRecordTriggerKeepsOriginAndNewestTriggers(t *testing.T) {
	state := NewFlowContextState()
	for _, value := range []string{"1", "2", "3", "4", "5", "6"} {
		state.recordTrigger(&triggerTestData{TestContextData: TestContextData{interaction: modalInteraction("field", value)}})
	}

	res := roundTrip(t, *state)
	require.Len(t, res.Triggers, maxStoredTriggers)

	var values []string
	for _, inputs := range res.ModalInputs() {
		values = append(values, inputs["field"])
	}
	assert.Equal(t, []string{"6", "5", "4", "1"}, values)
}
