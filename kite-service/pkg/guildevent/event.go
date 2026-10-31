// Package guildevent has the events Kite derives from Discord's server events,
// for things Discord's events leave out, like what a deleted message
// contained. They're worked out from the cache from before Discord's event
// updated it.
//
// They're gateway events, so they go through the same execution path as the
// Discord events they come from. MessageEdit only decides which updates run
// listeners, which get Discord's event inside it.
package guildevent

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

// MessageDeleteType is the type of Discord's own event, as MessageDelete
// replaces that event for event listeners.
const MessageDeleteType ws.EventType = "MESSAGE_DELETE"

// Event is implemented by every event of this package.
type Event interface {
	ws.Event
	derived()
}

// MessageDelete is Discord's MESSAGE_DELETE with the message that was deleted.
// Discord only sends the IDs, so event listeners get it instead of Discord's
// event. It has the same fields as Discord's event, so stored ones of those
// decode into it.
type MessageDelete struct {
	gateway.MessageDeleteEvent
	// Message is the deleted message as it was cached, nil if the bot didn't
	// see it.
	Message *discord.Message `json:"message,omitempty"`
}

func (e *MessageDelete) Op() ws.OpCode           { return 0 }
func (e *MessageDelete) EventType() ws.EventType { return MessageDeleteType }
func (e *MessageDelete) derived()                {}

// MessageEdit is Discord's MESSAGE_UPDATE when the message was edited. Discord
// also sends MESSAGE_UPDATE when it adds a link preview or the message is
// pinned, which must not run Message Update listeners again. It has the type
// of Discord's event.
type MessageEdit struct {
	gateway.MessageUpdateEvent
}

func (e *MessageEdit) derived() {}

// AutoModActionType is the type of Discord's own event, as AutoModAction
// replaces that event for event listeners.
const AutoModActionType ws.EventType = "AUTO_MODERATION_ACTION_EXECUTION"

// AutoModAction is Discord's AUTO_MODERATION_ACTION_EXECUTION with the name of
// the rule, which Discord leaves out. Event listeners get it instead of
// Discord's event. It has the same fields as Discord's event, so stored ones
// of those decode into it.
type AutoModAction struct {
	gateway.AutoModerationActionExecutionEvent
	// RuleName is empty if the rule couldn't be fetched.
	RuleName string `json:"rule_name,omitempty"`
}

func (e *AutoModAction) Op() ws.OpCode           { return 0 }
func (e *AutoModAction) EventType() ws.EventType { return AutoModActionType }
func (e *AutoModAction) derived()                {}

// autoModActionNames are Discord's names of the actions AutoMod takes,
// lowercased.
var autoModActionNames = map[discord.AutoModerationActionType]string{
	discord.AutoModerationBlockMessage:           "block_message",
	discord.AutoModerationSendAlertMessage:       "send_alert_message",
	discord.AutoModerationTimeout:                "timeout",
	discord.AutoModerationBlockMemberInteraction: "block_member_interaction",
}

// autoModTriggerNames are Discord's names of the trigger types of rules,
// lowercased.
var autoModTriggerNames = map[discord.AutoModerationTriggerType]string{
	discord.AutoModerationKeyword:       "keyword",
	discord.AutoModerationSpam:          "spam",
	discord.AutoModerationKeywordPreset: "keyword_preset",
	discord.AutoModerationMentionSpam:   "mention_spam",
	discord.AutoModerationMemberProfile: "member_profile",
}

// ActionName names the action AutoMod took, like block_message, as the Event
// Filter matches it. Actions Kite doesn't know yet are named by their number.
func (e *AutoModAction) ActionName() string {
	if name, ok := autoModActionNames[e.Action.Type]; ok {
		return name
	}
	return fmt.Sprint(e.Action.Type)
}

// TriggerName names what the rule checks, like keyword. Trigger types Kite
// doesn't know yet are named by their number.
func (e *AutoModAction) TriggerName() string {
	if name, ok := autoModTriggerNames[e.RuleTriggerType]; ok {
		return name
	}
	return fmt.Sprint(e.RuleTriggerType)
}

// New returns an empty event of a type of this package, nil for other types.
// Used to decode stored events. Discord's MESSAGE_DELETE and
// AUTO_MODERATION_ACTION_EXECUTION never run a flow, so stored ones are always
// events of this package.
func New(t ws.EventType) Event {
	switch t {
	case MessageDeleteType:
		return new(MessageDelete)
	case AutoModActionType:
		return new(AutoModAction)
	}
	return nil
}
