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

// The user someone has the Discord ID 456, and bot-owner owns the bot 999.
func (s *memoryUserStore) UserByDiscordID(ctx context.Context, discordID string) (*model.User, error) {
	if discordID != "456" {
		return nil, store.ErrNotFound
	}
	return s.User(ctx, "someone")
}

func (s *memoryUserStore) UserByAppDiscordID(ctx context.Context, discordID string) (*model.User, error) {
	if discordID != "999" {
		return nil, store.ErrNotFound
	}
	return s.User(ctx, "bot-owner")
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

func TestUserBanKite(t *testing.T) {
	users := &memoryUserStore{bans: map[string]*model.UserBan{}}
	h := NewUserBanHandler(users, nil)
	create := handler.TypedWithBody(h.HandleUserBanCreate)

	// The user is set by exactly one ID, and admins can't ban themselves.
	code, res := call(t, create, nil, wire.UserBanCreateRequest{Feature: "kite", Reason: "Nuke bot"})
	assert.Equal(t, http.StatusBadRequest, code, res)
	code, res = call(t, create, nil, wire.UserBanCreateRequest{UserID: "owner", DiscordID: "999", Feature: "kite", Reason: "Nuke bot"})
	assert.Equal(t, http.StatusBadRequest, code, res)
	code, res = call(t, create, nil, wire.UserBanCreateRequest{UserID: "admin", Feature: "kite", Reason: "Nuke bot"})
	assert.Equal(t, http.StatusBadRequest, code, res)
	code, res = call(t, create, nil, wire.UserBanCreateRequest{DiscordID: "123", Feature: "kite", Reason: "Nuke bot"})
	assert.Equal(t, http.StatusNotFound, code, res)

	code, res = call(t, create, nil, wire.UserBanCreateRequest{UserID: "owner", Feature: "kite", Reason: "Nuke bot"})
	require.Equal(t, http.StatusOK, code, res)
	assert.NotNil(t, users.bans["owner/kite"])

	// A Discord ID is the user's own or that of one of their bots.
	code, res = call(t, create, nil, wire.UserBanCreateRequest{DiscordID: "456", Feature: "kite", Reason: "Nuke bot"})
	require.Equal(t, http.StatusOK, code, res)
	assert.NotNil(t, users.bans["someone/kite"])
	code, res = call(t, create, nil, wire.UserBanCreateRequest{DiscordID: "999", Feature: "kite", Reason: "Nuke bot"})
	require.Equal(t, http.StatusOK, code, res)
	assert.NotNil(t, users.bans["bot-owner/kite"])
}
