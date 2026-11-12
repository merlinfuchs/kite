package eval

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func auditLogContext(entry discord.AuditLogEntry) Context {
	env := NewEventEnv(&gateway.GuildAuditLogEntryCreateEvent{AuditLogEntry: entry, GuildID: testGuildID}, nil)
	return NewContext(Env{
		"event":     env,
		"user":      env.User,
		"channel":   env.Channel,
		"audit_log": env.AuditLog,
	})
}

func TestAuditLogPlaceholders(t *testing.T) {
	tests := map[string]struct {
		entry    discord.AuditLogEntry
		template string
		want     string
	}{
		"ban": {
			discord.AuditLogEntry{ID: 1, ActionType: discord.MemberBanAdd, UserID: 7, TargetID: 8, Reason: "spam"},
			"{{audit_log}}|{{user.id}}|{{audit_log.target}}|{{audit_log.target_id}}|{{audit_log.reason}}",
			"member_ban_add|7|<@8>|8|spam",
		},
		"roles": {
			discord.AuditLogEntry{ID: 2, ActionType: discord.MemberRoleUpdate, UserID: 7, TargetID: 8, Changes: []discord.AuditLogChange{
				{Key: "$add", NewValue: []byte(`[{"id":"1","name":"Mod"},{"id":"2","name":"VIP"}]`)},
				{Key: "$remove", NewValue: []byte(`[{"id":"3","name":"Muted"}]`)},
			}},
			"{{audit_log.action}}|{{audit_log.roles_added}}|{{audit_log.roles_removed}}",
			"member_role_update|Mod, VIP|Muted",
		},
		"timeout": {
			discord.AuditLogEntry{ID: 3, ActionType: discord.MemberUpdate, UserID: 7, TargetID: 8, Changes: []discord.AuditLogChange{
				{Key: "communication_disabled_until", NewValue: []byte(`"2030-01-01T00:00:00+00:00"`)},
			}},
			"{{audit_log.action}}|{{audit_log.timeout_until}}",
			"member_timeout_add|1893456000",
		},
		"channel": {
			discord.AuditLogEntry{ID: 4, ActionType: discord.ChannelUpdate, UserID: 7, TargetID: 21, Changes: []discord.AuditLogChange{
				{Key: "name", OldValue: []byte(`"general"`), NewValue: []byte(`"chat"`)},
			}},
			"{{audit_log.target}}|{{channel.id}}|{{audit_log.changes}}",
			"<#21>|21|name: general → chat",
		},
		// Pruning has no target.
		"prune": {
			discord.AuditLogEntry{ID: 5, ActionType: discord.MemberPrune, UserID: 7},
			"{{audit_log.action}}|{{audit_log.target}}|{{audit_log.timeout_until}}",
			"member_prune||0",
		},
	}

	for name, tt := range tests {
		if got := evalString(t, auditLogContext(tt.entry), tt.template); got != tt.want {
			t.Errorf("%s: got %q, want %q", name, got, tt.want)
		}
	}
}
