package flow

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

// eventFilterTargets lists what event filters can check for each Discord event
// type. A target that isn't in the list doesn't exist for the event and would
// compare against nothing, so saving such a filter is an error. New event
// listener types have to be added here, with the targets their event data has.
var eventFilterTargets = map[string][]EventFilterTarget{
	"message_create": {
		EventFilterTypeMessageContent, EventFilterTypeUserID, EventFilterTypeGuildID,
		EventFilterTypeChannelID, EventFilterTypeMessageID, EventFilterTypeRoleID,
		EventFilterTypeMessageHasAttachment, EventFilterTypeMessageIsReply,
		EventFilterTypeMessageMentionsBot,
	},
	"message_update": {
		EventFilterTypeMessageContent, EventFilterTypeUserID, EventFilterTypeGuildID,
		EventFilterTypeChannelID, EventFilterTypeMessageID, EventFilterTypeRoleID,
		EventFilterTypeMessageHasAttachment, EventFilterTypeMessageIsReply,
		EventFilterTypeMessageMentionsBot,
	},
	"message_delete": {
		EventFilterTypeGuildID, EventFilterTypeChannelID, EventFilterTypeMessageID,
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
}

// EventFilterTargets returns what event filters can check for an event type.
// Schedules and webhooks have no Discord event, so there's nothing to filter.
func EventFilterTargets(eventType string) []EventFilterTarget {
	return eventFilterTargets[eventType]
}

func isBooleanEventFilterTarget(target EventFilterTarget) bool {
	switch target {
	case EventFilterTypeMessageHasAttachment,
		EventFilterTypeMessageIsReply,
		EventFilterTypeMessageMentionsBot,
		EventFilterTypeUserIsBot:
		return true
	}
	return false
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

		mode := node.Data.EventFilterMode
		switch {
		case isBooleanEventFilterTarget(target):
			if node.Data.EventFilterValue != "true" && node.Data.EventFilterValue != "false" {
				return fmt.Errorf("event filter %s needs the value true or false", target)
			}
			fallthrough
		case target == EventFilterTypeRoleID, target == EventFilterTypeEmoji:
			if mode != ComparsionModeEqual && mode != ComparsionModeNotEqual {
				return fmt.Errorf("event filter %s can only be equal or not equal", target)
			}
		}
	}

	return nil
}

// FilterEvent reports whether the event passes every event filter of the flow.
// botID is the app's own user, zero if unknown.
func (n *CompiledFlowNode) FilterEvent(ctx *FlowContext, botID discord.UserID) (bool, error) {
	for _, node := range n.Parents.Default {
		if node.IsEventFilter() && !eventFilterMatches(&node.Data, ctx.Data, botID) {
			return false, nil
		}
	}

	return true, nil
}

// eventFilterMatches reports whether an event passes a filter. The value of a
// role filter is a role ID, and equal means the member has the role. The value
// of an emoji filter is a unicode emoji, or the ID, name or mention of a
// custom emoji. Yes or no filters, like user_is_bot, have the value "true" or
// "false". EventFilterIgnoreCase only applies to message content.
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
		return (eventFilterBool(target, event, botID) == (value == "true")) != notEqual
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
	}

	return true
}

func eventMessage(event any) *discord.Message {
	switch e := event.(type) {
	case *gateway.MessageCreateEvent:
		return &e.Message
	case *gateway.MessageUpdateEvent:
		return &e.Message
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
	case EventFilterTypeMessageIsReply:
		return msg.Type == discord.InlinedReplyMessage
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
