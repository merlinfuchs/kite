package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

const assetBucketName = "kite-assets"

type AssetStore struct {
	pg          *Client
	objectStore store.ObjectStore
}

func NewAssetStore(ctx context.Context, pg *Client, objectStore store.ObjectStore) (*AssetStore, error) {
	store := &AssetStore{pg: pg, objectStore: objectStore}

	err := objectStore.CreateBucketIfNotExists(ctx, assetBucketName)
	if err != nil {
		return store, fmt.Errorf("failed to create asset bucket: %w", err)
	}

	return store, nil
}

func (s *AssetStore) CreateAsset(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	row, err := s.pg.Q.CreateAsset(ctx, pgmodel.CreateAssetParams{
		ID:          asset.ID,
		Name:        asset.Name,
		ContentHash: asset.ContentHash,
		ContentType: asset.ContentType,
		ContentSize: int32(asset.ContentSize),
		AppID:       asset.AppID,
		ModuleID: pgtype.Text{
			String: asset.ModuleID.String,
			Valid:  asset.ModuleID.Valid,
		},
		CreatorUserID: asset.CreatorUserID,
		CreatedAt: pgtype.Timestamp{
			Time:  asset.CreatedAt.UTC(),
			Valid: true,
		},
		UpdatedAt: pgtype.Timestamp{
			Time:  asset.UpdatedAt.UTC(),
			Valid: true,
		},
		ExpiresAt: pgtype.Timestamp{
			Time:  asset.ExpiresAt.Time.UTC(),
			Valid: asset.ExpiresAt.Valid,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	err = s.objectStore.UploadObject(ctx, assetBucketName, &model.Object{
		Name:        asset.ContentHash,
		Content:     asset.Content,
		ContentType: asset.ContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload asset object: %w", err)
	}

	return rowToAsset(row)
}

func (s *AssetStore) Asset(ctx context.Context, id string) (*model.Asset, error) {
	row, err := s.pg.Q.GetAsset(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToAsset(row)
}

func (s *AssetStore) AssetWithContent(ctx context.Context, id string) (*model.Asset, error) {
	asset, err := s.Asset(ctx, id)
	if err != nil {
		return nil, err
	}

	object, err := s.objectStore.DownloadObject(ctx, assetBucketName, asset.ContentHash)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("failed to download asset object: %w", err)
	}

	asset.Content = object.Content
	return asset, nil
}

func (s *AssetStore) DeleteAsset(ctx context.Context, id string) error {
	asset, err := s.pg.Q.DeleteAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	remainingCount, err := s.pg.Q.CountAssetsByContentHash(ctx, asset.ContentHash)
	if err != nil {
		return fmt.Errorf("failed to count assets by content hash: %w", err)
	}

	if remainingCount == 0 {
		err = s.objectStore.DeleteObject(ctx, assetBucketName, asset.ContentHash)
		if err != nil && !errors.Is(err, store.ErrNotFound) {
			return fmt.Errorf("failed to delete asset object: %w", err)
		}
	}

	return nil
}

func (s *AssetStore) DeleteExpiredAssets(ctx context.Context, timestamp time.Time) error {
	assets, err := s.pg.Q.GetExpiredAssets(ctx, pgtype.Timestamp{
		Time:  timestamp.UTC(),
		Valid: true,
	})
	if err != nil {
		return fmt.Errorf("failed to get expired assets: %w", err)
	}

	for _, asset := range assets {
		err := s.DeleteAsset(ctx, asset.ID)
		if err != nil {
			slog.Error(
				"failed to delete expired asset",
				slog.String("asset_id", asset.ID),
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

func rowToAsset(row pgmodel.Asset) (*model.Asset, error) {
	return &model.Asset{
		ID:            row.ID,
		Name:          row.Name,
		ContentHash:   row.ContentHash,
		ContentType:   row.ContentType,
		ContentSize:   int(row.ContentSize),
		AppID:         row.AppID,
		ModuleID:      null.NewString(row.ModuleID.String, row.ModuleID.Valid),
		CreatorUserID: row.CreatorUserID,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		ExpiresAt:     null.NewTime(row.ExpiresAt.Time, row.ExpiresAt.Valid),
	}, nil
}
