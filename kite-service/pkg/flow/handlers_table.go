package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionTableInsert:       executeActionTableInsert,
		FlowNodeTypeActionTableGet:          executeActionTableGet,
		FlowNodeTypeActionTableFind:         executeActionTableFind,
		FlowNodeTypeActionTableUpdate:       executeActionTableUpdate,
		FlowNodeTypeActionTableDelete:       executeActionTableDelete,
		FlowNodeTypeActionTableColumns:      executeActionTableColumns,
		FlowNodeTypeActionTableColumnAdd:    executeActionTableColumnAdd,
		FlowNodeTypeActionTableColumnUpdate: executeActionTableColumnUpdate,
		FlowNodeTypeActionTableColumnDelete: executeActionTableColumnDelete,
	})
}

func tableID(n *CompiledFlowNode, ctx *FlowContext) (string, error) {
	value, err := ctx.EvalTemplate(n.Data.TableID)
	if err != nil {
		return "", err
	}
	if value.String() == "" {
		return "", errors.New("table ID is required")
	}
	return value.String(), nil
}

func tableJSON(n *CompiledFlowNode, ctx *FlowContext, value string) (map[string]any, error) {
	evaluated, err := ctx.EvalTemplate(value)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(evaluated.String()), &data); err != nil {
		return nil, fmt.Errorf("table data must be a JSON object: %w", err)
	}
	return data, nil
}

func tableResult(rowID string, data map[string]any) thing.Thing {
	result := map[string]thing.Thing{"id": thing.NewString(rowID)}
	for key, value := range data {
		result[key] = thing.NewAny(value)
	}
	return thing.NewObject(result)
}

func executeActionTableInsert(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	data, err := tableJSON(n, ctx, n.Data.TableData)
	if err != nil {
		return err
	}
	row, err := ctx.Table.InsertRow(ctx, id, data)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, tableResult(row.ID, row.Data))
	return n.ExecuteChildren(ctx)
}

func executeActionTableGet(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	rowID, err := ctx.EvalTemplate(n.Data.TableRowID)
	if err != nil {
		return err
	}
	row, err := ctx.Table.GetRow(ctx, id, rowID.String())
	if errors.Is(err, provider.ErrNotFound) {
		ctx.StoreNodeResult(n, thing.Null)
		return n.ExecuteChildren(ctx)
	}
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, tableResult(row.ID, row.Data))
	return n.ExecuteChildren(ctx)
}

func executeActionTableFind(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	evaluated, err := ctx.EvalTemplate(n.Data.TableFilters)
	if err != nil {
		return err
	}
	var filters []provider.TableFilter
	if evaluated.String() != "" {
		if err := json.Unmarshal([]byte(evaluated.String()), &filters); err != nil {
			return fmt.Errorf("table filters must be a JSON array: %w", err)
		}
	}
	limit := 100
	if n.Data.TableLimit != "" {
		value, err := ctx.EvalTemplate(n.Data.TableLimit)
		if err != nil {
			return err
		}
		limit, err = strconv.Atoi(value.String())
		if err != nil || limit < 1 || limit > 100 {
			return errors.New("table limit must be between 1 and 100")
		}
	}
	rows, err := ctx.Table.FindRows(ctx, id, filters, limit)
	if err != nil {
		return err
	}
	result := make([]thing.Thing, len(rows))
	for i, row := range rows {
		result[i] = tableResult(row.ID, row.Data)
	}
	ctx.StoreNodeResult(n, thing.NewArray(result))
	return n.ExecuteChildren(ctx)
}

func executeActionTableUpdate(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	rowID, err := ctx.EvalTemplate(n.Data.TableRowID)
	if err != nil {
		return err
	}
	data, err := tableJSON(n, ctx, n.Data.TableData)
	if err != nil {
		return err
	}
	row, err := ctx.Table.UpdateRow(ctx, id, rowID.String(), data)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, tableResult(row.ID, row.Data))
	return n.ExecuteChildren(ctx)
}

func executeActionTableDelete(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	rowID, err := ctx.EvalTemplate(n.Data.TableRowID)
	if err != nil {
		return err
	}
	if err := ctx.Table.DeleteRow(ctx, id, rowID.String()); err != nil {
		return err
	}
	return n.ExecuteChildren(ctx)
}

func tableColumnResult(column provider.TableColumn) thing.Thing {
	values := map[string]thing.Thing{
		"id": thing.NewString(column.ID), "name": thing.NewString(column.Name),
		"type": thing.NewString(column.Type), "required": thing.NewBool(column.Required),
	}
	return thing.NewObject(values)
}

func evalTableColumn(n *CompiledFlowNode, ctx *FlowContext) (provider.TableColumn, error) {
	name, err := ctx.EvalTemplate(n.Data.TableColumnName)
	if err != nil {
		return provider.TableColumn{}, err
	}
	typ, err := ctx.EvalTemplate(n.Data.TableColumnType)
	if err != nil {
		return provider.TableColumn{}, err
	}
	columnID, err := ctx.EvalTemplate(n.Data.TableColumnID)
	if err != nil {
		return provider.TableColumn{}, err
	}
	column := provider.TableColumn{ID: columnID.String(), Name: name.String(), Type: typ.String()}
	if column.ID == "" {
		column.ID = name.String()
	}
	if column.ID == "" || column.Name == "" || column.Type == "" {
		return provider.TableColumn{}, errors.New("column ID, name, and type are required")
	}
	switch column.Type {
	case "text", "number", "boolean", "date", "json", "reference":
	default:
		return provider.TableColumn{}, errors.New("column type is invalid")
	}
	return column, nil
}

func executeActionTableColumns(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	columns, err := ctx.Table.ListColumns(ctx, id)
	if err != nil {
		return err
	}
	result := make([]thing.Thing, len(columns))
	for i, column := range columns {
		result[i] = tableColumnResult(column)
	}
	ctx.StoreNodeResult(n, thing.NewArray(result))
	return n.ExecuteChildren(ctx)
}

func executeActionTableColumnAdd(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	column, err := evalTableColumn(n, ctx)
	if err != nil {
		return err
	}
	column, err = ctx.Table.AddColumn(ctx, id, column)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, tableColumnResult(column))
	return n.ExecuteChildren(ctx)
}

func executeActionTableColumnUpdate(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	column, err := evalTableColumn(n, ctx)
	if err != nil {
		return err
	}
	columnID, err := ctx.EvalTemplate(n.Data.TableColumnID)
	if err != nil {
		return err
	}
	updated, err := ctx.Table.UpdateColumn(ctx, id, columnID.String(), column)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, tableColumnResult(updated))
	return n.ExecuteChildren(ctx)
}

func executeActionTableColumnDelete(n *CompiledFlowNode, ctx *FlowContext) error {
	id, err := tableID(n, ctx)
	if err != nil {
		return err
	}
	columnID, err := ctx.EvalTemplate(n.Data.TableColumnID)
	if err != nil {
		return err
	}
	if columnID.String() == "" {
		return errors.New("column ID is required")
	}
	if err := ctx.Table.DeleteColumn(ctx, id, columnID.String()); err != nil {
		return err
	}
	return n.ExecuteChildren(ctx)
}
