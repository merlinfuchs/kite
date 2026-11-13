package auditlog

import (
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

func entry(action discord.AuditLogEvent, changes ...discord.AuditLogChange) discord.AuditLogEntry {
	return discord.AuditLogEntry{ActionType: action, Changes: changes}
}

func ch(key, old, new string) discord.AuditLogChange {
	c := discord.AuditLogChange{Key: discord.AuditLogChangeKey(key)}
	if old != "" {
		c.OldValue = []byte(old)
	}
	if new != "" {
		c.NewValue = []byte(new)
	}
	return c
}

func TestAction(t *testing.T) {
	tests := []struct {
		name  string
		entry discord.AuditLogEntry
		want  string
	}{
		{"ban", entry(22), "member_ban_add"},
		{"role update", entry(25, ch("$add", "", `[{"id":"1","name":"Mod"}]`)), "member_role_update"},
		{"nickname", entry(24, ch("nick", `"a"`, `"b"`)), "member_update"},
		{"timeout", entry(24, ch("communication_disabled_until", "", `"2030-01-01T00:00:00+00:00"`)), ActionMemberTimeoutAdd},
		{"timeout changed", entry(24, ch("communication_disabled_until", `"2029-01-01T00:00:00+00:00"`, `"2030-01-01T00:00:00+00:00"`)), ActionMemberTimeoutAdd},
		{"timeout removed", entry(24, ch("communication_disabled_until", `"2030-01-01T00:00:00+00:00"`, "null")), ActionMemberTimeoutRemove},
		{"timeout removed without new value", entry(24, ch("communication_disabled_until", `"2030-01-01T00:00:00+00:00"`, "")), ActionMemberTimeoutRemove},
		{"unknown", entry(250), "250"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Action(tt.entry); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChannelID(t *testing.T) {
	created := discord.AuditLogEntry{ActionType: 10, TargetID: 5}
	if got := ChannelID(created); got != 5 {
		t.Errorf("channel create: got %d, want the target", got)
	}

	deleted := discord.AuditLogEntry{ActionType: 72, TargetID: 3, Options: discord.AuditEntryInfo{ChannelID: 7}}
	if got := ChannelID(deleted); got != 7 {
		t.Errorf("message delete: got %d, want the channel of the message", got)
	}
	if Target(deleted) != TargetUser {
		t.Error("message delete isn't targeted at the author")
	}

	status := discord.AuditLogEntry{ActionType: 192, Options: discord.AuditEntryInfo{ChannelID: 9}}
	if got := ChannelID(status); got != 9 {
		t.Errorf("voice channel status: got %d, want the channel in the options", got)
	}

	if got := ChannelID(discord.AuditLogEntry{ActionType: 22, TargetID: 3}); got != 0 {
		t.Errorf("ban: got %d, want none", got)
	}
}

func TestRoles(t *testing.T) {
	e := entry(25,
		ch("$add", "", `[{"id":"1","name":"Mod"},{"id":"2","name":"VIP"}]`),
		ch("$remove", "", `[{"id":"3","name":"Muted"}]`),
	)

	added := RolesAdded(e)
	if len(added) != 2 || added[0].ID != 1 || added[1].Name != "VIP" {
		t.Errorf("added = %+v", added)
	}
	if removed := RolesRemoved(e); len(removed) != 1 || removed[0].Name != "Muted" {
		t.Errorf("removed = %+v", removed)
	}
	if got := RolesAdded(entry(22)); got != nil {
		t.Errorf("ban added roles %+v", got)
	}
}

func TestTimeoutUntil(t *testing.T) {
	e := entry(24, ch("communication_disabled_until", "", `"2030-01-01T00:00:00+00:00"`))
	want := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if got := TimeoutUntil(e); !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
	if got := TimeoutUntil(entry(24, ch("communication_disabled_until", `"2030-01-01T00:00:00+00:00"`, "null"))); !got.IsZero() {
		t.Errorf("removed timeout ends at %v", got)
	}
}

func TestChanges(t *testing.T) {
	e := entry(11,
		ch("name", `"general"`, `"chat"`),
		ch("topic", "", `"Hello"`),
		ch("nsfw", "false", "true"),
	)
	want := "name: general → chat\ntopic: none → Hello\nnsfw: false → true"
	if got := Changes(e); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	roles := entry(25, ch("$add", "", `[{"id":"1","name":"Mod"},{"id":"2","name":"VIP"}]`))
	if got := Changes(roles); got != "roles added: Mod, VIP" {
		t.Errorf("roles: got %q", got)
	}

	timeout := entry(24, ch("communication_disabled_until", "", `"2030-01-01T00:00:00+00:00"`))
	if got := Changes(timeout); got != "communication_disabled_until: none → <t:1893456000:f>" {
		t.Errorf("timeout: got %q", got)
	}
}
