package flow

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"

	"github.com/kitecloud/kite/kite-service/pkg/auditlog"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

// eventFilterTargets lists what event filters can check for each Discord event
// type. A target that isn't in the list doesn't exist for the event and would
// compare against nothing, so saving such a filter is an error. New event
// listener types have to be added here, with the targets their event data has,
// and the editor's copy regenerated with `go test ./pkg/flow -update`.
var eventFilterTargets = map[string][]EventFilterTarget{
	"message_create": {
		EventFilterTypeMessageContent, EventFilterTypeUserID, EventFilterTypeGuildID,
		EventFilterTypeChannelID, EventFilterTypeMessageID, EventFilterTypeRoleID,
		EventFilterTypeMessageHasAttachment, EventFilterTypeMessageMentionsBot,
	},
	"message_update": {
		EventFilterTypeMessageContent, EventFilterTypeUserID, EventFilterTypeGuildID,
		EventFilterTypeChannelID, EventFilterTypeMessageID, EventFilterTypeRoleID,
		EventFilterTypeMessageHasAttachment, EventFilterTypeMessageMentionsBot,
	},
	// DMs have no server, so no guild or roles.
	"direct_message_create": {
		EventFilterTypeMessageContent, EventFilterTypeUserID, EventFilterTypeChannelID,
		EventFilterTypeMessageID, EventFilterTypeMessageHasAttachment,
		EventFilterTypeMessageMentionsBot,
	},
	// The content and author are only known if the bot saw the message.
	"message_delete": {
		EventFilterTypeMessageContent, EventFilterTypeUserID, EventFilterTypeGuildID,
		EventFilterTypeChannelID, EventFilterTypeMessageID,
	},
	"guild_member_add": {
		EventFilterTypeUserID, EventFilterTypeGuildID, EventFilterTypeRoleID,
		EventFilterTypeUserIsBot,
	},
	"guild_member_remove": {
		EventFilterTypeUserID, EventFilterTypeGuildID, EventFilterTypeUserIsBot,
	},
	"message_reaction_add": {
		EventFilterTypeUserID, EventFilterTypeGuildID, EventFilterTypeChannelID,
		EventFilterTypeMessageID, EventFilterTypeRoleID, EventFilterTypeEmoji,
		EventFilterTypeUserIsBot,
	},
	// Discord only sends the user's ID with removed reactions, so there are no
	// roles and no way to tell bots apart.
	"message_reaction_remove": {
		EventFilterTypeUserID, EventFilterTypeGuildID, EventFilterTypeChannelID,
		EventFilterTypeMessageID, EventFilterTypeEmoji,
	},
	"guild_create": {EventFilterTypeGuildID},
	"guild_delete": {EventFilterTypeGuildID},
	// Boosts come from the boost system message, whose content says nothing.
	"guild_boost": {
		EventFilterTypeUserID, EventFilterTypeGuildID, EventFilterTypeChannelID,
		EventFilterTypeMessageID, EventFilterTypeRoleID,
	},
	// The user is whoever did it.
	"guild_audit_log_entry_create": {
		EventFilterTypeAuditLogAction, EventFilterTypeUserID, EventFilterTypeGuildID,
		EventFilterTypeChannelID,
	},
}

// EventFilterTargets returns what event filters can check for an event type.
// Schedules and webhooks have no Discord event, so there's nothing to filter.
func EventFilterTargets(eventType string) []EventFilterTarget {
	return eventFilterTargets[eventType]
}

func isBooleanEventFilterTarget(target EventFilterTarget) bool {
	switch target {
	case EventFilterTypeMessageHasAttachment,
		EventFilterTypeMessageMentionsBot,
		EventFilterTypeUserIsBot:
		return true
	}
	return false
}

// eventFilterModes returns the modes a target can use, nil if it can use all.
// Yes or no filters, roles and emojis are either equal or not, and audit log
// actions are picked from a list.
func eventFilterModes(target EventFilterTarget) []ComparsionMode {
	switch {
	case isBooleanEventFilterTarget(target), target == EventFilterTypeRoleID, target == EventFilterTypeEmoji:
		return []ComparsionMode{ComparsionModeEqual, ComparsionModeNotEqual}
	case target == EventFilterTypeAuditLogAction:
		return []ComparsionMode{ComparsionModeOneOf, ComparsionModeNotOneOf}
	}
	return nil
}

// ValidateEventFilters checks that the event filters of an event listener fit
// its event, so a filter can't silently stop the listener from ever running.
func ValidateEventFilters(entry *CompiledFlowNode) error {
	targets := EventFilterTargets(entry.Data.EventType)

	for _, node := range entry.Parents.Default {
		if !node.IsEventFilter() {
			continue
		}

		target := node.Data.EventFilterTarget
		if !slices.Contains(targets, target) {
			if len(targets) == 0 {
				return fmt.Errorf("%s event listeners can't have event filters", entry.Data.EventType)
			}
			return fmt.Errorf("event filter %s isn't available for %s events", target, entry.Data.EventType)
		}

		if modes := eventFilterModes(target); modes != nil && !slices.Contains(modes, node.Data.EventFilterMode) {
			return fmt.Errorf("event filter %s can only use the modes %v", target, modes)
		}
	}

	return nil
}

// FilterEvent reports whether the event passes every event filter of the flow.
// botID is the app's own user, zero if unknown.
func (n *CompiledFlowNode) FilterEvent(data FlowContextData, botID discord.UserID) bool {
	for _, node := range n.Parents.Default {
		if node.IsEventFilter() && !eventFilterMatches(&node.Data, data, botID) {
			return false
		}
	}

	return true
}

