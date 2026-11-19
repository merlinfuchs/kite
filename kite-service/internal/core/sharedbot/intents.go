package sharedbot

import (
	"slices"
	"sync/atomic"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

// appIntents are the privileged data the templates of an app read, see
// bundle.Bundle.SharedBotIntents. Intents are global, so the Kite bot gets
// message text and member events for every server once Discord grants them.
// Apps without a template that reads them get messages without their text
// and no member events, like a bot without the intents. They're updated when
// the apps are loaded again, also while scopes use them.
type appIntents struct {
	messageContent atomic.Bool
	guildMembers   atomic.Bool
}

// set takes the intents of the official bundles the app installed.
func (a *appIntents) set(bundleIDs []string) {
	var content, members bool
	for _, id := range bundleIDs {
		intents := bundleIntents[id]
		content = content || slices.Contains(intents, bundle.SharedBotIntentMessageContent)
		members = members || slices.Contains(intents, bundle.SharedBotIntentGuildMembers)
	}
	a.messageContent.Store(content)
	a.guildMembers.Store(members)
}

// bundleIntents are the intents of the official bundles by ID.
var bundleIntents = func() map[string][]string {
	res := make(map[string][]string)
	for _, b := range bundle.Official() {
		if len(b.SharedBotIntents) > 0 {
			res[b.ID] = b.SharedBotIntents
		}
	}
	return res
}()

// filterEvent is the event as an app with the intents gets it, or false if
// it doesn't get it at all.
func filterEvent(e gateway.Event, intents *appIntents) (gateway.Event, bool) {
	switch e := e.(type) {
	case *gateway.MessageCreateEvent:
		if !intents.messageContent.Load() {
			copied := *e
			copied.Message = withoutContent(e.Message)
			return &copied, true
		}
	case *gateway.MessageUpdateEvent:
		if !intents.messageContent.Load() {
			copied := *e
			copied.Message = withoutContent(e.Message)
			return &copied, true
		}
	case *guildevent.MessageEdit:
		if !intents.messageContent.Load() {
			copied := *e
			copied.Message = withoutContent(e.Message)
			return &copied, true
		}
	case *guildevent.MessageDelete:
		if e.Message != nil && !intents.messageContent.Load() {
			copied := *e
			stripped := withoutContent(*e.Message)
			copied.Message = &stripped
			return &copied, true
		}
	case *gateway.GuildMemberAddEvent, *gateway.GuildMemberRemoveEvent,
		*gateway.GuildMemberUpdateEvent, *gateway.GuildMembersChunkEvent,
		*guildevent.BoostRemove:
		if !intents.guildMembers.Load() {
			return nil, false
		}
	}
	return e, true
}

// withoutContent is the message without what Discord leaves out for bots
// without the Message Content intent, also of the message it replies to or
// forwards.
func withoutContent(m discord.Message) discord.Message {
	m.Content = ""
	m.Embeds = nil
	m.Attachments = nil
	m.Components = nil
	m.MessageSnapshots = nil
	if m.ReferencedMessage != nil {
		referenced := withoutContent(*m.ReferencedMessage)
		m.ReferencedMessage = &referenced
	}
	return m
}
