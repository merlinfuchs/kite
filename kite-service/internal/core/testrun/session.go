package testrun

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

func (s *testSession) addMessage(msg discord.Message) discord.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg.ID = discord.MessageID(flow.TestSnowflake())
	msg.Author = s.session.Ready().User
	msg.Timestamp = discord.NewTimestamp(time.Now())
	s.messages = append(s.messages, msg)
	return msg
}

// setOriginal records the message an interaction responded with, which
// edits of its response change.
func (s *testSession) setOriginal(interactionID discord.InteractionID, messageID discord.MessageID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.originals[interactionID] = messageID
}

// editMessage applies an edit to a message and stores it, so its new
// components can be used.
func (s *testSession) editMessage(msg discord.Message, content option.NullableString, embeds *[]discord.Embed, components *discord.TopLevelComponents) discord.Message {
	if content != nil {
		msg.Content = content.Val
	}
	if embeds != nil {
		msg.Embeds = *embeds
	}
	if components != nil {
		msg.Components = *components
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if i := s.messageIndex(msg.ID); i != -1 {
		s.messages[i] = msg
	} else {
		s.messages = append(s.messages, msg)
	}
	return msg
}

// editOriginal edits the response of the current interaction: the message it
// created, the message of a clicked component, or the message a defer was
// waiting for.
func (s *testSession) editOriginal(content option.NullableString, embeds *[]discord.Embed, components *discord.TopLevelComponents) discord.Message {
	s.mu.Lock()
	interaction := s.current
	var original discord.Message
	found := false
	if interaction != nil {
		if id, ok := s.originals[interaction.ID]; ok {
			i := s.messageIndex(id)
			if i != -1 {
				original, found = s.messages[i], true
			}
		}
		if !found && interaction.Message != nil {
			original, found = *interaction.Message, true
		}
	}
	s.mu.Unlock()

	if !found {
		var msg discord.Message
		if interaction != nil {
			msg.ChannelID, msg.GuildID = interaction.ChannelID, interaction.GuildID
		}
		original = s.addMessage(msg)
		if interaction != nil {
			s.setOriginal(interaction.ID, original.ID)
		}
	}
	return s.editMessage(original, content, embeds, components)
}

// deleteMessage forgets a message, so its components can't be used anymore.
func (s *testSession) deleteMessage(id discord.MessageID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if i := s.messageIndex(id); i != -1 {
		s.messages = slices.Delete(s.messages, i, i+1)
	}
}

// deleteOriginal forgets the response of the current interaction.
func (s *testSession) deleteOriginal() {
	s.mu.Lock()
	var id discord.MessageID
	if s.current != nil {
		id = s.originals[s.current.ID]
		if !id.IsValid() && s.current.Message != nil {
			id = s.current.Message.ID
		}
	}
	s.mu.Unlock()
	s.deleteMessage(id)
}

// messageIndex is the index of a sent message, -1 if there's none. The lock
// has to be held.
func (s *testSession) messageIndex(id discord.MessageID) int {
	return slices.IndexFunc(s.messages, func(m discord.Message) bool { return m.ID == id })
}

func (s *testSession) fakeMessage(id discord.MessageID) (discord.Message, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.messageIndex(id)
	if i == -1 {
		return discord.Message{}, false
	}
	return s.messages[i], true
}

func (s *testSession) addChannel(channel discord.Channel) discord.Channel {
	s.mu.Lock()
	defer s.mu.Unlock()
	channel.ID = discord.ChannelID(flow.TestSnowflake())
	s.channels[channel.ID] = channel
	return channel
}

func (s *testSession) fakeChannel(id discord.ChannelID) (discord.Channel, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	channel, ok := s.channels[id]
	return channel, ok
}

func (s *testSession) guildOfChannel(id discord.ChannelID) discord.GuildID {
	if channel, ok := s.fakeChannel(id); ok {
		return channel.GuildID
	}
	if channel, err := s.session.Channel(id); err == nil {
		return channel.GuildID
	}
	return 0
}

func (s *testSession) linkTemplate(messageID discord.MessageID, templateID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[messageID] = templateID
}

func (s *testSession) setModal(data *api.InteractionResponseData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modal = data
}

// markResponded records a response to an interaction, false if it had one.
func (s *testSession) markResponded(id discord.InteractionID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.respondedTo[id] {
		return false
	}
	s.respondedTo[id] = true
	return true
}

func (s *testSession) responded(id discord.InteractionID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.respondedTo[id]
}

func (s *testSession) currentInteraction() *discord.InteractionEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

func (s *testSession) addResumePoint(rp flow.ResumePoint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	point := testResumePoint{rp: rp, root: s.root}
	if rp.Type == flow.ResumePointTypeTimer {
		s.timers = append(s.timers, point)
		return
	}
	s.resumePoints[rp.ID] = point
}

func (s *testSession) variable(id string, scope null.String) (thing.Thing, bool, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.variables[testVariableKey{id: id, scope: scope}]
	return v.value, v.deleted, ok
}

func (s *testSession) setVariable(id string, scope null.String, value thing.Thing, deleted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variables[testVariableKey{id: id, scope: scope}] = testVariableValue{value: value, deleted: deleted, updatedAt: time.Now()}
}

// setVariables sets the values of stored variables the test starts with.
func (s *testSession) setVariables(ctx context.Context, variables []Variable) error {
	if len(variables) == 0 {
		return nil
	}

	existing, err := s.env.VariableStore.VariablesByAppWithoutTotals(ctx, s.appID)
	if err != nil {
		return fmt.Errorf("failed to get variables: %w", err)
	}
	scoped := make(map[string]bool, len(existing))
	for _, v := range existing {
		scoped[v.ID] = v.Scoped
	}

	for _, v := range variables {
		isScoped, ok := scoped[v.VariableID]
		switch {
		case !ok:
			return testErrorf("the app has no stored variable with the ID %q", v.VariableID)
		case isScoped && !v.Scope.Valid:
			return testErrorf("variable %s has a value per scope, it needs the scope", v.VariableID)
		case !isScoped && v.Scope.Valid:
			return testErrorf("variable %s has one value for the app, it has no scope", v.VariableID)
		}
		s.setVariable(v.VariableID, v.Scope, v.Value, v.Value.IsNil())
	}
	return nil
}

// variableChanges returns the values of a variable the test run changed, by
// scope.
func (s *testSession) variableChanges(id string) map[null.String]testVariableValue {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := map[null.String]testVariableValue{}
	for key, v := range s.variables {
		if key.id == id {
			res[key.scope] = v
		}
	}
	return res
}
