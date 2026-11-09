package access

import "testing"

func TestAuditRouteFor(t *testing.T) {
	tests := []struct {
		pattern string
		action  string
	}{
		{"PUT /v1/apps/{appID}", "app.update"},
		{"PATCH /v1/apps/{appID}/commands/{commandID}", "command.update"},
		{"PUT /v1/apps/{appID}/commands/{commandID}/enabled", "command.enabled_update"},
		{"DELETE /v1/apps/{appID}/collaborators/invites/{discordUsername}", "collaborator.invite_revoke"},
		{"POST /v1/apps/{appID}/commands/{commandID}/test", ""},
		{"DELETE /v1/apps/{appID}", ""},
		{"POST /v1/apps", ""},
	}

	for _, tt := range tests {
		route, ok := auditRouteFor(tt.pattern)
		if tt.action == "" {
			if ok {
				t.Errorf("%s: got action %q, want none", tt.pattern, route.Action)
			}
			continue
		}
		if !ok || route.Action != tt.action {
			t.Errorf("%s: got action %q, want %q", tt.pattern, route.Action, tt.action)
		}
	}
}
