package flow

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	arikawajson "github.com/diamondburned/arikawa/v3/utils/json"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eventData is a trigger without an interaction, which TestContextData can't be.
type eventData struct {
	TestContextData
	event ws.Event
}

func (d *eventData) Interaction() *discord.InteractionEvent {
	return nil
}

func (d *eventData) Event() ws.Event {
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
	state.recordTrigger(&TestContextData{interaction: commandInteraction()})

	res := roundTrip(t, *state)
	require.Len(t, res.Triggers, 1)
	require.NotNil(t, res.Triggers[0].Interaction)

	i := res.Triggers[0].Interaction
	assert.Empty(t, i.Token)
	assert.Nil(t, i.Message)
	assert.Equal(t, discord.UserID(3), i.Member.User.ID)

	cmd, ok := i.Data.(*discord.CommandInteraction)
	require.True(t, ok)
	assert.Equal(t, "spam", cmd.Options[0].String())
}

func TestRecordTriggerRoundTripsEvent(t *testing.T) {
	state := NewFlowContextState()
	state.recordTrigger(&eventData{event: &gateway.MessageCreateEvent{
		Message: discord.Message{ID: 1, Content: "hello"},
	}})

	res := roundTrip(t, *state)
	require.Len(t, res.Triggers, 1)

	event, ok := res.Triggers[0].Event.(*gateway.MessageCreateEvent)
	require.True(t, ok)
	assert.Equal(t, "hello", event.Content)
}

func TestRecordTriggerAppendsWithoutTouchingCopies(t *testing.T) {
	state := NewFlowContextState()
	state.recordTrigger(&TestContextData{interaction: commandInteraction()})

	resumed := state.Copy()
	resumed.recordTrigger(&TestContextData{interaction: modalInteraction("name", "bob")})

	require.Len(t, resumed.Triggers, 2)
	assert.Equal(t, discord.InteractionID(1), resumed.Triggers[0].Interaction.ID)
	assert.Equal(t, discord.InteractionID(6), resumed.Triggers[1].Interaction.ID)
	// The state the first resume point was created from is untouched.
	assert.Len(t, state.Triggers, 1)
}

func TestRecordTriggerKeepsFirstAndNewestTriggers(t *testing.T) {
	state := NewFlowContextState()
	for _, value := range []string{"1", "2", "3", "4", "5", "6"} {
		state.recordTrigger(&TestContextData{interaction: modalInteraction("field", value)})
	}

	res := roundTrip(t, *state)

	var values []string
	for _, trigger := range res.Triggers {
		modal, ok := trigger.Interaction.Data.(*discord.ModalInteraction)
		require.True(t, ok)
		row := modal.Components[0].(*discord.ActionRowComponent)
		values = append(values, (*row)[0].(*discord.TextInputComponent).Value)
	}
	assert.Equal(t, []string{"1", "4", "5", "6"}, values)
}

func TestFlowTriggerScheduleEventRoundTrip(t *testing.T) {
	// A button sent by a scheduled flow resumes with the schedule event as the
	// trigger that started it.
	occurrence := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	trigger := newFlowTrigger(&eventData{event: &schedule.Event{Time: occurrence}})
	require.NotNil(t, trigger)

	data, err := json.Marshal(trigger)
	require.NoError(t, err)

	var decoded FlowTrigger
	require.NoError(t, json.Unmarshal(data, &decoded))

	event, ok := decoded.Event.(*schedule.Event)
	require.True(t, ok, "decoded event is %T", decoded.Event)
	assert.True(t, event.Time.Equal(occurrence))
}
