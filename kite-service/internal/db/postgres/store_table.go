package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (c *Client) TablesByApp(ctx context.Context, appID string) ([]*model.Table, error) {
	rows, err := c.DB.Query(ctx, `SELECT id, app_id, name, columns, created_at, updated_at
		FROM content_tables WHERE app_id = $1 ORDER BY created_at DESC`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.Table
	for rows.Next() {
		table, err := scanTable(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, table)
	}
	return result, rows.Err()
}

func (c *Client) Table(ctx context.Context, appID, id string) (*model.Table, error) {
	row := c.DB.QueryRow(ctx, `SELECT id, app_id, name, columns, created_at, updated_at
		FROM content_tables WHERE app_id = $1 AND id = $2`, appID, id)
	table, err := scanTable(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	return table, err
}

func (c *Client) CreateTable(ctx context.Context, table *model.Table) (*model.Table, error) {
	columns, err := json.Marshal(table.Columns)
	if err != nil {
		return nil, err
	}
	_, err = c.DB.Exec(ctx, `INSERT INTO content_tables
		(id, app_id, name, columns, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		table.ID, table.AppID, table.Name, columns, table.CreatedAt, table.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c.Table(ctx, table.AppID, table.ID)
}

func (c *Client) UpdateTable(ctx context.Context, table *model.Table) (*model.Table, error) {
	columns, err := json.Marshal(table.Columns)
	if err != nil {
		return nil, err
	}
	_, err = c.DB.Exec(ctx, `UPDATE content_tables SET name = $1, columns = $2, updated_at = $3
		WHERE app_id = $4 AND id = $5`, table.Name, columns, table.UpdatedAt, table.AppID, table.ID)
	if err != nil {
		return nil, err
	}
	return c.Table(ctx, table.AppID, table.ID)
}

func (c *Client) DeleteTable(ctx context.Context, appID, id string) error {
	tag, err := c.DB.Exec(ctx, `DELETE FROM content_tables WHERE app_id = $1 AND id = $2`, appID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (c *Client) TableRows(ctx context.Context, appID, tableID string) ([]*model.TableRow, error) {
	if _, err := c.Table(ctx, appID, tableID); err != nil {
		return nil, err
	}
	rows, err := c.DB.Query(ctx, `SELECT id, table_id, data, created_at, updated_at
		FROM content_table_rows WHERE table_id = $1 ORDER BY created_at DESC`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.TableRow
	for rows.Next() {
		row := &model.TableRow{}
		var data []byte
		if err := rows.Scan(&row.ID, &row.TableID, &data, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, &row.Data); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (c *Client) CreateTableRow(ctx context.Context, row *model.TableRow) (*model.TableRow, error) {
	data, err := json.Marshal(row.Data)
	if err != nil {
		return nil, err
	}
	_, err = c.DB.Exec(ctx, `INSERT INTO content_table_rows (id, table_id, data, created_at, updated_at)
		SELECT $1, id, $3, $4, $5 FROM content_tables WHERE id = $2`, row.ID, row.TableID, data, row.CreatedAt, row.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c.tableRow(ctx, row.TableID, row.ID)
}

func (c *Client) UpdateTableRow(ctx context.Context, row *model.TableRow) (*model.TableRow, error) {
	data, err := json.Marshal(row.Data)
	if err != nil {
		return nil, err
	}
	tag, err := c.DB.Exec(ctx, `UPDATE content_table_rows SET data = $1, updated_at = $2
		WHERE table_id = $3 AND id = $4`, data, row.UpdatedAt, row.TableID, row.ID)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, store.ErrNotFound
	}
	return c.tableRow(ctx, row.TableID, row.ID)
}

func (c *Client) DeleteTableRow(ctx context.Context, tableID, id string) error {
	tag, err := c.DB.Exec(ctx, `DELETE FROM content_table_rows WHERE table_id = $1 AND id = $2`, tableID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func (c *Client) CreateTableAudit(ctx context.Context, audit *model.TableAudit) error {
	_, err := c.DB.Exec(ctx, `INSERT INTO content_table_audit
			(id, app_id, table_id, actor_user_id, action, row_id, before_data, after_data, created_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		audit.ID, audit.AppID, audit.TableID, audit.ActorUserID, audit.Action,
		audit.RowID, audit.BeforeData, audit.AfterData, audit.CreatedAt)
	return err
}

func (c *Client) TableAudits(ctx context.Context, appID, tableID string, limit int) ([]*model.TableAudit, error) {
	rows, err := c.DB.Query(ctx, `SELECT id, app_id, table_id, actor_user_id, action, row_id, before_data, after_data, created_at, undone_at
			FROM content_table_audit WHERE app_id = $1 AND table_id = $2
			ORDER BY created_at DESC LIMIT $3`, appID, tableID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*model.TableAudit
	for rows.Next() {
		audit := &model.TableAudit{}
		if err := rows.Scan(&audit.ID, &audit.AppID, &audit.TableID, &audit.ActorUserID, &audit.Action, &audit.RowID, &audit.BeforeData, &audit.AfterData, &audit.CreatedAt, &audit.UndoneAt); err != nil {
			return nil, err
		}
		result = append(result, audit)
	}
	return result, rows.Err()
}

func (c *Client) TableAudit(ctx context.Context, appID, tableID, id string) (*model.TableAudit, error) {
	audit := &model.TableAudit{}
	err := c.DB.QueryRow(ctx, `SELECT id, app_id, table_id, actor_user_id, action, row_id, before_data, after_data, created_at, undone_at
			FROM content_table_audit WHERE app_id = $1 AND table_id = $2 AND id = $3`,
		appID, tableID, id).Scan(&audit.ID, &audit.AppID, &audit.TableID, &audit.ActorUserID,
		&audit.Action, &audit.RowID, &audit.BeforeData, &audit.AfterData, &audit.CreatedAt, &audit.UndoneAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	}
	return audit, err
}

func (c *Client) MarkTableAuditUndone(ctx context.Context, appID, tableID, id string, at time.Time) error {
	tag, err := c.DB.Exec(ctx, `UPDATE content_table_audit SET undone_at = $1
		WHERE app_id = $2 AND table_id = $3 AND id = $4 AND undone_at IS NULL`,
		at, appID, tableID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return store.ErrOutdated
	}
	return nil
}

func (c *Client) tableRow(ctx context.Context, tableID, id string) (*model.TableRow, error) {
	row := c.DB.QueryRow(ctx, `SELECT id, table_id, data, created_at, updated_at
		FROM content_table_rows WHERE table_id = $1 AND id = $2`, tableID, id)
	result := &model.TableRow{}
	var data []byte
	if err := row.Scan(&result.ID, &result.TableID, &data, &result.CreatedAt, &result.UpdatedAt); errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &result.Data); err != nil {
		return nil, err
	}
	return result, nil
}

type rowScanner interface {
	Scan(...any) error
}

func scanTable(row rowScanner) (*model.Table, error) {
	table := &model.Table{}
	var columns []byte
	if err := row.Scan(&table.ID, &table.AppID, &table.Name, &columns, &table.CreatedAt, &table.UpdatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(columns, &table.Columns); err != nil {
		return nil, err
	}
	return table, nil
}
