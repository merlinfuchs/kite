package flow

import (
	"encoding/json"
	"flag"
	"os"
	"slices"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBotID discord.UserID = 99

type testEventData struct {
	FlowContextData
	event ws.Event
}

func (d testEventData) Event() ws.Event { return d.event }

func filter(target EventFilterTarget, mode ComparsionMode, value string) *FlowNodeData {
	return &FlowNodeData{
		EventFilterTarget: target,
		EventFilterMode:   mode,
		EventFilterValue:  value,
	}
}

func matches(f *FlowNodeData, event ws.Event) bool {
	return eventFilterMatches(f, testEventData{event: event}, testBotID)
}

func TestEventFilterMessageContent(t *testing.T) {
	msg := &gateway.MessageCreateEvent{Message: discord.Message{Content: "!Help me"}}

	assert.False(t, matches(filter(EventFilterTypeMessageContent, ComparsionModeStartsWith, "!help"), msg))

	ignoreCase := filter(EventFilterTypeMessageContent, ComparsionModeStartsWith, "!help")
	ignoreCase.EventFilterIgnoreCase = true
	assert.True(t, matches(ignoreCase, msg))

	edited := &gateway.MessageUpdateEvent{Message: discord.Message{Content: "edited"}}
	assert.True(t, matches(filter(EventFilterTypeMessageContent, ComparsionModeEqual, "edited"), edited))
}

func TestEventFilterMessageFlags(t *testing.T) {
	plain := &gateway.MessageCreateEvent{}
	rich := &gateway.MessageCreateEvent{Message: discord.Message{
		Attachments: []discord.Attachment{{ID: 1}},
		Mentions:    []discord.GuildUser{{User: discord.User{ID: testBotID}}},
	}}

	for _, target := range []EventFilterTarget{
		EventFilterTypeMessageHasAttachment,
		EventFilterTypeMessageMentionsBot,
	} {
		yes := filter(target, ComparsionModeEqual, "")
		no := filter(target, ComparsionModeNotEqual, "")

		assert.True(t, matches(yes, rich), target)
		assert.False(t, matches(yes, plain), target)
		assert.True(t, matches(no, plain), target)
	}

	other := &gateway.MessageCreateEvent{Message: discord.Message{
		Mentions: []discord.GuildUser{{User: discord.User{ID: 5}}},
	}}
	assert.False(t, matches(filter(EventFilterTypeMessageMentionsBot, ComparsionModeEqual, ""), other))
}

func TestEventFilterRole(t *testing.T) {
	member := &discord.Member{RoleIDs: []discord.RoleID{123}}
	msg := &gateway.MessageCreateEvent{Member: member}
	reaction := &gateway.MessageReactionAddEvent{Member: member}
	dm := &gateway.MessageCreateEvent{}

	assert.True(t, matches(filter(EventFilterTypeRoleID, ComparsionModeEqual, "123"), msg))
	assert.True(t, matches(filter(EventFilterTypeRoleID, ComparsionModeEqual, "<@&123>"), reaction))
	assert.False(t, matches(filter(EventFilterTypeRoleID, ComparsionModeNotEqual, "123"), msg))
	assert.False(t, matches(filter(EventFilterTypeRoleID, ComparsionModeEqual, "123"), dm))
	assert.True(t, matches(filter(EventFilterTypeRoleID, ComparsionModeNotEqual, "123"), dm))
}

func TestEventFilterEmoji(t *testing.T) {
	custom := &gateway.MessageReactionAddEvent{Emoji: discord.Emoji{ID: 456, Name: "kite"}}
	heart := &gateway.MessageReactionRemoveEvent{Emoji: discord.Emoji{Name: "❤️"}}

	for _, value := range []string{"456", "kite", "<:kite:456>", "<a:kite:456>"} {
		assert.True(t, matches(filter(EventFilterTypeEmoji, ComparsionModeEqual, value), custom), value)
	}
	assert.False(t, matches(filter(EventFilterTypeEmoji, ComparsionModeEqual, "👍"), custom))
	assert.True(t, matches(filter(EventFilterTypeEmoji, ComparsionModeEqual, "❤"), heart))
	assert.True(t, matches(filter(EventFilterTypeEmoji, ComparsionModeNotEqual, "👍"), heart))
}

func TestEventFilterUserIsBot(t *testing.T) {
	ignoreBots := filter(EventFilterTypeUserIsBot, ComparsionModeNotEqual, "")

	assert.False(t, matches(ignoreBots, &gateway.GuildMemberAddEvent{Member: discord.Member{User: discord.User{Bot: true}}}))
	assert.True(t, matches(ignoreBots, &gateway.GuildMemberRemoveEvent{User: discord.User{}}))
	assert.False(t, matches(ignoreBots, &gateway.MessageReactionAddEvent{Member: &discord.Member{User: discord.User{Bot: true}}}))
}

func TestEventFilterAuditLogAction(t *testing.T) {
	ban := &gateway.GuildAuditLogEntryCreateEvent{AuditLogEntry: discord.AuditLogEntry{ActionType: discord.MemberBanAdd}, GuildID: 1}
	timeout := &gateway.GuildAuditLogEntryCreateEvent{AuditLogEntry: discord.AuditLogEntry{
		ActionType: discord.MemberUpdate,
		Changes:    []discord.AuditLogChange{{Key: "communication_disabled_until", NewValue: []byte(`"2030-01-01T00:00:00Z"`)}},
	}, GuildID: 1}

	assert.True(t, matches(filter(EventFilterTypeAuditLogAction, ComparsionModeOneOf, "member_kick, member_ban_add"), ban))
	assert.False(t, matches(filter(EventFilterTypeAuditLogAction, ComparsionModeNotOneOf, "member_kick,member_ban_add"), ban))
	assert.True(t, matches(filter(EventFilterTypeAuditLogAction, ComparsionModeOneOf, "member_timeout_add,member_timeout_remove"), timeout))
	// Timeouts are named on their own, not as member updates.
	assert.False(t, matches(filter(EventFilterTypeAuditLogAction, ComparsionModeOneOf, "member_update"), timeout))
}

func TestEventFilterOneOf(t *testing.T) {
	msg := &gateway.MessageCreateEvent{Message: discord.Message{Content: "Yes"}}

	assert.True(t, matches(filter(EventFilterTypeMessageContent, ComparsionModeOneOf, "yes,Yes"), msg))
	assert.False(t, matches(filter(EventFilterTypeMessageContent, ComparsionModeNotOneOf, "no, Yes"), msg))

	ignoreCase := filter(EventFilterTypeMessageContent, ComparsionModeOneOf, "YES,no")
	ignoreCase.EventFilterIgnoreCase = true
	assert.True(t, matches(ignoreCase, msg))
}

func TestEventFilterDeletedMessage(t *testing.T) {
	seen := &guildevent.MessageDelete{
		MessageDeleteEvent: gateway.MessageDeleteEvent{ID: 1, ChannelID: 2, GuildID: 3},
		Message:            &discord.Message{ID: 1, Content: "secret"},
	}
	unseen := &guildevent.MessageDelete{MessageDeleteEvent: gateway.MessageDeleteEvent{ID: 1, ChannelID: 2, GuildID: 3}}

	assert.True(t, matches(filter(EventFilterTypeMessageContent, ComparsionModeContains, "sec"), seen))
	assert.False(t, matches(filter(EventFilterTypeMessageContent, ComparsionModeContains, "sec"), unseen))
}

func eventFlowWithFilter(eventType string, f FlowNodeData) FlowData {
	return FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryEvent, Data: FlowNodeData{EventType: eventType}},
			{ID: "filter", Type: FlowNodeTypeOptionEventFilter, Data: f},
		},
		Edges: []FlowEdge{{ID: "e", Source: "filter", Target: "entry"}},
	}
}

