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
		Scope: variable.Scope,
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
		Scope: variable.Scope,
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
		Scope:     row.Scope,
		AppID:     row.AppID,
		ModuleID:  null.NewString(row.ModuleID.String, row.ModuleID.Valid),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
