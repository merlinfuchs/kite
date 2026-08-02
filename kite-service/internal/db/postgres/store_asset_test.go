package postgres

import (
	"context"
	"errors"
	"testing"

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
