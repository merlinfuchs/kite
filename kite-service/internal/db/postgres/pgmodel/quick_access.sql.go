// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: quick_access.sql

package pgmodel

import (
	"context"
	"time"
)

const getQuickAccessItems = `-- name: GetQuickAccessItems :many
SELECT id, guild_id, name, updated_at, 'DEPLOYMENT' as type FROM deployments WHERE deployments.guild_id = $1
UNION ALL 
SELECT id, guild_id, name, updated_at, 'WORKSPACE' as type FROM workspaces WHERE workspaces.guild_id = $1
ORDER BY updated_at DESC
LIMIT $2
`

type GetQuickAccessItemsParams struct {
	GuildID string
	Limit   int32
}

type GetQuickAccessItemsRow struct {
	ID        string
	GuildID   string
	Name      string
	UpdatedAt time.Time
	Type      string
}

func (q *Queries) GetQuickAccessItems(ctx context.Context, arg GetQuickAccessItemsParams) ([]GetQuickAccessItemsRow, error) {
	rows, err := q.db.QueryContext(ctx, getQuickAccessItems, arg.GuildID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetQuickAccessItemsRow
	for rows.Next() {
		var i GetQuickAccessItemsRow
		if err := rows.Scan(
			&i.ID,
			&i.GuildID,
			&i.Name,
			&i.UpdatedAt,
			&i.Type,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}