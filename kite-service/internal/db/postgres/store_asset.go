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
	// Upload before inserting the row. The other order leaves a row pointing
	// at content that was never stored whenever the upload fails, which breaks
	// every later read of that asset. Objects are content-addressed, so an
	// upload with no row is harmless and reusable.
	err := s.objectStore.UploadObject(ctx, assetBucketName, &model.Object{
		Name:        asset.ContentHash,
		Content:     asset.Content,
		ContentType: asset.ContentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload asset object: %w", err)
	}

	row, err := s.pg.Q.CreateAsset(ctx, pgmodel.CreateAssetParams{
		ID:          asset.ID,
		Name:        asset.Name,
		ContentHash: asset.ContentHash,
		ContentType: asset.ContentType,
		ContentSize: int32(asset.ContentSize),
		AppID:       asset.AppID,
		BundleInstallID: pgtype.Text{
			String: asset.BundleInstallID.String,
			Valid:  asset.BundleInstallID.Valid,
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

func (s *AssetStore) AssetStorageByApp(ctx context.Context, appID string) (int64, error) {
	return s.pg.Q.GetAssetStorageByApp(ctx, appID)
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
		return s.deleteObject(ctx, asset.ContentHash)
	}

	return nil
}

func (s *AssetStore) deleteObject(ctx context.Context, contentHash string) error {
	err := s.objectStore.DeleteObject(ctx, assetBucketName, contentHash)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return fmt.Errorf("failed to delete asset object: %w", err)
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

const maxUnusedObjectDeleteFailuresInARow = 5

// Objects are uploaded before their asset is inserted, so minAge keeps the
// objects of uploads that are still in progress.
func (s *AssetStore) DeleteUnusedObjects(ctx context.Context, minAge time.Duration) error {
	objects, err := s.objectStore.ListObjects(ctx, assetBucketName)
	if err != nil {
		if errors.Is(err, store.ErrObjectStoreDisabled) {
			return nil
		}
		return err
	}

	contentHashes, err := s.pg.Q.GetAssetContentHashes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get asset content hashes: %w", err)
	}

	failed, failedInARow := 0, 0
	for _, name := range unusedObjectNames(objects, contentHashes, minAge, time.Now()) {
		// An asset may have started using the file since the hashes were
		// loaded.
		count, err := s.pg.Q.CountAssetsByContentHash(ctx, name)
		if err == nil && count == 0 {
			err = s.deleteObject(ctx, name)
		}
		if err == nil {
			failedInARow = 0
			continue
		}

		slog.Error("Failed to delete unused asset object", slog.String("name", name), slog.String("error", err.Error()))
		failed++
		failedInARow++
		// The object store is probably throttling or down, so leave the rest
		// to the next run instead of retrying every object against it.
		if failedInARow >= maxUnusedObjectDeleteFailuresInARow {
			return fmt.Errorf("stopped after failing to delete %d unused asset objects in a row: %w", failedInARow, err)
		}
	}
	if failed > 0 {
		return fmt.Errorf("failed to delete %d unused asset objects", failed)
	}

	return nil
}

func unusedObjectNames(objects []model.ObjectInfo, usedContentHashes []string, minAge time.Duration, now time.Time) []string {
	used := make(map[string]bool, len(usedContentHashes))
	for _, contentHash := range usedContentHashes {
		used[contentHash] = true
	}

	var names []string
	for _, object := range objects {
		if !used[object.Name] && now.Sub(object.LastModified) >= minAge {
			names = append(names, object.Name)
		}
	}
	return names
}

func rowToAsset(row pgmodel.Asset) (*model.Asset, error) {
	return &model.Asset{
		ID:              row.ID,
		Name:            row.Name,
		ContentHash:     row.ContentHash,
		ContentType:     row.ContentType,
		ContentSize:     int(row.ContentSize),
		AppID:           row.AppID,
		BundleInstallID: null.NewString(row.BundleInstallID.String, row.BundleInstallID.Valid),
		CreatorUserID:   row.CreatorUserID,
		CreatedAt:       row.CreatedAt.Time,
		UpdatedAt:       row.UpdatedAt.Time,
		ExpiresAt:       null.NewTime(row.ExpiresAt.Time, row.ExpiresAt.Valid),
	}, nil
}
