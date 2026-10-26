package bundle

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type memoryBundleStore struct {
	store.BundleStore
	bundles map[string]*model.Bundle
	reports []*model.BundleReport
}

func (s *memoryBundleStore) SubmitBundleReport(ctx context.Context, report *model.BundleReport) error {
	s.reports = append(s.reports, report)
	return nil
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

	// Anyone with its link sees a user's bundle, with the owner as they're on
	// Discord, but not their email.
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
		ID:            "kit",
		OwnerUserID:   null.StringFrom("owner"),
		Listing:       &model.BundleListing{},
		ListedVersion: 2,
		Owner:         &model.User{ID: "owner", DisplayName: "Jack", Email: "secret@example.com", DiscordUsername: "jack"},
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

func TestBundleGetRemoved(t *testing.T) {
	bundles := &memoryBundleStore{bundles: map[string]*model.Bundle{
		"kit": {
			ID:            "kit",
			OwnerUserID:   null.StringFrom("owner"),
			RemovedAt:     null.TimeFrom(time.Now()),
			RemovedReason: "Steals tokens.",
		},
	}}
	h := NewBundleHandler(bundles, &memoryUserStore{})

	// Apps that added it learn why it's gone, its owner still sees it.
	code, res := get(t, h, "kit", "someone")
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, "bundle_removed", res["error"].(map[string]any)["code"])
	code, res = get(t, h, "kit", "owner")
	assert.Equal(t, http.StatusOK, code, res)
}

func TestBundleReportCreate(t *testing.T) {
	bundles := &memoryBundleStore{bundles: map[string]*model.Bundle{
		"shared":     {ID: "shared", OwnerUserID: null.StringFrom("owner")},
		"removed":    {ID: "removed", OwnerUserID: null.StringFrom("owner"), RemovedAt: null.TimeFrom(time.Now())},
		"moderation": {ID: "moderation"},
	}}
	h := NewBundleHandler(bundles, &memoryUserStore{})
	report := func(userID, bundleID string, req wire.BundleReportCreateRequest) int {
		srv := handler.APIHandler(func(c *handler.Context) error {
			c.Session = &model.Session{UserID: userID}
			return handler.TypedWithBody(h.HandleBundleReportCreate)(c)
		})
		body, err := json.Marshal(req)
		require.NoError(t, err)
		r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.SetPathValue("bundleID", bundleID)
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, r)
		return rec.Code
	}
	req := wire.BundleReportCreateRequest{Category: "scam", Details: "Asks for tokens."}

	assert.Equal(t, http.StatusNotFound, report("someone", "removed", req))
	// Only bundles of other users can be reported.
	assert.Equal(t, http.StatusBadRequest, report("owner", "shared", req))
	assert.Equal(t, http.StatusBadRequest, report("someone", "moderation", req))
	assert.Equal(t, http.StatusBadRequest, report("someone", "shared", wire.BundleReportCreateRequest{Category: "nope", Details: "x"}))

	assert.Equal(t, http.StatusOK, report("someone", "shared", req))
	require.Len(t, bundles.reports, 1)
	assert.Equal(t, "someone", bundles.reports[0].ReporterUserID)
	assert.Equal(t, model.BundleReportCategoryScam, bundles.reports[0].Category)
}
