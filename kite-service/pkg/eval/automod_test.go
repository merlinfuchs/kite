package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
)

func autoModContext(e gateway.AutoModerationActionExecutionEvent, ruleName string) Context {
	env := NewEventEnv(&guildevent.AutoModAction{AutoModerationActionExecutionEvent: e, RuleName: ruleName}, nil)
	return NewContext(Env{
		"event":   env,
		"user":    env.User,
		"channel": env.Channel,
		"automod": env.AutoMod,
	})
}

func TestAutoModPlaceholders(t *testing.T) {
	tests := map[string]struct {
		event    gateway.AutoModerationActionExecutionEvent
		ruleName string
		template string
		want     string
	}{
		"blocked message": {
			gateway.AutoModerationActionExecutionEvent{
				GuildID: testGuildID, UserID: 8, ChannelID: 21, RuleID: 5,
				Action:          discord.AutoModerationAction{Type: discord.AutoModerationBlockMessage},
				RuleTriggerType: discord.AutoModerationKeyword,
				Content:         "join discord.gg/abc", MatchedKeyword: "*discord.gg*", MatchedContent: "discord.gg",
			},
			"No invite links",
			"{{automod}}|{{user.id}}|{{channel.id}}|{{automod.rule_id}}|{{automod.rule_name}}|{{automod.trigger}}|{{automod.content}}|{{automod.matched_keyword}}|{{automod.matched_content}}",
			"block_message|8|21|5|No invite links|keyword|join discord.gg/abc|*discord.gg*|discord.gg",
		},
		// A blocked message is never posted, so there's no message.
		"no message": {
			gateway.AutoModerationActionExecutionEvent{
				GuildID: testGuildID, UserID: 8, ChannelID: 21,
				Action: discord.AutoModerationAction{Type: discord.AutoModerationBlockMessage},
			},
			"",
			"{{automod.message_id}}|{{automod.rule_name}}|{{automod.alert_message_id}}",
			"||",
		},
		"alert": {
			gateway.AutoModerationActionExecutionEvent{
				GuildID: testGuildID, UserID: 8, ChannelID: 21, MessageID: 30, AlertSystemMessageID: 31,
				Action:          discord.AutoModerationAction{Type: discord.AutoModerationSendAlertMessage},
				RuleTriggerType: discord.AutoModerationMentionSpam,
			},
			"Mention spam",
			"{{automod.action}}|{{automod.trigger}}|{{automod.message_id}}|{{automod.alert_message_id}}",
			"send_alert_message|mention_spam|30|31",
		},
		"timeout": {
			gateway.AutoModerationActionExecutionEvent{
				GuildID: testGuildID, UserID: 8, ChannelID: 21,
				Action:          discord.AutoModerationAction{Type: discord.AutoModerationTimeout, Metadata: discord.AutoModerationActionMetadata{DurationSeconds: 600}},
				RuleTriggerType: discord.AutoModerationKeywordPreset,
			},
			"Slurs",
			"{{automod.action}}|{{automod.trigger}}|{{automod.timeout_seconds}}",
			"timeout|keyword_preset|600",
		},
		// Member profiles aren't posted in a channel.
		"member profile": {
			gateway.AutoModerationActionExecutionEvent{
				GuildID: testGuildID, UserID: 8,
				Action:          discord.AutoModerationAction{Type: discord.AutoModerationBlockMemberInteraction},
				RuleTriggerType: discord.AutoModerationMemberProfile,
			},
			"Names",
			"{{automod.action}}|{{automod.trigger}}|{{channel.id}}",
			"block_member_interaction|member_profile|",
		},
	}

	for name, tt := range tests {
		if got := evalString(t, autoModContext(tt.event, tt.ruleName), tt.template); got != tt.want {
			t.Errorf("%s: got %q, want %q", name, got, tt.want)
		}
	}
}
