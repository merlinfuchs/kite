package bundle

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type memoryBundleStore struct {
	store.BundleStore
	bundles map[string]*model.Bundle
}

func (s *memoryBundleStore) Bundle(ctx context.Context, id string) (*model.Bundle, error) {
	b, ok := s.bundles[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	copied := *b
	return &copied, nil
}

func get(t *testing.T, h *BundleHandler, bundleID, userID string) (int, map[string]any) {
	t.Helper()
	srv := handler.APIHandler(func(c *handler.Context) error {
		if userID != "" {
			c.Session = &model.Session{UserID: userID}
		}
		return handler.Typed(h.HandleBundleGet)(c)
	})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetPathValue("bundleID", bundleID)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func TestBundleGet(t *testing.T) {
	h := NewBundleHandler(&memoryBundleStore{bundles: map[string]*model.Bundle{
		"moderation": {ID: "moderation", Listing: &model.BundleListing{}},
		"kit":        {ID: "kit", OwnerUserID: null.StringFrom("owner")},
	}})

	code, res := get(t, h, "moderation", "")
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, true, res["data"].(map[string]any)["official"])

	// A user's bundle is only visible to them.
	code, _ = get(t, h, "kit", "")
	assert.Equal(t, http.StatusNotFound, code)
	code, _ = get(t, h, "kit", "other")
	assert.Equal(t, http.StatusNotFound, code)
	code, res = get(t, h, "kit", "owner")
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, true, res["data"].(map[string]any)["yours"])
	assert.Equal(t, false, res["data"].(map[string]any)["official"])
}
