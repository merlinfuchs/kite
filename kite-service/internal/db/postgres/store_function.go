package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) FunctionsByApp(ctx context.Context, appID string) ([]*model.Function, error) {
	rows, err := c.Q.GetFunctionsByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	res := make([]*model.Function, len(rows))
	for i, row := range rows {
		function, err := rowToFunction(row)
		if err != nil {
			return nil, err
		}
		res[i] = function
	}

	return res, nil
}

func (c *Client) CountFunctionsByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountFunctionsByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) Function(ctx context.Context, appID string, id string) (*model.Function, error) {
	row, err := c.Q.GetFunction(ctx, pgmodel.GetFunctionParams{
		ID:    id,
		AppID: appID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToFunction(row)
}

func (c *Client) CreateFunction(ctx context.Context, function *model.Function) (*model.Function, error) {
	flowSource, err := json.Marshal(function.FlowSource)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.CreateFunction(ctx, pgmodel.CreateFunctionParams{
		ID:          function.ID,
		Name:        function.Name,
		Description: function.Description,
		AppID:       function.AppID,
		ModuleID: pgtype.Text{
			String: function.ModuleID.String,
			Valid:  function.ModuleID.Valid,
		},
		CreatorUserID: function.CreatorUserID,
		FlowSource:    flowSource,
		CreatedAt:     pgtype.Timestamp{Time: function.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: function.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToFunction(row)
}

func (c *Client) UpdateFunction(ctx context.Context, function *model.Function) (*model.Function, error) {
	flowSource, err := json.Marshal(function.FlowSource)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.UpdateFunction(ctx, pgmodel.UpdateFunctionParams{
		ID:          function.ID,
		AppID:       function.AppID,
		Name:        function.Name,
		Description: function.Description,
		FlowSource:  flowSource,
		UpdatedAt:   pgtype.Timestamp{Time: function.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToFunction(row)
}

func (c *Client) DeleteFunction(ctx context.Context, appID string, id string) error {
	return c.Q.DeleteFunction(ctx, pgmodel.DeleteFunctionParams{
		ID:    id,
		AppID: appID,
	})
}

func rowToFunction(row pgmodel.Function) (*model.Function, error) {
	var flowSource flow.FlowData
	if err := json.Unmarshal(row.FlowSource, &flowSource); err != nil {
		return nil, err
	}

	return &model.Function{
		ID:            row.ID,
		Name:          row.Name,
		Description:   row.Description,
		AppID:         row.AppID,
		ModuleID:      null.NewString(row.ModuleID.String, row.ModuleID.Valid),
		CreatorUserID: row.CreatorUserID,
		FlowSource:    flowSource,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}
