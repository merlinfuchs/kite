package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// CreateAsset must upload the object before inserting the row, so that a
// failed upload never leaves a row pointing at content that isn't stored.
//
// The nil *Client is the assertion: if the insert ran first it would panic on
// s.pg.Q, so reaching a clean error proves the upload happens first.
func TestCreateAssetUploadsBeforeInsert(t *testing.T) {
	s := &AssetStore{pg: nil, objectStore: store.DisabledObjectStore{}}

	asset, err := s.CreateAsset(context.Background(), &model.Asset{
		ID:          "asset-id",
		AppID:       "app-id",
		ContentHash: "hash",
		ContentType: "image/png",
		Content:     []byte("content"),
	})
	if err == nil {
		t.Fatal("expected an error when the object store is disabled")
	}
	if asset != nil {
		t.Fatalf("expected no asset, got %+v", asset)
	}
	if !errors.Is(err, store.ErrObjectStoreDisabled) {
		t.Fatalf("expected ErrObjectStoreDisabled, got: %v", err)
	}
}

func TestUnusedObjectNames(t *testing.T) {
	now := time.Now()
	objects := []model.ObjectInfo{
		{Name: "used", LastModified: now.Add(-2 * time.Hour)},
		{Name: "unused", LastModified: now.Add(-2 * time.Hour)},
		{Name: "uploading", LastModified: now.Add(-time.Minute)},
	}

	names := unusedObjectNames(objects, []string{"used"}, time.Hour, now)
	if len(names) != 1 || names[0] != "unused" {
		t.Fatalf("expected only the old unused object, got %v", names)
	}
}

func TestDeleteUnusedObjectsWithoutObjectStore(t *testing.T) {
	s := &AssetStore{pg: nil, objectStore: store.DisabledObjectStore{}}

	if err := s.DeleteUnusedObjects(context.Background(), time.Hour); err != nil {
		t.Fatalf("expected no error without an object store, got: %v", err)
	}
}
