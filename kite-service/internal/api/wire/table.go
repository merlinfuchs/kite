package wire

import (
	"errors"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

type TableColumn struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Required         bool   `json:"required,omitempty"`
	Default          any    `json:"default,omitempty"`
	ReferenceTableID string `json:"reference_table_id,omitempty"`
}

type Table struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Columns   []TableColumn `json:"columns"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

type TableRow struct {
	ID        string         `json:"id"`
	TableID   string         `json:"table_id"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type TableListResponse = []*Table
type TableGetResponse = Table

type TableCreateRequest struct {
	Name    string        `json:"name"`
	Columns []TableColumn `json:"columns"`
}

type TableUpdateRequest = TableCreateRequest
type TableCreateResponse = Table
type TableUpdateResponse = Table
type TableDeleteResponse = Empty

type TableRowListResponse = []*TableRow

type TableRowRequest struct {
	Data map[string]any `json:"data"`
}

type TableRowCreateResponse = TableRow
type TableRowUpdateResponse = TableRow
type TableRowDeleteResponse = Empty

type TableAudit struct {
	ID          string     `json:"id"`
	Action      string     `json:"action"`
	RowID       string     `json:"row_id,omitempty"`
	BeforeData  any        `json:"before,omitempty"`
	AfterData   any        `json:"after,omitempty"`
	ActorUserID string     `json:"actor_user_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UndoneAt    *time.Time `json:"undone_at,omitempty"`
}

type TableAuditListResponse = []*TableAudit
type TableAuditUndoResponse = TableAudit

var tableNameRegex = variableNameRegex

func (req TableCreateRequest) Validate() error {
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100), validation.Match(tableNameRegex)),
		validation.Field(&req.Columns, validation.Required, validation.Length(1, 50)),
	); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, column := range req.Columns {
		if err := validation.ValidateStruct(&column,
			validation.Field(&column.ID, validation.Required, validation.Length(1, 50)),
			validation.Field(&column.Name, validation.Required, validation.Length(1, 100)),
			validation.Field(&column.Type, validation.Required, validation.In("text", "number", "boolean", "date", "json", "reference")),
		); err != nil {
			return err
		}
		if seen[column.ID] {
			return errors.New("column IDs must be unique")
		}
		if column.Type != "reference" && column.ReferenceTableID != "" {
			return errors.New("only reference columns can reference another table")
		}
		if column.Type == "reference" && column.ReferenceTableID == "" {
			return errors.New("reference columns must select a table")
		}
		seen[column.ID] = true
	}
	return nil
}

func (req TableRowRequest) Validate() error {
	if req.Data == nil {
		return errors.New("data is required")
	}
	return nil
}

func (req TableCreateRequest) Model(id, appID string, now time.Time) *model.Table {
	columns := make([]model.TableColumn, len(req.Columns))
	for i, column := range req.Columns {
		columns[i] = model.TableColumn{
			ID: column.ID, Name: strings.TrimSpace(column.Name), Type: column.Type,
			Required: column.Required, Default: column.Default,
			ReferenceTableID: column.ReferenceTableID,
		}
	}
	return &model.Table{ID: id, AppID: appID, Name: strings.TrimSpace(req.Name), Columns: columns, CreatedAt: now, UpdatedAt: now}
}

func TableToWire(table *model.Table) *Table {
	columns := make([]TableColumn, len(table.Columns))
	for i, column := range table.Columns {
		columns[i] = TableColumn{
			ID: column.ID, Name: column.Name, Type: column.Type,
			Required: column.Required, Default: column.Default,
			ReferenceTableID: column.ReferenceTableID,
		}

	}
	return &Table{ID: table.ID, Name: table.Name, Columns: columns, CreatedAt: table.CreatedAt, UpdatedAt: table.UpdatedAt}
}

func TableRowToWire(row *model.TableRow) *TableRow {
	return &TableRow{ID: row.ID, TableID: row.TableID, Data: row.Data, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}
}
