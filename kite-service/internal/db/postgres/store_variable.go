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
	"github.com/kitecloud/kite/kite-service/pkg/provider"
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

func (c *Client) VariablesByAppWithoutTotals(ctx context.Context, appID string) ([]*model.Variable, error) {
	rows, err := c.Q.GetVariablesByAppWithoutTotals(ctx, appID)
	if err != nil {
		return nil, err
	}

	variables := make([]*model.Variable, len(rows))
	for i, row := range rows {
		variables[i] = rowToVariable(row)
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

// CreateVariable returns store.ErrAlreadyExists if the app has a variable
// with the name.
func (c *Client) CreateVariable(ctx context.Context, variable *model.Variable) (*model.Variable, error) {
	v, err := c.createVariableWithTx(ctx, nil, variable)
	return v, uniqueViolation(err)
}

func (c *Client) createVariableWithTx(ctx context.Context, tx pgx.Tx, variable *model.Variable) (*model.Variable, error) {
	q := c.Q
	if tx != nil {
		q = c.Q.WithTx(tx)
	}

	row, err := q.CreateVariable(ctx, pgmodel.CreateVariableParams{
		ID:     variable.ID,
		Name:   variable.Name,
		Scoped: variable.Scoped,
		AppID:  variable.AppID,
		BundleInstallID: pgtype.Text{
			String: variable.BundleInstallID.String,
			Valid:  variable.BundleInstallID.Valid,
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
		return nil, uniqueViolation(err)
	}

	return rowToVariable(row), nil
}

func (c *Client) DeleteVariable(ctx context.Context, id string) error {
	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	q := c.Q.WithTx(tx)
	installID, err := q.DeleteVariable(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return err
	}
	if err := deleteEmptyBundleInstall(ctx, q, installID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func rowToVariable(row pgmodel.Variable) *model.Variable {
	return &model.Variable{
		ID:              row.ID,
		Name:            row.Name,
		Scoped:          row.Scoped,
		AppID:           row.AppID,
		BundleInstallID: null.NewString(row.BundleInstallID.String, row.BundleInstallID.Valid),
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
	}
}

func (c *Client) VariableValues(ctx context.Context, appID string, variableID string) ([]*model.VariableValue, error) {
	rows, err := c.Q.GetVariableValues(ctx, pgmodel.GetVariableValuesParams{
		VariableID: variableID,
		AppID:      appID,
	})
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

func (c *Client) SearchVariableValues(ctx context.Context, appID string, variableID string, search string, limit int, offset int) ([]*model.VariableValue, int, error) {
	total, err := c.Q.CountSearchVariableValues(ctx, pgmodel.CountSearchVariableValuesParams{
		VariableID: variableID,
		AppID:      appID,
		Search:     search,
	})
	if err != nil {
		return nil, 0, err
	}

	rows, err := c.Q.SearchVariableValues(ctx, pgmodel.SearchVariableValuesParams{
		VariableID: variableID,
		AppID:      appID,
		Search:     search,
		RowLimit:   int32(limit),
		RowOffset:  int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	values := make([]*model.VariableValue, 0, len(rows))
	for _, row := range rows {
		v, err := rowToVariableValue(row)
		if err != nil {
			return nil, 0, err
		}
		values = append(values, &v)
	}

	return values, int(total), nil
}

func (c *Client) ListVariableValues(ctx context.Context, appID string, variableID string, scopePrefix string, sort string, limit int, offset int) ([]*model.VariableValue, error) {
	rows, err := c.Q.ListVariableValues(ctx, pgmodel.ListVariableValuesParams{
		VariableID:  variableID,
		AppID:       appID,
		ScopePrefix: scopePrefix,
		Sort:        sort,
		RowLimit:    int32(limit),
		RowOffset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	values := make([]*model.VariableValue, 0, len(rows))
	for _, row := range rows {
		v, err := rowToVariableValue(row)
		if err != nil {
			return nil, err
		}
		values = append(values, &v)
	}

	return values, nil
}

func (c *Client) VariableValue(ctx context.Context, appID string, variableID string, scope null.String) (*model.VariableValue, error) {
	row, err := c.Q.GetVariableValue(ctx, pgmodel.GetVariableValueParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: scope.String, Valid: scope.Valid},
		AppID:      appID,
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

func (c *Client) SetVariableValue(ctx context.Context, appID string, value model.VariableValue) error {
	_, err := c.writeVariableValueWithTx(ctx, nil, appID, value)
	return err
}

func (c *Client) UpdateVariableValue(ctx context.Context, appID string, operation model.VariableValueOperation, value model.VariableValue) (*model.VariableValue, error) {
	if operation.IsOverwrite() {
		return c.writeVariableValueWithTx(ctx, nil, appID, value)
	}

	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	currentValue, err := c.variableValueWithTx(ctx, tx, appID, value.VariableID, value.Scope)
	if errors.Is(err, store.ErrNotFound) {
		// There's no row to lock yet, so the first value is inserted. If
		// another run inserted one in the meantime, the insert waits for it
		// and locks it instead, and this change is made to it. Otherwise each
		// would store its first value and only the last change would count.
		first := value
		first.Data = operation.Apply(thing.Null, false, value.Data)
		data, err := encodeVariableValue(first.Data)
		if err != nil {
			return nil, err
		}
		row, err := c.Q.WithTx(tx).InsertOrLockVariableValue(ctx, pgmodel.InsertOrLockVariableValueParams(variableValueParams(appID, first, data)))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		} else if err != nil {
			return nil, fmt.Errorf("failed to insert variable value: %w", err)
		}
		v, err := rowToVariableValue(row.VariableValue)
		if err != nil {
			return nil, err
		}
		if row.Inserted {
			if err := tx.Commit(ctx); err != nil {
				return nil, fmt.Errorf("failed to commit transaction: %w", err)
			}
			return &v, nil
		}
		currentValue = &v
	} else if err != nil {
		return nil, fmt.Errorf("failed to get current variable value: %w", err)
	}

	value.Data = operation.Apply(currentValue.Data, true, value.Data)
	newValue, err := c.writeVariableValueWithTx(ctx, tx, appID, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set variable value: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newValue, nil
}

func (c *Client) DeleteVariableValue(ctx context.Context, appID string, variableID string, scope null.String) error {
	err := c.Q.DeleteVariableValue(ctx, pgmodel.DeleteVariableValueParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: scope.String, Valid: scope.Valid},
		AppID:      appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) DeleteAllVariableValues(ctx context.Context, appID string, variableID string) error {
	err := c.Q.DeleteAllVariableValues(ctx, pgmodel.DeleteAllVariableValuesParams{
		VariableID: variableID,
		AppID:      appID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) variableValueWithTx(ctx context.Context, tx pgx.Tx, appID string, variableID string, scope null.String) (*model.VariableValue, error) {
	q := c.Q
	if tx != nil {
		q = c.Q.WithTx(tx)
	}

	row, err := q.GetVariableValueForUpdate(ctx, pgmodel.GetVariableValueForUpdateParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: scope.String, Valid: scope.Valid},
		AppID:      appID,
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

// writeVariableValueWithTx stores the value. It returns store.ErrNotFound
// if the variable isn't the app's.
func (c *Client) writeVariableValueWithTx(ctx context.Context, tx pgx.Tx, appID string, value model.VariableValue) (*model.VariableValue, error) {
	q := c.Q
	if tx != nil {
		q = c.Q.WithTx(tx)
	}

	data, err := encodeVariableValue(value.Data)
	if err != nil {
		return nil, err
	}
	row, err := q.SetVariableValue(ctx, variableValueParams(appID, value, data))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	v, err := rowToVariableValue(row)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func variableValueParams(appID string, value model.VariableValue, data []byte) pgmodel.SetVariableValueParams {
	return pgmodel.SetVariableValueParams{
		VariableID: value.VariableID,
		AppID:      appID,
		Scope:      pgtype.Text{String: value.Scope.String, Valid: value.Scope.Valid},
		Value:      data,
		CreatedAt:  pgtype.Timestamp{Time: value.CreatedAt, Valid: true},
		UpdatedAt:  pgtype.Timestamp{Time: value.UpdatedAt, Valid: true},
	}
}

// encodeVariableValue is the JSON a value is stored as, which can be up to
// provider.MaxVariableValueSize.
func encodeVariableValue(v thing.Thing) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variable value: %w", err)
	}
	if len(data) > provider.MaxVariableValueSize {
		return nil, provider.ErrVariableValueTooLarge
	}
	return data, nil
}

func rowToVariableValue(row pgmodel.VariableValue) (model.VariableValue, error) {
	var data thing.Thing
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
