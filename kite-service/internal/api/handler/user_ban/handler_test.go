package userban

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryUserStore struct {
	store.UserStore
	bans map[string]*model.UserBan
}

func (s *memoryUserStore) User(ctx context.Context, id string) (*model.User, error) {
	if id == "unknown" {
		return nil, store.ErrNotFound
	}
	return &model.User{ID: id, DisplayName: "Merlin", DiscordID: "123", DiscordUsername: "merlin"}, nil
}

func (s *memoryUserStore) BanUser(ctx context.Context, ban *model.UserBan) error {
	s.bans[ban.UserID+"/"+string(ban.Feature)] = ban
	return nil
}

func (s *memoryUserStore) UnbanUser(ctx context.Context, userID string, feature model.UserBanFeature) error {
	key := userID + "/" + string(feature)
	if _, ok := s.bans[key]; !ok {
		return store.ErrNotFound
	}
	delete(s.bans, key)
	return nil
}

func call(t *testing.T, h handler.HandlerFunc, path map[string]string, req any) (int, map[string]any) {
	t.Helper()
	srv := handler.APIHandler(func(c *handler.Context) error {
		c.Session = &model.Session{UserID: "admin"}
		return h(c)
	})

	body, err := json.Marshal(req)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")
	for k, v := range path {
		r.SetPathValue(k, v)
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func TestUserBan(t *testing.T) {
	users := &memoryUserStore{bans: map[string]*model.UserBan{}}
	h := NewUserBanHandler(users, nil)
	create := handler.TypedWithBody(h.HandleUserBanCreate)
	remove := handler.Typed(h.HandleUserBanDelete)

	// Bans need a known feature and a reason for the user.
	code, res := call(t, create, nil, wire.UserBanCreateRequest{UserID: "owner", Feature: "everything", Reason: "Scams"})
	assert.Equal(t, http.StatusBadRequest, code, res)
	code, res = call(t, create, nil, wire.UserBanCreateRequest{UserID: "owner", Feature: "marketplace"})
	assert.Equal(t, http.StatusBadRequest, code, res)
	code, res = call(t, create, nil, wire.UserBanCreateRequest{UserID: "unknown", Feature: "marketplace", Reason: "Scams"})
	assert.Equal(t, http.StatusNotFound, code, res)

	code, res = call(t, create, nil, wire.UserBanCreateRequest{UserID: "owner", Feature: "marketplace", Reason: "Scams"})
	require.Equal(t, http.StatusOK, code, res)
	ban := users.bans["owner/marketplace"]
	require.NotNil(t, ban)
	assert.Equal(t, "Scams", ban.Reason)
	assert.Equal(t, "admin", ban.BannedByUserID.String)

	path := map[string]string{"userID": "owner", "feature": "marketplace"}
	code, res = call(t, remove, path, nil)
	require.Equal(t, http.StatusOK, code, res)
	assert.Empty(t, users.bans)
	code, res = call(t, remove, path, nil)
	assert.Equal(t, http.StatusNotFound, code, res)
}
