package guildevent

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func TestNewKnowsEveryEvent(t *testing.T) {
	for _, e := range []Event{&MessageDelete{}, &AutoModAction{}} {
		got := New(e.EventType())
		if got == nil || got.EventType() != e.EventType() {
			t.Errorf("New(%q) = %T", e.EventType(), got)
		}
	}
	if New("GUILD_BAN_ADD") != nil {
		t.Error("New returned an event for a Discord event type")
	}
}

func TestAutoModActionNames(t *testing.T) {
	tests := []struct {
		action  discord.AutoModerationActionType
		trigger discord.AutoModerationTriggerType
		want    string
	}{
		{discord.AutoModerationBlockMessage, discord.AutoModerationKeyword, "block_message keyword"},
		{discord.AutoModerationSendAlertMessage, discord.AutoModerationSpam, "send_alert_message spam"},
		{discord.AutoModerationTimeout, discord.AutoModerationKeywordPreset, "timeout keyword_preset"},
		{discord.AutoModerationBlockMemberInteraction, discord.AutoModerationMemberProfile, "block_member_interaction member_profile"},
		{discord.AutoModerationBlockMessage, discord.AutoModerationMentionSpam, "block_message mention_spam"},
		{9, 9, "9 9"},
	}

	for _, tt := range tests {
		e := &AutoModAction{AutoModerationActionExecutionEvent: gateway.AutoModerationActionExecutionEvent{
			Action:          discord.AutoModerationAction{Type: tt.action},
			RuleTriggerType: tt.trigger,
		}}
		if got := e.ActionName() + " " + e.TriggerName(); got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}
