package model

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// Every Discord event type needs its event filter targets, or saving a
// listener with a filter fails.
func TestDiscordEventTypesHaveEventFilterTargets(t *testing.T) {
	for _, eventType := range []EventListenerType{
		EventListenerTypeDiscordMessageCreate,
		EventListenerTypeDiscordMessageUpdate,
		EventListenerTypeDiscordMessageDelete,
		EventListenerTypeDiscordGuildMemberAdd,
		EventListenerTypeDiscordGuildMemberRemove,
		EventListenerTypeDiscordMessageReactionAdd,
		EventListenerTypeDiscordMessageReactionRemove,
		EventListenerTypeDiscordGuildCreate,
		EventListenerTypeDiscordGuildDelete,
		EventListenerTypeDiscordGuildBoost,
		EventListenerTypeDiscordDirectMessageCreate,
		EventListenerTypeDiscordGuildAuditLogEntryCreate,
	} {
		if len(flow.EventFilterTargets(string(eventType))) == 0 {
			t.Errorf("%s has no event filter targets", eventType)
		}
	}
}
