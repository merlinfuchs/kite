package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (c *Client) UpsertEditorSession(ctx context.Context, session *store.EditorSession) error {
	tag, err := c.DB.Exec(ctx, `
		INSERT INTO editor_sessions (
			id, app_id, target_type, target_id, user_id,
			active_seconds, save_count, created_at, last_edited_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), CASE WHEN $7 > 0 THEN NOW() ELSE NULL END, NOW())
		ON CONFLICT (id) DO UPDATE SET
			active_seconds = GREATEST(editor_sessions.active_seconds, EXCLUDED.active_seconds),
			save_count = GREATEST(editor_sessions.save_count, EXCLUDED.save_count),
			last_edited_at = CASE
				WHEN EXCLUDED.save_count > editor_sessions.save_count THEN NOW()
				ELSE editor_sessions.last_edited_at
			END,
			updated_at = NOW()
		WHERE editor_sessions.app_id = EXCLUDED.app_id
			AND editor_sessions.target_type = EXCLUDED.target_type
			AND editor_sessions.target_id = EXCLUDED.target_id
			AND editor_sessions.user_id = EXCLUDED.user_id
	`,
		session.ID,
		session.AppID,
		session.TargetType,
		session.TargetID,
		session.UserID,
		session.ActiveSeconds,
		session.SaveCount,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert editor session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrAlreadyExists
	}
	return nil
}

func (c *Client) EditorUsageByApp(
	ctx context.Context,
	appID string,
	since time.Time,
) ([]*store.EditorUsage, error) {
	rows, err := c.DB.Query(ctx, `
		SELECT
			target_type,
			target_id,
			SUM(active_seconds),
			SUM(save_count),
			COUNT(*),
			MAX(last_edited_at)
		FROM editor_sessions
		WHERE app_id = $1 AND created_at >= $2
		GROUP BY target_type, target_id
		ORDER BY target_type, target_id
	`, appID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get editor usage for app: %w", err)
	}
	defer rows.Close()

	usage := make([]*store.EditorUsage, 0)
	for rows.Next() {
		item := &store.EditorUsage{}
		var lastEditedAt pgtype.Timestamp
		if err := rows.Scan(
			&item.TargetType,
			&item.TargetID,
			&item.ActiveSeconds,
			&item.SaveCount,
			&item.SessionCount,
			&lastEditedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan app editor usage: %w", err)
		}
		if lastEditedAt.Valid {
			item.LastEditedAt = &lastEditedAt.Time
		}
		usage = append(usage, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read app editor usage: %w", err)
	}
	return usage, nil
}
