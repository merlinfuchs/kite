package store

import (
	"context"
	"time"
)

type EditorSession struct {
	ID            string
	AppID         string
	TargetType    string
	TargetID      string
	UserID        string
	ActiveSeconds int
	SaveCount     int
}

type EditorUsage struct {
	TargetType    string     `json:"target_type"`
	TargetID      string     `json:"target_id"`
	ActiveSeconds int64      `json:"active_seconds"`
	SaveCount     int64      `json:"save_count"`
	SessionCount  int64      `json:"session_count"`
	LastEditedAt  *time.Time `json:"last_edited_at"`
}

type EditorSessionStore interface {
	UpsertEditorSession(ctx context.Context, session *EditorSession) error
	EditorUsageByApp(ctx context.Context, appID string, since time.Time) ([]*EditorUsage, error)
}
