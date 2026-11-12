package userbadge

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// UserBadgeHandler lets the admins give users the badges that don't follow
// from what they did, like contributor.
type UserBadgeHandler struct {
	userStore store.UserStore
}

func NewUserBadgeHandler(userStore store.UserStore) *UserBadgeHandler {
	return &UserBadgeHandler{userStore: userStore}
}

// HandleUserBadgeGrantList lists the badges admins gave, newest first.
func (h *UserBadgeHandler) HandleUserBadgeGrantList(c *handler.Context) (*wire.UserBadgeGrantListResponse, error) {
	grants, err := h.userStore.UserBadgeGrants(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get user badge grants: %w", err)
	}

	res := make([]*wire.UserBadgeGrant, len(grants))
	for i, g := range grants {
		res[i] = wire.UserBadgeGrantToWire(g)
	}
	return &res, nil
}

// HandleUserBadgeGrantCreate gives the user with the Discord ID a badge.
func (h *UserBadgeHandler) HandleUserBadgeGrantCreate(c *handler.Context, req wire.UserBadgeGrantCreateRequest) (*wire.UserBadgeGrantCreateResponse, error) {
	user, err := h.userStore.UserByDiscordID(c.Context(), req.DiscordID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_user", "no user of Kite has this Discord ID")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	grant := &model.UserBadgeGrant{
		UserID:          user.ID,
		Kind:            model.UserBadgeKind(req.Kind),
		GrantedByUserID: null.StringFrom(c.Session.UserID),
		CreatedAt:       time.Now().UTC(),
		User:            user,
	}
	if err := h.userStore.GrantUserBadge(c.Context(), grant); err != nil {
		return nil, fmt.Errorf("failed to grant user badge: %w", err)
	}
	return wire.UserBadgeGrantToWire(grant), nil
}

// HandleUserBadgeGrantDelete takes a badge an admin gave away again.
func (h *UserBadgeHandler) HandleUserBadgeGrantDelete(c *handler.Context) (*wire.UserBadgeGrantDeleteResponse, error) {
	err := h.userStore.RevokeUserBadge(c.Context(), c.Param("userID"), model.UserBadgeKind(c.Param("kind")))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_user_badge", "the user doesn't have this badge")
		}
		return nil, fmt.Errorf("failed to revoke user badge: %w", err)
	}
	return &wire.UserBadgeGrantDeleteResponse{}, nil
}
