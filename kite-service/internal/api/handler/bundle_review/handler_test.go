package bundlereview

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	pending map[string]*model.BundleReview
	decided []*model.BundleReview
	bundles map[string]*model.Bundle
}

func (s *memoryBundleStore) RemoveBundle(ctx context.Context, id string, reason string, resolverUserID string, at time.Time) error {
	b, ok := s.bundles[id]
	if !ok || !b.OwnerUserID.Valid || b.RemovedAt.Valid {
		return store.ErrNotFound
	}
	b.RemovedAt = null.TimeFrom(at)
	b.RemovedReason = reason
	return nil
}

func pendingReview(r *model.BundleReview, content json.RawMessage) *model.PendingBundleReview {
	return &model.PendingBundleReview{
		Review: r,
		Bundle: &model.Bundle{
			ID:          r.BundleID,
			OwnerUserID: null.StringFrom("owner"),
			Content:     content,
			Owner:       &model.User{ID: "owner", DisplayName: "Merlin", Email: "secret@example.com", DiscordUsername: "merlin"},
		},
	}
}

func (s *memoryBundleStore) PendingBundleReviews(ctx context.Context) ([]*model.PendingBundleReview, error) {
	var res []*model.PendingBundleReview
	for _, r := range s.pending {
		res = append(res, pendingReview(r, nil))
	}
	return res, nil
}

func (s *memoryBundleStore) PendingBundleReview(ctx context.Context, id string) (*model.PendingBundleReview, error) {
	r, ok := s.pending[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	return pendingReview(r, json.RawMessage(`{"commands":[]}`)), nil
}

func (s *memoryBundleStore) DecideBundleReview(ctx context.Context, decision *model.BundleReview) (*model.BundleReview, error) {
	pending, ok := s.pending[decision.ID]
	if !ok {
		return nil, store.ErrNotFound
	}
	delete(s.pending, decision.ID)
	decided := *decision
	decided.BundleID = pending.BundleID
	decided.Version = pending.Version
	s.decided = append(s.decided, &decided)
	return &decided, nil
}

type memoryUserStore struct{ store.UserStore }

func (s *memoryUserStore) User(ctx context.Context, id string) (*model.User, error) {
	return &model.User{ID: id, DisplayName: "Merlin", Email: "secret@example.com", DiscordID: "123", DiscordUsername: "merlin"}, nil
}

func call(t *testing.T, h handler.HandlerFunc, reviewID string, req any) (int, map[string]any) {
	t.Helper()
	return callAs(t, h, "admin", reviewID, req)
}

// callAs calls the handler as the user, with the ID as the review, report
// and bundle of the route.
func callAs(t *testing.T, h handler.HandlerFunc, userID string, id string, req any) (int, map[string]any) {
	t.Helper()
	srv := handler.APIHandler(func(c *handler.Context) error {
		c.Session = &model.Session{UserID: userID}
		return h(c)
	})

	body, err := json.Marshal(req)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("reviewID", id)
	r.SetPathValue("reportID", id)
	r.SetPathValue("bundleID", id)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func TestBundleReviewDecide(t *testing.T) {
	bundles := &memoryBundleStore{pending: map[string]*model.BundleReview{
		"r1": {ID: "r1", BundleID: "b1", Version: 2, Status: model.BundleReviewStatusPending},
		"r2": {ID: "r2", BundleID: "b2", Version: 1, Status: model.BundleReviewStatusPending},
	}}
	h := NewBundleReviewHandler(bundles, &memoryUserStore{})
	decide := handler.TypedWithBody(h.HandleBundleReviewDecide)

	code, res := call(t, handler.Typed(h.HandlePendingBundleReviewList), "", nil)
	require.Equal(t, http.StatusOK, code, res)
	list := res["data"].([]any)
	require.Len(t, list, 2)
	// Admins see who made it, but not their email.
	owner := list[0].(map[string]any)["bundle"].(map[string]any)["owner"].(map[string]any)
	assert.Equal(t, "merlin", owner["discord_username"])
	assert.NotContains(t, owner, "email")
	// The list has no content, the review of one has.
	assert.Nil(t, list[0].(map[string]any)["bundle"].(map[string]any)["content"])
	code, res = call(t, handler.Typed(h.HandlePendingBundleReviewGet), "r1", nil)
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, map[string]any{"commands": []any{}}, res["data"].(map[string]any)["bundle"].(map[string]any)["content"])

	// Rejecting needs a reason for the owner.
	code, res = call(t, decide, "r1", wire.BundleReviewDecideRequest{})
	assert.Equal(t, http.StatusBadRequest, code, res)

	code, res = call(t, decide, "r1", wire.BundleReviewDecideRequest{Approve: true, Reason: "ignored"})
	require.Equal(t, http.StatusOK, code, res)
	code, res = call(t, decide, "r2", wire.BundleReviewDecideRequest{Reason: "Bans everyone who joins."})
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, "rejected", res["data"].(map[string]any)["status"])

	require.Len(t, bundles.decided, 2)
	assert.Equal(t, model.BundleReviewStatusApproved, bundles.decided[0].Status)
	assert.Empty(t, bundles.decided[0].Reason)
	assert.Equal(t, "admin", bundles.decided[0].ReviewerUserID.String)
	assert.Equal(t, "Bans everyone who joins.", bundles.decided[1].Reason)

	// A review is decided once.
	code, res = call(t, decide, "r1", wire.BundleReviewDecideRequest{Approve: true})
	assert.Equal(t, http.StatusNotFound, code, res)
	code, res = call(t, handler.Typed(h.HandlePendingBundleReviewGet), "r1", nil)
	assert.Equal(t, http.StatusNotFound, code, res)
}

func TestRequireUserIn(t *testing.T) {
	srv := func(userID string) int {
		h := handler.APIHandler(func(c *handler.Context) error {
			c.Session = &model.Session{UserID: userID}
			return handler.RequireUserIn([]string{"admin"})(func(c *handler.Context) error {
				return c.JSON(http.StatusOK, map[string]any{})
			})(c)
		})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		return rec.Code
	}
	assert.Equal(t, http.StatusOK, srv("admin"))
	assert.Equal(t, http.StatusForbidden, srv("someone"))
}

func TestBundleRemove(t *testing.T) {
	bundles := &memoryBundleStore{bundles: map[string]*model.Bundle{
		"shared":     {ID: "shared", OwnerUserID: null.StringFrom("owner")},
		"moderation": {ID: "moderation"},
	}}
	h := NewBundleReviewHandler(bundles, &memoryUserStore{})
	remove := handler.TypedWithBody(h.HandleBundleRemove)

	// Taking it down needs a reason for the owner, and happens once.
	code, res := call(t, remove, "shared", wire.BundleRemoveRequest{})
	assert.Equal(t, http.StatusBadRequest, code, res)
	code, res = call(t, remove, "shared", wire.BundleRemoveRequest{Reason: "Steals tokens."})
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, "Steals tokens.", bundles.bundles["shared"].RemovedReason)
	code, res = call(t, remove, "shared", wire.BundleRemoveRequest{Reason: "Again."})
	assert.Equal(t, http.StatusNotFound, code, res)
	code, res = call(t, remove, "moderation", wire.BundleRemoveRequest{Reason: "Official."})
	assert.Equal(t, http.StatusNotFound, code, res)

	// Only its owner can see it anymore.
	assert.False(t, bundles.bundles["shared"].VisibleTo("someone"))
	assert.True(t, bundles.bundles["shared"].VisibleTo("owner"))
}
