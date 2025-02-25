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
	"github.com/kitecloud/kite/kite-service/pkg/message"
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

	var data message.MessageData
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

func (c *Client) MessageInstance(ctx context.Context, messageID string, instanceID uint64) (*model.MessageInstance, error) {
	row, err := c.Q.GetMessageInstance(ctx, pgmodel.GetMessageInstanceParams{
		MessageID: messageID,
		ID:        int64(instanceID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMessageInstance(row)
}

func (c *Client) MessageInstancesByMessage(ctx context.Context, messageID string, includeHidden bool) ([]*model.MessageInstance, error) {
	var rows []pgmodel.MessageInstance
	var err error

	if includeHidden {
		rows, err = c.Q.GetMessageInstancesByMessageWithHidden(ctx, messageID)
	} else {
		rows, err = c.Q.GetMessageInstancesByMessage(ctx, messageID)
	}
	if err != nil {
		return nil, err
	}

	instances := make([]*model.MessageInstance, len(rows))
	for i, row := range rows {
		msg, err := rowToMessageInstance(row)
		if err != nil {
			return nil, err
		}
		instances[i] = msg
	}

	return instances, nil
}

func (c *Client) MessageInstanceByDiscordMessageID(ctx context.Context, discordMessageID string) (*model.MessageInstance, error) {
	row, err := c.Q.GetMessageInstanceByDiscordMessageId(ctx, discordMessageID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMessageInstance(row)
}

func (c *Client) CreateMessageInstance(ctx context.Context, instance *model.MessageInstance) (*model.MessageInstance, error) {
	flowSources, err := json.Marshal(instance.FlowSources)
	if err != nil {
		return nil, err
	}

	res, err := c.Q.CreateMessageInstance(ctx, pgmodel.CreateMessageInstanceParams{
		MessageID:        instance.MessageID,
		DiscordGuildID:   instance.DiscordGuildID,
		DiscordChannelID: instance.DiscordChannelID,
		DiscordMessageID: instance.DiscordMessageID,
		Ephemeral:        instance.Ephemeral,
		Hidden:           instance.Hidden,
		FlowSources:      flowSources,
		CreatedAt:        pgtype.Timestamp{Time: instance.CreatedAt.UTC(), Valid: true},
		UpdatedAt:        pgtype.Timestamp{Time: instance.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToMessageInstance(res)
}

func (c *Client) UpdateMessageInstance(ctx context.Context, instance *model.MessageInstance) (*model.MessageInstance, error) {
	flowSources, err := json.Marshal(instance.FlowSources)
	if err != nil {
		return nil, err
	}

	res, err := c.Q.UpdateMessageInstance(ctx, pgmodel.UpdateMessageInstanceParams{
		ID:          int64(instance.ID),
		MessageID:   instance.MessageID,
		FlowSources: flowSources,
		UpdatedAt:   pgtype.Timestamp{Time: instance.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToMessageInstance(res)
}

func (c *Client) DeleteMessageInstance(ctx context.Context, messageID string, instanceID uint64) error {
	err := c.Q.DeleteMessageInstance(ctx, pgmodel.DeleteMessageInstanceParams{
		MessageID: messageID,
		ID:        int64(instanceID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func (c *Client) DeleteMessageInstanceByDiscordMessageID(ctx context.Context, discordMessageID string) error {
	err := c.Q.DeleteMessageInstanceByDiscordMessageId(ctx, discordMessageID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return store.ErrNotFound
		}
		return err
	}

	return nil
}

func rowToMessageInstance(row pgmodel.MessageInstance) (*model.MessageInstance, error) {
	var flowSources map[string]flow.FlowData
	if err := json.Unmarshal(row.FlowSources, &flowSources); err != nil {
		return nil, err
	}

	return &model.MessageInstance{
		ID:               uint64(row.ID),
		MessageID:        row.MessageID,
		DiscordGuildID:   row.DiscordGuildID,
		DiscordChannelID: row.DiscordChannelID,
		DiscordMessageID: row.DiscordMessageID,
		Ephemeral:        row.Ephemeral,
		Hidden:           row.Hidden,
		FlowSources:      flowSources,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}, nil
}
