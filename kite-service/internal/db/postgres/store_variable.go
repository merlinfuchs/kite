package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) VariablesByApp(ctx context.Context, appID string) ([]*model.Variable, error) {
	rows, err := c.Q.GetVariablesByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	variables := make([]*model.Variable, len(rows))
	for i, row := range rows {
		v := rowToVariable(row.Variable)
		v.TotalValues = null.NewInt(row.TotalValues, true)
		variables[i] = v
	}

	return variables, nil
}

func (c *Client) CountVariablesByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountVariablesByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) Variable(ctx context.Context, id string) (*model.Variable, error) {
	row, err := c.Q.GetVariable(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	v := rowToVariable(row.Variable)
	v.TotalValues = null.NewInt(row.TotalValues, true)
	return v, nil
}

func (c *Client) VariableByName(ctx context.Context, appID, name string) (*model.Variable, error) {
	row, err := c.Q.GetVariableByName(ctx, pgmodel.GetVariableByNameParams{
		AppID: appID,
		Name:  name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	v := rowToVariable(row.Variable)
	v.TotalValues = null.NewInt(row.TotalValues, true)
	return v, nil
}

func (c *Client) CreateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error) {
	row, err := c.Q.CreateVariable(ctx, pgmodel.CreateVariableParams{
		ID:     variable.ID,
		Name:   variable.Name,
		Scoped: variable.Scoped,
		AppID:  variable.AppID,
		ModuleID: pgtype.Text{
			String: variable.ModuleID.String,
			Valid:  variable.ModuleID.Valid,
		},
		CreatedAt: pgtype.Timestamp{
			Time:  variable.CreatedAt.UTC(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:  variable.UpdatedAt.UTC(),
			Valid: true,
		},
	})
	if err != nil {
		return nil, err
	}

	return rowToVariable(row), nil
}

func (c *Client) UpdateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error) {
	row, err := c.Q.UpdateVariable(ctx, pgmodel.UpdateVariableParams{
		ID:     variable.ID,
		Name:   variable.Name,
		Scoped: variable.Scoped,
		UpdatedAt: pgtype.Timestamp{
			Time:  variable.UpdatedAt.UTC(),
			Valid: true,
		},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToVariable(row), nil
}

func (c *Client) DeleteVariable(ctx context.Context, id string) error {
	err := c.Q.DeleteVariable(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func rowToVariable(row pgmodel.Variable) *model.Variable {
	return &model.Variable{
		ID:        row.ID,
		Name:      row.Name,
		Scoped:    row.Scoped,
		AppID:     row.AppID,
		ModuleID:  null.NewString(row.ModuleID.String, row.ModuleID.Valid),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}

func (c *Client) VariableValues(ctx context.Context, variableID string) ([]*model.VariableValue, error) {
	rows, err := c.Q.GetVariableValues(ctx, variableID)
	if err != nil {
		return nil, err
	}

	var values []*model.VariableValue
	for _, row := range rows {
		v, err := rowToVariableValue(row)
		if err != nil {
			return nil, err
		}
		values = append(values, &v)
	}

	return values, nil
}

func (c *Client) VariableValue(ctx context.Context, variableID string, scope null.String) (*model.VariableValue, error) {
	row, err := c.Q.GetVariableValue(ctx, pgmodel.GetVariableValueParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: scope.String, Valid: scope.Valid},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	v, err := rowToVariableValue(row)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Client) SetVariableValue(ctx context.Context, value model.VariableValue) error {
	_, err := c.setVariableValueWithTx(ctx, nil, value)
	return err
}

func (c *Client) UpdateVariableValue(ctx context.Context, operation model.VariableValueOperation, value model.VariableValue) (*model.VariableValue, error) {
	if operation == flow.VariableOperationOverwrite {
		return c.setVariableValueWithTx(ctx, nil, value)
	}

	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	currentValue, err := c.variableValueWithTx(ctx, tx, value.VariableID, value.Scope)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// Current trasaction is rolled back, we set the value outside of the transaction
			return c.setVariableValueWithTx(ctx, nil, value)
		}
		return nil, fmt.Errorf("failed to get current variable value: %w", err)
	}

	switch operation {
	case flow.VariableOperationAppend:
		value.Data = currentValue.Data.Append(value.Data)
	case flow.VariableOperationPrepend:
		value.Data = value.Data.Append(currentValue.Data)
	case flow.VariableOperationIncrement:
		value.Data = currentValue.Data.Add(value.Data)
	case flow.VariableOperationDecrement:
		value.Data = currentValue.Data.Sub(value.Data)
	}

	newValue, err := c.setVariableValueWithTx(ctx, tx, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set variable value: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newValue, nil
}

func (c *Client) DeleteVariableValue(ctx context.Context, variableID string, scope null.String) error {
	err := c.Q.DeleteVariableValue(ctx, pgmodel.DeleteVariableValueParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: scope.String, Valid: scope.Valid},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) DeleteAllVariableValues(ctx context.Context, variableID string) error {
	err := c.Q.DeleteAllVariableValues(ctx, variableID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) variableValueWithTx(ctx context.Context, tx pgx.Tx, variableID string, scope null.String) (*model.VariableValue, error) {
	q := c.Q
	if tx != nil {
		q = c.Q.WithTx(tx)
	}

	row, err := q.GetVariableValueForUpdate(ctx, pgmodel.GetVariableValueForUpdateParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: scope.String, Valid: scope.Valid},
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	v, err := rowToVariableValue(row)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func (c *Client) setVariableValueWithTx(ctx context.Context, tx pgx.Tx, value model.VariableValue) (*model.VariableValue, error) {
	q := c.Q
	if tx != nil {
		q = c.Q.WithTx(tx)
	}

	data, err := json.Marshal(value.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variable value: %w", err)
	}

	row, err := q.SetVariableValue(ctx, pgmodel.SetVariableValueParams{
		VariableID: value.VariableID,
		Scope:      pgtype.Text{String: value.Scope.String, Valid: value.Scope.Valid},
		Value:      data,
		CreatedAt:  pgtype.Timestamp{Time: value.CreatedAt, Valid: true},
		UpdatedAt:  pgtype.Timestamp{Time: value.UpdatedAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	v, err := rowToVariableValue(row)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func rowToVariableValue(row pgmodel.VariableValue) (model.VariableValue, error) {
	var data thing.Any
	err := json.Unmarshal(row.Value, &data)
	if err != nil {
		return model.VariableValue{}, fmt.Errorf("failed to unmarshal variable value: %w", err)
	}

	return model.VariableValue{
		ID:         uint64(row.ID),
		VariableID: row.VariableID,
		Scope:      null.NewString(row.Scope.String, row.Scope.Valid),
		Data:       data,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}, nil
}