func TestFilterEvent(t *testing.T) {
	entry, err := CompileEventListener(eventFlowWithFilter("guild_member_add", *filter(EventFilterTypeUserIsBot, ComparsionModeNotEqual, "")))
	require.NoError(t, err)

	bot := &gateway.GuildMemberAddEvent{Member: discord.Member{User: discord.User{Bot: true}}}
	human := &gateway.GuildMemberAddEvent{Member: discord.Member{User: discord.User{}}}
	assert.False(t, entry.FilterEvent(testEventData{event: bot}, testBotID))
	assert.True(t, entry.FilterEvent(testEventData{event: human}, testBotID))
}

func TestValidateEventFilters(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		filter    *FlowNodeData
		wantErr   bool
	}{
		{"content on messages", "message_create", filter(EventFilterTypeMessageContent, ComparsionModeContains, "hi"), false},
		{"content on edited messages", "message_update", filter(EventFilterTypeMessageContent, ComparsionModeContains, "hi"), false},
		{"content on member joins", "guild_member_add", filter(EventFilterTypeMessageContent, ComparsionModeContains, "hi"), true},
		{"emoji on messages", "message_create", filter(EventFilterTypeEmoji, ComparsionModeEqual, "👍"), true},
		{"emoji on reactions", "message_reaction_add", filter(EventFilterTypeEmoji, ComparsionModeEqual, "👍"), false},
		{"role on removed reactions", "message_reaction_remove", filter(EventFilterTypeRoleID, ComparsionModeEqual, "1"), true},
		{"role contains", "message_create", filter(EventFilterTypeRoleID, ComparsionModeContains, "1"), true},
		{"yes or no contains", "message_create", filter(EventFilterTypeMessageHasAttachment, ComparsionModeContains, ""), true},
		{"ignore bots on joins", "guild_member_add", filter(EventFilterTypeUserIsBot, ComparsionModeNotEqual, ""), false},
		{"unknown event type", "something_new", filter(EventFilterTypeGuildID, ComparsionModeEqual, "1"), true},
		{"audit log action one of", "guild_audit_log_entry_create", filter(EventFilterTypeAuditLogAction, ComparsionModeOneOf, "member_ban_add"), false},
		{"audit log action equal", "guild_audit_log_entry_create", filter(EventFilterTypeAuditLogAction, ComparsionModeEqual, "member_ban_add"), true},
		{"audit log action on messages", "message_create", filter(EventFilterTypeAuditLogAction, ComparsionModeOneOf, "member_ban_add"), true},
		{"channels one of", "message_create", filter(EventFilterTypeChannelID, ComparsionModeOneOf, "1,2"), false},
		{"role one of", "message_create", filter(EventFilterTypeRoleID, ComparsionModeOneOf, "1,2"), true},
		{"content of deleted messages", "message_delete", filter(EventFilterTypeMessageContent, ComparsionModeContains, "hi"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := CompileEventListener(eventFlowWithFilter(tt.eventType, *tt.filter))
			require.NoError(t, err)

			err = ValidateEventFilters(entry)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

var update = flag.Bool("update", false, "update the generated files")

// The editor reads the event filter targets of each event type and the modes
// of each target from this file, run `go test ./pkg/flow -update` to update it.
const eventFilterTargetsFile = "../../../kite-web/src/lib/flow/eventFilterTargets.gen.json"

func TestEventFilterTargetsFile(t *testing.T) {
	file := struct {
		Events  map[string][]EventFilterTarget         `json:"events"`
		Modes   map[EventFilterTarget][]ComparsionMode `json:"modes"`
		Boolean []EventFilterTarget                    `json:"boolean"`
	}{Events: eventFilterTargets, Modes: map[EventFilterTarget][]ComparsionMode{}}

	for _, targets := range eventFilterTargets {
		for _, target := range targets {
			if modes := eventFilterModes(target); modes != nil {
				file.Modes[target] = modes
			}
			if isBooleanEventFilterTarget(target) && !slices.Contains(file.Boolean, target) {
				file.Boolean = append(file.Boolean, target)
			}
		}
	}
	slices.Sort(file.Boolean)

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')

	if *update {
		if err := os.WriteFile(eventFilterTargetsFile, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	current, err := os.ReadFile(eventFilterTargetsFile)
	if err != nil || string(current) != string(data) {
		t.Fatalf("%s is out of date, run `go test ./pkg/flow -update`", eventFilterTargetsFile)
	}
}
