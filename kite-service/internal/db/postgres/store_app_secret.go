package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (c *Client) AppSecretsByApp(ctx context.Context, appID string) ([]*model.AppSecret, error) {
	rows, err := c.Q.GetAppSecretsByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	res := make([]*model.AppSecret, len(rows))
	for i, row := range rows {
		res[i] = rowToAppSecret(row)
	}
	return res, nil
}

func (c *Client) AppSecret(ctx context.Context, appID string, id string) (*model.AppSecret, error) {
	row, err := c.Q.GetAppSecret(ctx, pgmodel.GetAppSecretParams{AppID: appID, ID: id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return rowToAppSecret(row), nil
}

func (c *Client) AppSecretsByNames(ctx context.Context, appID string, names []string) ([]*model.AppSecret, error) {
	rows, err := c.Q.GetAppSecretsByNames(ctx, pgmodel.GetAppSecretsByNamesParams{
		AppID: appID,
		Names: names,
	})
	if err != nil {
		return nil, err
	}

	res := make([]*model.AppSecret, len(rows))
	for i, row := range rows {
		res[i] = rowToAppSecret(row)
	}
	return res, nil
}

func (c *Client) CountAppSecretsByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountAppSecretsByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) CreateAppSecret(ctx context.Context, secret *model.AppSecret) (*model.AppSecret, error) {
	row, err := c.Q.CreateAppSecret(ctx, pgmodel.CreateAppSecretParams{
		ID:             secret.ID,
		AppID:          secret.AppID,
		Name:           pgtype.Text{String: secret.Name, Valid: true},
		ValueEncrypted: secret.ValueEncrypted,
		CreatedAt:      pgtype.Timestamp{Time: secret.CreatedAt.UTC(), Valid: true},
		UpdatedAt:      pgtype.Timestamp{Time: secret.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, secretWriteError(err)
	}
	return rowToAppSecret(row), nil
}

func (c *Client) UpdateAppSecret(ctx context.Context, secret *model.AppSecret) (*model.AppSecret, error) {
	row, err := c.Q.UpdateAppSecret(ctx, pgmodel.UpdateAppSecretParams{
		AppID:          secret.AppID,
		ID:             secret.ID,
		Name:           pgtype.Text{String: secret.Name, Valid: true},
		ValueEncrypted: secret.ValueEncrypted,
		UpdatedAt:      pgtype.Timestamp{Time: secret.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, secretWriteError(err)
	}
	return rowToAppSecret(row), nil
}

func (c *Client) DeleteAppSecret(ctx context.Context, appID string, id string) error {
	n, err := c.Q.DeleteAppSecret(ctx, pgmodel.DeleteAppSecretParams{AppID: appID, ID: id})
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}

// secretWriteError reports a name another secret of the app has already.
func secretWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return store.ErrAlreadyExists
	}
	return err
}

func rowToAppSecret(row pgmodel.AppSecret) *model.AppSecret {
	return &model.AppSecret{
		ID:             row.ID,
		AppID:          row.AppID,
		Name:           row.Name.String,
		ValueEncrypted: row.ValueEncrypted,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}
}
