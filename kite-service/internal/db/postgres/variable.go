package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
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
		values = append(values, rowToVariableValue(row))
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

	return rowToVariableValue(row), nil
}

func (c *Client) SetVariableValue(ctx context.Context, value model.VariableValue) error {
	_, err := c.Q.SetVariableValue(ctx, pgmodel.SetVariableValueParams{
		VariableID: value.VariableID,
		Scope: pgtype.Text{
			String: value.Scope.String,
			Valid:  value.Scope.Valid,
		},
		Value: value.Value,
		CreatedAt: pgtype.Timestamp{
			Time:  value.CreatedAt.UTC(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:  value.UpdatedAt.UTC(),
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	return nil
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

func rowToVariableValue(row pgmodel.VariableValue) *model.VariableValue {
	return &model.VariableValue{
		ID:         uint64(row.ID),
		VariableID: row.VariableID,
		Scope:      null.NewString(row.Scope.String, row.Scope.Valid),
		Value:      row.Value,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}
