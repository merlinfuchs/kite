// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: deployments.sql

package pgmodel

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/sqlc-dev/pqtype"
)

const upsertDeployment = `-- name: UpsertDeployment :one
INSERT INTO deployments (
    id,
    key, 
    name, 
    description, 
    guild_id, 
    plugin_version_id, 
    wasm_bytes, 
    manifest_default_config, 
    manifest_events, 
    manifest_commands, 
    config, 
    created_at, 
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13
) ON CONFLICT (key, guild_id) DO UPDATE SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    plugin_version_id = EXCLUDED.plugin_version_id,
    wasm_bytes = EXCLUDED.wasm_bytes,
    manifest_default_config = EXCLUDED.manifest_default_config,
    manifest_events = EXCLUDED.manifest_events,
    manifest_commands = EXCLUDED.manifest_commands,
    config = EXCLUDED.config,
    updated_at = EXCLUDED.updated_at
RETURNING id, key, name, description, guild_id, plugin_version_id, wasm_bytes, manifest_default_config, manifest_events, manifest_commands, config, created_at, updated_at
`

type UpsertDeploymentParams struct {
	ID                    string
	Key                   string
	Name                  string
	Description           string
	GuildID               string
	PluginVersionID       sql.NullString
	WasmBytes             []byte
	ManifestDefaultConfig pqtype.NullRawMessage
	ManifestEvents        []string
	ManifestCommands      []string
	Config                pqtype.NullRawMessage
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (q *Queries) UpsertDeployment(ctx context.Context, arg UpsertDeploymentParams) (Deployment, error) {
	row := q.db.QueryRowContext(ctx, upsertDeployment,
		arg.ID,
		arg.Key,
		arg.Name,
		arg.Description,
		arg.GuildID,
		arg.PluginVersionID,
		arg.WasmBytes,
		arg.ManifestDefaultConfig,
		pq.Array(arg.ManifestEvents),
		pq.Array(arg.ManifestCommands),
		arg.Config,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i Deployment
	err := row.Scan(
		&i.ID,
		&i.Key,
		&i.Name,
		&i.Description,
		&i.GuildID,
		&i.PluginVersionID,
		&i.WasmBytes,
		&i.ManifestDefaultConfig,
		pq.Array(&i.ManifestEvents),
		pq.Array(&i.ManifestCommands),
		&i.Config,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