// eventFilterMatches reports whether an event passes a filter. The value of a
// role filter is a role ID, and equal means the member has the role. The value
// of an emoji filter is a unicode emoji, or the ID, name or mention of a
// custom emoji. Yes or no filters, like user_is_bot, have no value, equal means
// yes and not equal no. EventFilterIgnoreCase only applies to message content.
func eventFilterMatches(filter *FlowNodeData, data FlowContextData, botID discord.UserID) bool {
	event := data.Event()
	value := filter.EventFilterValue
	notEqual := filter.EventFilterMode == ComparsionModeNotEqual

	switch target := filter.EventFilterTarget; {
	case target == EventFilterTypeRoleID:
		return eventMemberHasRole(event, value) != notEqual
	case target == EventFilterTypeEmoji:
		return eventEmojiMatches(event, value) != notEqual
	case isBooleanEventFilterTarget(target):
		return eventFilterBool(target, event, botID) != notEqual
	}

	var subject string
	switch filter.EventFilterTarget {
	case EventFilterTypeMessageContent:
		if msg := eventMessage(event); msg != nil {
			subject = msg.Content
		}
	case EventFilterTypeUserID:
		subject = data.UserID().String()
	case EventFilterTypeGuildID:
		subject = data.GuildID().String()
	case EventFilterTypeChannelID:
		subject = data.ChannelID().String()
	case EventFilterTypeMessageID:
		subject = data.MessageID().String()
	case EventFilterTypeAuditLogAction:
		if e, ok := event.(*gateway.GuildAuditLogEntryCreateEvent); ok {
			subject = auditlog.Action(e.AuditLogEntry)
		}
	}

	if filter.EventFilterIgnoreCase && filter.EventFilterTarget == EventFilterTypeMessageContent {
		subject = strings.ToLower(subject)
		value = strings.ToLower(value)
	}

	switch filter.EventFilterMode {
	case ComparsionModeEqual:
		return subject == value
	case ComparsionModeNotEqual:
		return subject != value
	case ComparsionModeContains:
		return strings.Contains(subject, value)
	case ComparsionModeStartsWith:
		return strings.HasPrefix(subject, value)
	case ComparsionModeEndsWith:
		return strings.HasSuffix(subject, value)
	case ComparsionModeOneOf:
		return isOneOf(subject, value)
	case ComparsionModeNotOneOf:
		return !isOneOf(subject, value)
	}

	return true
}

// isOneOf reports whether subject is one of the values separated by commas.
func isOneOf(subject string, values string) bool {
	for _, v := range strings.Split(values, ",") {
		if strings.TrimSpace(v) == subject {
			return true
		}
	}
	return false
}

func eventMessage(event any) *discord.Message {
	switch e := event.(type) {
	case *gateway.MessageCreateEvent:
		return &e.Message
	case *gateway.MessageUpdateEvent:
		return &e.Message
	case *guildevent.MessageDelete:
		return e.Message
	}
	return nil
}

func eventFilterBool(target EventFilterTarget, event any, botID discord.UserID) bool {
	if target == EventFilterTypeUserIsBot {
		switch e := event.(type) {
		case *gateway.GuildMemberAddEvent:
			return e.User.Bot
		case *gateway.GuildMemberRemoveEvent:
			return e.User.Bot
		case *gateway.MessageReactionAddEvent:
			return e.Member != nil && e.Member.User.Bot
		}
		return false
	}

	msg := eventMessage(event)
	if msg == nil {
		return false
	}

	switch target {
	case EventFilterTypeMessageHasAttachment:
		return len(msg.Attachments) > 0
	case EventFilterTypeMessageMentionsBot:
		return botID.IsValid() && slices.ContainsFunc(msg.Mentions, func(u discord.GuildUser) bool {
			return u.ID == botID
		})
	}
	return false
}

func eventMemberHasRole(event any, value string) bool {
	var member *discord.Member
	switch e := event.(type) {
	case *gateway.MessageCreateEvent:
		member = e.Member
	case *gateway.MessageUpdateEvent:
		member = e.Member
	case *gateway.MessageReactionAddEvent:
		member = e.Member
	case *gateway.GuildMemberAddEvent:
		member = &e.Member
	}
	if member == nil {
		return false
	}

	// Accept a role mention too, it's what people copy from Discord.
	roleID := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(value), "<@&"), ">")
	return slices.ContainsFunc(member.RoleIDs, func(id discord.RoleID) bool {
		return id.String() == roleID
	})
}

var customEmojiPattern = regexp.MustCompile(`^<a?:(\w+):(\d+)>$`)

func eventEmojiMatches(event any, value string) bool {
	var emoji discord.Emoji
	switch e := event.(type) {
	case *gateway.MessageReactionAddEvent:
		emoji = e.Emoji
	case *gateway.MessageReactionRemoveEvent:
		emoji = e.Emoji
	default:
		return false
	}

	value = strings.TrimSpace(value)
	if match := customEmojiPattern.FindStringSubmatch(value); match != nil {
		value = match[2]
	}

	if emoji.ID.IsValid() {
		return value == emoji.ID.String() || value == emoji.Name
	}

	// Some unicode emojis come with or without a variation selector.
	return strings.ReplaceAll(value, "\uFE0F", "") == strings.ReplaceAll(emoji.Name, "\uFE0F", "")
}
