// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: variables.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countVariablesByApp = `-- name: CountVariablesByApp :one
SELECT COUNT(*) FROM variables WHERE app_id = $1
`

func (q *Queries) CountVariablesByApp(ctx context.Context, appID string) (int64, error) {
	row := q.db.QueryRow(ctx, countVariablesByApp, appID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createVariable = `-- name: CreateVariable :one
INSERT INTO variables (
    id,
    name,
    scoped,
    app_id,
    module_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, name, scoped, app_id, module_id, created_at, updated_at
`

type CreateVariableParams struct {
	ID        string
	Name      string
	Scoped    bool
	AppID     string
	ModuleID  pgtype.Text
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

func (q *Queries) CreateVariable(ctx context.Context, arg CreateVariableParams) (Variable, error) {
	row := q.db.QueryRow(ctx, createVariable,
		arg.ID,
		arg.Name,
		arg.Scoped,
		arg.AppID,
		arg.ModuleID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i Variable
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Scoped,
		&i.AppID,
		&i.ModuleID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteAllVariableValues = `-- name: DeleteAllVariableValues :exec
DELETE FROM variable_values WHERE variable_id = $1
`

func (q *Queries) DeleteAllVariableValues(ctx context.Context, variableID string) error {
	_, err := q.db.Exec(ctx, deleteAllVariableValues, variableID)
	return err
}

const deleteVariable = `-- name: DeleteVariable :exec
DELETE FROM variables WHERE id = $1
`

func (q *Queries) DeleteVariable(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteVariable, id)
	return err
}

const deleteVariableValue = `-- name: DeleteVariableValue :exec
DELETE FROM variable_values WHERE variable_id = $1 AND scope = $2
`

type DeleteVariableValueParams struct {
	VariableID string
	Scope      pgtype.Text
}

func (q *Queries) DeleteVariableValue(ctx context.Context, arg DeleteVariableValueParams) error {
	_, err := q.db.Exec(ctx, deleteVariableValue, arg.VariableID, arg.Scope)
	return err
}

const getVariable = `-- name: GetVariable :one
SELECT variables.id, variables.name, variables.scoped, variables.app_id, variables.module_id, variables.created_at, variables.updated_at, COUNT(variable_values.*) as total_values FROM variables 
LEFT JOIN variable_values ON variables.id = variable_values.variable_id
WHERE variables.id = $1
GROUP BY variables.id
`

type GetVariableRow struct {
	Variable    Variable
	TotalValues int64
}

func (q *Queries) GetVariable(ctx context.Context, id string) (GetVariableRow, error) {
	row := q.db.QueryRow(ctx, getVariable, id)
	var i GetVariableRow
	err := row.Scan(
		&i.Variable.ID,
		&i.Variable.Name,
		&i.Variable.Scoped,
		&i.Variable.AppID,
		&i.Variable.ModuleID,
		&i.Variable.CreatedAt,
		&i.Variable.UpdatedAt,
		&i.TotalValues,
	)
	return i, err
}

const getVariableByName = `-- name: GetVariableByName :one
SELECT variables.id, variables.name, variables.scoped, variables.app_id, variables.module_id, variables.created_at, variables.updated_at, COUNT(variable_values.*) as total_values FROM variables 
LEFT JOIN variable_values ON variables.id = variable_values.variable_id
WHERE app_id = $1 AND name = $2
GROUP BY variables.id
`

type GetVariableByNameParams struct {
	AppID string
	Name  string
}

type GetVariableByNameRow struct {
	Variable    Variable
	TotalValues int64
}

func (q *Queries) GetVariableByName(ctx context.Context, arg GetVariableByNameParams) (GetVariableByNameRow, error) {
	row := q.db.QueryRow(ctx, getVariableByName, arg.AppID, arg.Name)
	var i GetVariableByNameRow
	err := row.Scan(
		&i.Variable.ID,
		&i.Variable.Name,
		&i.Variable.Scoped,
		&i.Variable.AppID,
		&i.Variable.ModuleID,
		&i.Variable.CreatedAt,
		&i.Variable.UpdatedAt,
		&i.TotalValues,
	)
	return i, err
}

const getVariableValue = `-- name: GetVariableValue :one
SELECT id, variable_id, scope, value, created_at, updated_at FROM variable_values WHERE variable_id = $1 AND scope IS NOT DISTINCT FROM $2
`

type GetVariableValueParams struct {
	VariableID string
	Scope      pgtype.Text
}

func (q *Queries) GetVariableValue(ctx context.Context, arg GetVariableValueParams) (VariableValue, error) {
	row := q.db.QueryRow(ctx, getVariableValue, arg.VariableID, arg.Scope)
	var i VariableValue
	err := row.Scan(
		&i.ID,
		&i.VariableID,
		&i.Scope,
		&i.Value,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getVariableValueForUpdate = `-- name: GetVariableValueForUpdate :one
SELECT id, variable_id, scope, value, created_at, updated_at FROM variable_values WHERE variable_id = $1 AND scope IS NOT DISTINCT FROM $2 FOR UPDATE
`

type GetVariableValueForUpdateParams struct {
	VariableID string
	Scope      pgtype.Text
}

func (q *Queries) GetVariableValueForUpdate(ctx context.Context, arg GetVariableValueForUpdateParams) (VariableValue, error) {
	row := q.db.QueryRow(ctx, getVariableValueForUpdate, arg.VariableID, arg.Scope)
	var i VariableValue
	err := row.Scan(
		&i.ID,
		&i.VariableID,
		&i.Scope,
		&i.Value,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getVariableValues = `-- name: GetVariableValues :many
SELECT id, variable_id, scope, value, created_at, updated_at FROM variable_values WHERE variable_id = $1
`

func (q *Queries) GetVariableValues(ctx context.Context, variableID string) ([]VariableValue, error) {
	rows, err := q.db.Query(ctx, getVariableValues, variableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []VariableValue
	for rows.Next() {
		var i VariableValue
		if err := rows.Scan(
			&i.ID,
			&i.VariableID,
			&i.Scope,
			&i.Value,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getVariablesByApp = `-- name: GetVariablesByApp :many
SELECT variables.id, variables.name, variables.scoped, variables.app_id, variables.module_id, variables.created_at, variables.updated_at, COUNT(variable_values.*) as total_values FROM variables 
LEFT JOIN variable_values ON variables.id = variable_values.variable_id
WHERE variables.app_id = $1 
GROUP BY variables.id
ORDER BY variables.created_at DESC
`

type GetVariablesByAppRow struct {
	Variable    Variable
	TotalValues int64
}

func (q *Queries) GetVariablesByApp(ctx context.Context, appID string) ([]GetVariablesByAppRow, error) {
	rows, err := q.db.Query(ctx, getVariablesByApp, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetVariablesByAppRow
	for rows.Next() {
		var i GetVariablesByAppRow
		if err := rows.Scan(
			&i.Variable.ID,
			&i.Variable.Name,
			&i.Variable.Scoped,
			&i.Variable.AppID,
			&i.Variable.ModuleID,
			&i.Variable.CreatedAt,
			&i.Variable.UpdatedAt,
			&i.TotalValues,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setVariableValue = `-- name: SetVariableValue :one
INSERT INTO variable_values (
    variable_id,
    scope,
    value,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5
) ON CONFLICT (variable_id, scope) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = EXCLUDED.updated_at
RETURNING id, variable_id, scope, value, created_at, updated_at
`

type SetVariableValueParams struct {
	VariableID string
	Scope      pgtype.Text
	Value      []byte
	CreatedAt  pgtype.Timestamp
	UpdatedAt  pgtype.Timestamp
}

func (q *Queries) SetVariableValue(ctx context.Context, arg SetVariableValueParams) (VariableValue, error) {
	row := q.db.QueryRow(ctx, setVariableValue,
		arg.VariableID,
		arg.Scope,
		arg.Value,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i VariableValue
	err := row.Scan(
		&i.ID,
		&i.VariableID,
		&i.Scope,
		&i.Value,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateVariable = `-- name: UpdateVariable :one
UPDATE variables SET
    name = $2,
    scoped = $3,
    updated_at = $4
WHERE id = $1 RETURNING id, name, scoped, app_id, module_id, created_at, updated_at
`

type UpdateVariableParams struct {
	ID        string
	Name      string
	Scoped    bool
	UpdatedAt pgtype.Timestamp
}

func (q *Queries) UpdateVariable(ctx context.Context, arg UpdateVariableParams) (Variable, error) {
	row := q.db.QueryRow(ctx, updateVariable,
		arg.ID,
		arg.Name,
		arg.Scoped,
		arg.UpdatedAt,
	)
	var i Variable
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Scoped,
		&i.AppID,
		&i.ModuleID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
