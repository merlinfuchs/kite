package user

import (
	"errors"
	"fmt"
	"slices"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type UserHandler struct {
	userStore    store.UserStore
	adminUserIDs []string
	badges       *plan.PlanManager
}

func NewUserHandler(userStore store.UserStore, adminUserIDs []string, badges *plan.PlanManager) *UserHandler {
	return &UserHandler{
		userStore:    userStore,
		adminUserIDs: adminUserIDs,
		badges:       badges,
	}
}

func (h *UserHandler) HandlerUserGet(c *handler.Context) (*wire.UserGetResponse, error) {
	userID := c.Param("userID")

	// We only want to return the email for the current user
	withEmail := false
	if userID == "@me" {
		userID = c.Session.UserID
		withEmail = true
	}

	user, err := h.userStore.User(c.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_user", "User not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	res := wire.UserToWire(user, withEmail)
	res.IsAdmin = withEmail && slices.Contains(h.adminUserIDs, user.ID)
	if withEmail {
		badges, err := h.badges.UserBadges(c.Context(), []string{user.ID})
		if err != nil {
			return nil, fmt.Errorf("failed to get badges: %w", err)
		}
		res.Badges = wire.UserBadgesToWire(badges[user.ID])
	}
	return res, nil
}
