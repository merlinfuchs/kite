package access

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

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

func TestReadAuditResponse(t *testing.T) {
	res, ok := readAuditResponse(&wire.EventListener{ID: "1", Type: "message_create", Enabled: true})
	if !ok || res.ID != "1" || res.Type != "message_create" || res.Enabled == nil || !*res.Enabled {
		t.Errorf("got %+v, %v", res, ok)
	}

	res, ok = readAuditResponse(&wire.App{ID: "2", Name: "Bot"})
	if !ok || res.ID != "2" || res.Name != "Bot" || res.Description != "" {
		t.Errorf("got %+v, %v", res, ok)
	}

	if _, ok := readAuditResponse(&[]*wire.Command{}); ok {
		t.Error("read a list")
	}
	if _, ok := readAuditResponse((*wire.Command)(nil)); ok {
		t.Error("read nil")
	}
}
