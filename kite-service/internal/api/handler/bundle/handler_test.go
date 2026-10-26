package bundle

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

type memoryUserStore struct{ store.UserStore }

func (s *memoryUserStore) User(ctx context.Context, id string) (*model.User, error) {
	return &model.User{ID: id, DisplayName: "Merlin", Email: "secret@example.com", DiscordID: "123", DiscordUsername: "merlin", DiscordAvatar: null.StringFrom("abc")}, nil
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
	bundles := &memoryBundleStore{bundles: map[string]*model.Bundle{
		"moderation": {ID: "moderation", Listing: &model.BundleListing{}},
		"kit": {
			ID:           "kit",
			OwnerUserID:  null.StringFrom("owner"),
			VersionNotes: []model.BundleVersionNotes{{Version: 2, Notes: "Added /kick."}},
		},
	}}
	h := NewBundleHandler(bundles, &memoryUserStore{})

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

	// Published, anyone sees it with the owner as they're on Discord, but
	// not their email.
	bundles.bundles["kit"].PublishedAt = null.TimeFrom(time.Now())
	code, res = get(t, h, "kit", "")
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, false, res["data"].(map[string]any)["yours"])
	assert.Equal(t, map[string]any{
		"display_name":     "Merlin",
		"discord_id":       "123",
		"discord_username": "merlin",
		"discord_avatar":   "abc",
	}, res["data"].(map[string]any)["owner"])
	notes := res["data"].(map[string]any)["version_notes"].([]any)
	require.Len(t, notes, 1)
	assert.Equal(t, "Added /kick.", notes[0].(map[string]any)["notes"])
}

type listedBundleStore struct {
	store.BundleStore
}

func (s *listedBundleStore) OfficialBundles(ctx context.Context) ([]*model.Bundle, error) {
	return []*model.Bundle{{ID: "moderation", Listing: &model.BundleListing{}}}, nil
}

func (s *listedBundleStore) ListedUserBundles(ctx context.Context) ([]*model.Bundle, error) {
	return []*model.Bundle{{
		ID:          "kit",
		OwnerUserID: null.StringFrom("owner"),
		Listing:     &model.BundleListing{ListedVersion: 2},
		Owner:       &model.User{ID: "owner", DisplayName: "Jack", Email: "secret@example.com", DiscordUsername: "jack"},
	}}, nil
}

func TestBundleList(t *testing.T) {
	h := NewBundleHandler(&listedBundleStore{}, &memoryUserStore{})
	srv := handler.APIHandler(handler.Typed(h.HandleBundleList))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())

	var res struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	require.Len(t, res.Data, 2)
	// Official ones first, then users' listed ones with who made them.
	assert.Equal(t, true, res.Data[0]["official"])
	assert.Nil(t, res.Data[0]["owner"])
	assert.Equal(t, false, res.Data[1]["official"])
	owner := res.Data[1]["owner"].(map[string]any)
	assert.Equal(t, "jack", owner["discord_username"])
	assert.NotContains(t, owner, "email")
}
