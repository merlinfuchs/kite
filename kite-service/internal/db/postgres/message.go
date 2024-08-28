package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) MessagesByApp(ctx context.Context, appID string) ([]*model.Message, error) {
	rows, err := c.Q.GetMessagesByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	messages := make([]*model.Message, len(rows))
	for i, row := range rows {
		msg, err := rowToMessage(row)
		if err != nil {
			return nil, err
		}
		messages[i] = msg
	}

	return messages, nil
}

func (c *Client) CountMessagesByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountMessagesByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) Message(ctx context.Context, id string) (*model.Message, error) {
	row, err := c.Q.GetMessage(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMessage(row)
}

func (c *Client) CreateMessage(ctx context.Context, variable *model.Message) (*model.Message, error) {
	flowSources, err := json.Marshal(variable.FlowSources)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(variable.Data)
	if err != nil {
		return nil, err
	}

	res, err := c.Q.CreateMessage(ctx, pgmodel.CreateMessageParams{
		ID:   variable.ID,
		Name: variable.Name,
		Description: pgtype.Text{
			String: variable.Description.String,
			Valid:  variable.Description.Valid,
		},
		AppID: variable.AppID,
		ModuleID: pgtype.Text{
			String: variable.ModuleID.String,
			Valid:  variable.ModuleID.Valid,
		},
		CreatorUserID: variable.CreatorUserID,
		FlowSources:   flowSources,
		Data:          data,
		CreatedAt:     pgtype.Timestamp{Time: variable.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: variable.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToMessage(res)
}

func (c *Client) UpdateMessage(ctx context.Context, variable *model.Message) (*model.Message, error) {
	flowSources, err := json.Marshal(variable.FlowSources)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(variable.Data)
	if err != nil {
		return nil, err
	}

	res, err := c.Q.UpdateMessage(ctx, pgmodel.UpdateMessageParams{
		ID:   variable.ID,
		Name: variable.Name,
		Description: pgtype.Text{
			String: variable.Description.String,
			Valid:  variable.Description.Valid,
		},
		FlowSources: flowSources,
		Data:        data,
		UpdatedAt:   pgtype.Timestamp{Time: variable.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMessage(res)
}

func (c *Client) DeleteMessage(ctx context.Context, id string) error {
	err := c.Q.DeleteMessage(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func rowToMessage(row pgmodel.Message) (*model.Message, error) {
	var flowSources map[string]flow.FlowData
	if err := json.Unmarshal(row.FlowSources, &flowSources); err != nil {
		return nil, err
	}

	var data model.MessageData
	if err := json.Unmarshal(row.Data, &data); err != nil {
		return nil, err
	}

	return &model.Message{
		ID:            row.ID,
		Name:          row.Name,
		Description:   null.NewString(row.Description.String, row.Description.Valid),
		AppID:         row.AppID,
		ModuleID:      null.NewString(row.ModuleID.String, row.ModuleID.Valid),
		CreatorUserID: row.CreatorUserID,
		FlowSources:   flowSources,
		Data:          data,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}
