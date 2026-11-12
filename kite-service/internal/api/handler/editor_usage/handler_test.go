package editor_usage

import (
	"testing"
	"time"
)

func TestValidateEditorSessionRequest(t *testing.T) {
	validSessionID := "a12e45f0-7248-49f2-bdac-12a857d2cdb9"
	validRequest := EditorSessionUpsertRequest{
		TargetType:    "command",
		TargetID:      "command-id",
		ActiveSeconds: 30,
		SaveCount:     2,
	}

	tests := []struct {
		name      string
		sessionID string
		request   EditorSessionUpsertRequest
		wantErr   bool
	}{
		{
			name:      "valid report",
			sessionID: validSessionID,
			request:   validRequest,
		},
		{
			name:      "invalid session id",
			sessionID: "not-a-uuid",
			request:   validRequest,
			wantErr:   true,
		},
		{
			name:      "negative active time",
			sessionID: validSessionID,
			request:   EditorSessionUpsertRequest{TargetType: "command", TargetID: "command-id", ActiveSeconds: -1},
			wantErr:   true,
		},
		{
			name:      "active time beyond one month",
			sessionID: validSessionID,
			request: EditorSessionUpsertRequest{
				TargetType:    "command",
				TargetID:      "command-id",
				ActiveSeconds: int((30 * 24 * time.Hour).Seconds()) + 1,
			},
			wantErr: true,
		},
		{
			name:      "negative save count",
			sessionID: validSessionID,
			request:   EditorSessionUpsertRequest{TargetType: "command", TargetID: "command-id", SaveCount: -1},
			wantErr:   true,
		},
		{
			name:      "save count above limit",
			sessionID: validSessionID,
			request: EditorSessionUpsertRequest{
				TargetType: "command",
				TargetID:   "command-id",
				SaveCount:  1_000_001,
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateEditorSessionRequest(test.sessionID, test.request)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateEditorSessionRequest() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
