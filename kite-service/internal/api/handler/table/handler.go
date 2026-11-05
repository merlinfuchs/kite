package table

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type Handler struct{ store store.TableStore }

func NewHandler(tableStore store.TableStore) *Handler { return &Handler{store: tableStore} }

func (h *Handler) List(c *handler.Context) (*wire.TableListResponse, error) {
	tables, err := h.store.TablesByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	result := make(wire.TableListResponse, len(tables))
	for i, table := range tables {
		result[i] = wire.TableToWire(table)
	}
	return &result, nil
}

func (h *Handler) Get(c *handler.Context) (*wire.TableGetResponse, error) {
	table, err := h.store.Table(c.Context(), c.App.ID, c.Param("tableID"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_table", "Table not found")
	}
	if err != nil {
		return nil, err
	}
	return wire.TableToWire(table), nil
}

func (h *Handler) Create(c *handler.Context, req wire.TableCreateRequest) (*wire.TableCreateResponse, error) {
	now := time.Now().UTC()
	table, err := h.store.CreateTable(c.Context(), req.Model(util.UniqueID(), c.App.ID, now))
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}
	if err := h.audit(c, table.ID, "table_create", "", nil, table); err != nil {
		return nil, fmt.Errorf("failed to record table audit: %w", err)
	}
	return wire.TableToWire(table), nil
}

func (h *Handler) Update(c *handler.Context, req wire.TableUpdateRequest) (*wire.TableUpdateResponse, error) {
	table, err := h.store.Table(c.Context(), c.App.ID, c.Param("tableID"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_table", "Table not found")
	}
	if err != nil {
		return nil, err
	}
	updated := req.Model(table.ID, c.App.ID, table.CreatedAt)
	updated.UpdatedAt = time.Now().UTC()
	previous := table
	table, err = h.store.UpdateTable(c.Context(), updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update table: %w", err)
	}
	if err := h.audit(c, table.ID, "table_update", "", previous, table); err != nil {
		return nil, fmt.Errorf("failed to record table audit: %w", err)
	}
	return wire.TableToWire(table), nil
}

func (h *Handler) Delete(c *handler.Context) (*wire.TableDeleteResponse, error) {
	table, err := h.store.Table(c.Context(), c.App.ID, c.Param("tableID"))
	if err != nil {
		return nil, err
	}
	if err := h.store.DeleteTable(c.Context(), c.App.ID, c.Param("tableID")); errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_table", "Table not found")
	} else if err != nil {
		return nil, err
	}
	if err := h.audit(c, table.ID, "table_delete", "", table, nil); err != nil {
		return nil, fmt.Errorf("failed to record table audit: %w", err)
	}
	return &wire.TableDeleteResponse{}, nil
}

