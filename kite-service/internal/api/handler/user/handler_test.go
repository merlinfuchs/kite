package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type memoryUserStore struct {
	store.UserStore
	bios   map[string]string
	banned bool
}

func (s *memoryUserStore) UserBan(ctx context.Context, userID string, feature model.UserBanFeature) (*model.UserBan, error) {
	if s.banned && feature == model.UserBanFeatureMarketplace {
		return &model.UserBan{UserID: userID, Feature: feature, Reason: "spam"}, nil
	}
	return nil, store.ErrNotFound
}

func (s *memoryUserStore) UpdateUserBio(ctx context.Context, userID string, bio string, updatedAt time.Time) (*model.User, error) {
	s.bios[userID] = bio
	return &model.User{ID: userID, Email: "jack@example.com", DisplayName: "Jack", DiscordID: "123", Bio: bio, UpdatedAt: updatedAt}, nil
}

func update(t *testing.T, s *memoryUserStore, body string) (int, map[string]any) {
	t.Helper()
	h := NewUserHandler(s, nil, nil)
	srv := handler.APIHandler(func(c *handler.Context) error {
		c.Session = &model.Session{UserID: "jack"}
		return handler.TypedWithBody(h.HandleUserUpdate)(c)
	})

	r := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func TestHandleUserUpdate(t *testing.T) {
	s := &memoryUserStore{bios: map[string]string{}}

	code, res := update(t, s, `{"bio":"  I make moderation templates.\r\n\r\n\r\nDM me!  "}`)
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, "I make moderation templates.\n\nDM me!", s.bios["jack"])
	assert.Equal(t, "I make moderation templates.\n\nDM me!", res["data"].(map[string]any)["bio"])

	code, _ = update(t, s, `{"bio":"`+strings.Repeat("a", model.MaxUserBioLength+1)+`"}`)
	assert.Equal(t, http.StatusBadRequest, code)

	code, _ = update(t, s, `{"bio":"1\n2\n3\n4\n5\n6"}`)
	assert.Equal(t, http.StatusBadRequest, code)

	code, _ = update(t, s, `{"bio":""}`)
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "", s.bios["jack"])
}

func TestHandleUserUpdateBanned(t *testing.T) {
	s := &memoryUserStore{bios: map[string]string{}, banned: true}

	code, res := update(t, s, `{"bio":"hi"}`)
	assert.Equal(t, http.StatusForbidden, code)
	assert.Equal(t, "marketplace_banned", res["error"].(map[string]any)["code"])
	assert.NotContains(t, s.bios, "jack")
}
