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

	var variables []*model.Variable
	for _, row := range rows {
		v := rowToVariable(row.Variable)
		v.TotalValues = null.NewInt(row.TotalValues, true)
		variables = append(variables, v)
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

func (c *Client) VariableScope(ctx context.Context, id string) (model.VariableScope, error) {
	row, err := c.Q.GetVariableScope(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", store.ErrNotFound
		}
		return "", err
	}

	return model.VariableScope(row), nil
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
		ID:    variable.ID,
		Name:  variable.Name,
		Type:  variable.Type,
		Scope: string(variable.Scope),
		AppID: variable.AppID,
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
		ID:    variable.ID,
		Name:  variable.Name,
		Type:  variable.Type,
		Scope: string(variable.Scope),
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
		Type:      row.Type,
		Scope:     model.VariableScope(row.Scope),
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

func (c *Client) VariableValue(ctx context.Context, variableID string, scope model.VariableValueScope) (*model.VariableValue, error) {
	scopeType, err := c.VariableScope(ctx, variableID)
	if err != nil {
		return nil, err
	}

	resolvedScope := scope.Resolve(scopeType)

	row, err := c.Q.GetVariableValue(ctx, pgmodel.GetVariableValueParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: resolvedScope.String, Valid: resolvedScope.Valid},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToVariableValue(row), nil
}

func (c *Client) SetVariableValue(ctx context.Context, value model.VariableValue, scope model.VariableValueScope) error {
	scopeType, err := c.VariableScope(ctx, value.VariableID)
	if err != nil {
		return err
	}

	resolvedScope := scope.Resolve(scopeType)

	_, err = c.Q.SetVariableValue(ctx, pgmodel.SetVariableValueParams{
		VariableID: value.VariableID,
		Scope:      pgtype.Text{String: resolvedScope.String, Valid: resolvedScope.Valid},
		Value:      value.Value,
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

func (c *Client) DeleteVariableValue(ctx context.Context, variableID string, scope model.VariableValueScope) error {
	scopeType, err := c.VariableScope(ctx, variableID)
	if err != nil {
		return err
	}

	resolvedScope := scope.Resolve(scopeType)

	err = c.Q.DeleteVariableValue(ctx, pgmodel.DeleteVariableValueParams{
		VariableID: variableID,
		Scope:      pgtype.Text{String: resolvedScope.String, Valid: resolvedScope.Valid},
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
		ID:         row.ID,
		VariableID: row.VariableID,
		Scope:      null.NewString(row.Scope.String, row.Scope.Valid),
		Value:      row.Value,
		CreatedAt:  row.CreatedAt.Time,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}
