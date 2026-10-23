package flow

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func filtered(target EventFilterTarget, mode ComparsionMode, value string) *CompiledFlowNode {
	return &CompiledFlowNode{
		Type: FlowNodeTypeEntryEvent,
		Parents: ConnectedFlowNodes{Default: []*CompiledFlowNode{{
			Type: FlowNodeTypeOptionEventFilter,
			Data: FlowNodeData{EventFilterTarget: target, EventFilterMode: mode, EventFilterValue: value},
		}}},
	}
}

func TestFilterEventAuditLogAction(t *testing.T) {
	ban := &gateway.GuildAuditLogEntryCreateEvent{AuditLogEntry: discord.AuditLogEntry{ActionType: discord.MemberBanAdd}, GuildID: 1}
	timeout := &gateway.GuildAuditLogEntryCreateEvent{AuditLogEntry: discord.AuditLogEntry{
		ActionType: discord.MemberUpdate,
		Changes:    []discord.AuditLogChange{{Key: "communication_disabled_until", NewValue: []byte(`"2030-01-01T00:00:00Z"`)}},
	}, GuildID: 1}

	tests := []struct {
		name  string
		event *gateway.GuildAuditLogEntryCreateEvent
		mode  ComparsionMode
		value string
		want  bool
	}{
		{"ban equal", ban, ComparsionModeEqual, "member_ban_add", true},
		{"ban one of", ban, ComparsionModeOneOf, "member_kick, member_ban_add", true},
		{"ban not one of", ban, ComparsionModeNotOneOf, "member_kick,member_ban_add", false},
		{"timeout one of", timeout, ComparsionModeOneOf, "member_timeout_add,member_timeout_remove", true},
		{"timeout isn't a member update", timeout, ComparsionModeOneOf, "member_update", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := filtered(EventFilterTypeAuditLogAction, tt.mode, tt.value)
			got, err := node.FilterEvent(&FlowContext{Data: &eventData{event: tt.event}})
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// One of works for the other targets too, like a few channels.
func TestFilterEventOneOfChannels(t *testing.T) {
	node := filtered(EventFilterTypeChannelID, ComparsionModeOneOf, "5,6")
	for id, want := range map[discord.ChannelID]bool{5: true, 6: true, 7: false} {
		e := &gateway.MessageCreateEvent{Message: discord.Message{ChannelID: id, GuildID: 1}}
		got, err := node.FilterEvent(&FlowContext{Data: &channelData{eventData: eventData{event: e}, channelID: id}})
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("channel %d: got %v, want %v", id, got, want)
		}
	}
}

type channelData struct {
	eventData
	channelID discord.ChannelID
}

func (d *channelData) ChannelID() discord.ChannelID {
	return d.channelID
}
