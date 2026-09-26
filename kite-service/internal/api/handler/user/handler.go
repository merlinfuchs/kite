package user

import (
	"errors"
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type UserHandler struct {
	userStore store.UserStore
}

func NewUserHandler(userStore store.UserStore) *UserHandler {
	return &UserHandler{
		userStore: userStore,
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

	return wire.UserToWire(user, withEmail), nil
}

func (h *UserHandler) HandlerUserUpdate(c *handler.Context, req wire.UserUpdateRequest) (*wire.UserUpdateResponse, error) {
	userID := c.Param("userID")
	if userID != "@me" && userID != c.Session.UserID {
		return nil, handler.ErrForbidden("forbidden", "You can only update your own user")
	}

	user, err := h.userStore.UpdateUserHideEmail(c.Context(), c.Session.UserID, req.HideEmail)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_user", "User not found")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return wire.UserToWire(user, true), nil
}