func (h *Handler) Rows(c *handler.Context) (*wire.TableRowListResponse, error) {
	rows, err := h.store.TableRows(c.Context(), c.App.ID, c.Param("tableID"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_table", "Table not found")
	}
	if err != nil {
		return nil, err
	}
	result := make(wire.TableRowListResponse, len(rows))
	for i, row := range rows {
		result[i] = wire.TableRowToWire(row)
	}
	return &result, nil
}

func (h *Handler) CreateRow(c *handler.Context, req wire.TableRowRequest) (*wire.TableRowCreateResponse, error) {
	table, err := h.store.Table(c.Context(), c.App.ID, c.Param("tableID"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_table", "Table not found")
	}
	if err != nil {
		return nil, err
	}
	if err := validateRow(table, req.Data); err != nil {
		return nil, err
	}
	row, err := h.store.CreateTableRow(c.Context(), &model.TableRow{ID: util.UniqueID(), TableID: c.Param("tableID"), Data: req.Data, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()})
	if err != nil {
		return nil, fmt.Errorf("failed to create row: %w", err)
	}
	if err := h.audit(c, table.ID, "row_create", row.ID, nil, row); err != nil {
		return nil, fmt.Errorf("failed to record row audit: %w", err)
	}
	return wire.TableRowToWire(row), nil
}

func (h *Handler) UpdateRow(c *handler.Context, req wire.TableRowRequest) (*wire.TableRowUpdateResponse, error) {
	table, err := h.store.Table(c.Context(), c.App.ID, c.Param("tableID"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_table", "Table not found")
	}
	if err != nil {
		return nil, err
	}
	if err := validateRow(table, req.Data); err != nil {
		return nil, err
	}
	previous, err := h.store.TableRows(c.Context(), c.App.ID, c.Param("tableID"))
	if err != nil {
		return nil, err
	}
	var before *model.TableRow
	for _, candidate := range previous {
		if candidate.ID == c.Param("rowID") {
			before = candidate
			break
		}
	}
	if before == nil {
		return nil, handler.ErrNotFound("unknown_row", "Row not found")
	}
	row, err := h.store.UpdateTableRow(c.Context(), &model.TableRow{ID: c.Param("rowID"), TableID: c.Param("tableID"), Data: req.Data, UpdatedAt: time.Now().UTC()})
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_row", "Row not found")
	}
	if err != nil {
		return nil, err
	}
	if err := h.audit(c, table.ID, "row_update", row.ID, before, row); err != nil {
		return nil, fmt.Errorf("failed to record row audit: %w", err)
	}
	return wire.TableRowToWire(row), nil
}

func validateRow(table *model.Table, data map[string]any) error {
	for _, column := range table.Columns {
		value, ok := data[column.ID]
		if !ok || value == nil || value == "" {
			if column.Required {
				return handler.ErrBadRequest("invalid_row", fmt.Sprintf("column %q is required", column.Name))
			}
			continue
		}
		valid := false
		switch column.Type {
		case "text", "date":
			_, valid = value.(string)
			if column.Type == "date" && valid {
				_, err := time.Parse(time.RFC3339, value.(string))
				if err != nil {
					_, err = time.Parse("2006-01-02T15:04", value.(string))
				}
				valid = err == nil
			}
		case "number":
			_, valid = value.(float64)
		case "boolean":
			_, valid = value.(bool)
		case "json", "reference":
			valid = true
		}
		if !valid {
			return handler.ErrBadRequest("invalid_row", fmt.Sprintf("column %q must contain a %s value", column.Name, column.Type))
		}
	}
	return nil
}

func (h *Handler) DeleteRow(c *handler.Context) (*wire.TableRowDeleteResponse, error) {
	rows, err := h.store.TableRows(c.Context(), c.App.ID, c.Param("tableID"))
	if err != nil {
		return nil, err
	}
	var before *model.TableRow
	for _, row := range rows {
		if row.ID == c.Param("rowID") {
			before = row
			break
		}
	}
	if err := h.store.DeleteTableRow(c.Context(), c.Param("tableID"), c.Param("rowID")); errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_row", "Row not found")
	} else if err != nil {
		return nil, err
	}
	if before != nil {
		if err := h.audit(c, c.Param("tableID"), "row_delete", before.ID, before, nil); err != nil {
			return nil, fmt.Errorf("failed to record row audit: %w", err)
		}
	}
	return &wire.TableRowDeleteResponse{}, nil
}

func (h *Handler) Audits(c *handler.Context) (*wire.TableAuditListResponse, error) {
	audits, err := h.store.TableAudits(c.Context(), c.App.ID, c.Param("tableID"), 100)
	if err != nil {
		return nil, err
	}
	result := make(wire.TableAuditListResponse, len(audits))
	for i, audit := range audits {
		result[i] = auditToWire(audit)
	}
	return &result, nil
}

func (h *Handler) Undo(c *handler.Context) (*wire.TableAuditUndoResponse, error) {
	audit, err := h.store.TableAudit(c.Context(), c.App.ID, c.Param("tableID"), c.Param("auditID"))
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("unknown_audit", "Audit entry not found")
	}
	if err != nil {
		return nil, err
	}
	if audit.UndoneAt != nil {
		return nil, handler.ErrConflict("audit_already_undone", "This audit entry has already been undone")
	}
	if audit.Action == "row_create" {
		err = h.store.DeleteTableRow(c.Context(), audit.TableID, audit.RowID)
	} else if audit.Action == "row_update" {
		var row model.TableRow
		err = json.Unmarshal(audit.BeforeData, &row)
		if err == nil {
			_, err = h.store.UpdateTableRow(c.Context(), &row)
		}
	} else if audit.Action == "row_delete" {
		var row model.TableRow
		err = json.Unmarshal(audit.BeforeData, &row)
		if err == nil {
			_, err = h.store.CreateTableRow(c.Context(), &row)
		}
	} else if audit.Action == "table_update" {
		var table model.Table
		err = json.Unmarshal(audit.BeforeData, &table)
		if err == nil {
			_, err = h.store.UpdateTable(c.Context(), &table)
		}
	} else {
		return nil, handler.ErrBadRequest("audit_not_undoable", "This audit entry cannot be undone")
	}
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := h.store.MarkTableAuditUndone(c.Context(), c.App.ID, audit.TableID, audit.ID, now); errors.Is(err, store.ErrOutdated) {
		return nil, handler.ErrConflict("audit_already_undone", "This audit entry has already been undone")
	} else if err != nil {
		return nil, err
	}
	undo := &model.TableAudit{ID: util.UniqueID(), AppID: c.App.ID, TableID: audit.TableID, ActorUserID: c.Session.UserID, Action: "undo_" + audit.Action, RowID: audit.RowID, BeforeData: audit.AfterData, AfterData: audit.BeforeData, CreatedAt: now}
	if err := h.store.CreateTableAudit(c.Context(), undo); err != nil {
		return nil, err
	}
	return auditToWire(undo), nil
}

func (h *Handler) audit(c *handler.Context, tableID, action, rowID string, before, after any) error {
	beforeData, err := json.Marshal(before)
	if err != nil {
		return err
	}
	afterData, err := json.Marshal(after)
	if err != nil {
		return err
	}
	return h.store.CreateTableAudit(c.Context(), &model.TableAudit{ID: util.UniqueID(), AppID: c.App.ID, TableID: tableID, ActorUserID: c.Session.UserID, Action: action, RowID: rowID, BeforeData: beforeData, AfterData: afterData, CreatedAt: time.Now().UTC()})
}

func auditToWire(audit *model.TableAudit) *wire.TableAudit {
	var before, after any
	_ = json.Unmarshal(audit.BeforeData, &before)
	_ = json.Unmarshal(audit.AfterData, &after)
	return &wire.TableAudit{ID: audit.ID, Action: audit.Action, RowID: audit.RowID, BeforeData: before, AfterData: after, ActorUserID: audit.ActorUserID, CreatedAt: audit.CreatedAt, UndoneAt: audit.UndoneAt}
}
