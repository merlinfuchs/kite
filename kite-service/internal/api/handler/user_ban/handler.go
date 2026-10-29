package userban

import (
	"errors"
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/notify"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// UserBanHandler lets the admins ban users from features, like the
// marketplace.
type UserBanHandler struct {
	userStore store.UserStore
	notifier  *notify.Notifier
}

func NewUserBanHandler(userStore store.UserStore, notifier *notify.Notifier) *UserBanHandler {
	return &UserBanHandler{userStore: userStore, notifier: notifier}
}

// HandleUserBanList lists all bans, newest first.
func (h *UserBanHandler) HandleUserBanList(c *handler.Context) (*wire.UserBanListResponse, error) {
	bans, err := h.userStore.UserBans(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get user bans: %w", err)
	}

	res := make([]*wire.UserBan, len(bans))
	for i, b := range bans {
		res[i] = wire.UserBanToWire(b)
	}
	return &res, nil
}

// HandleUserBanCreate bans a user from a feature and tells them why.
func (h *UserBanHandler) HandleUserBanCreate(c *handler.Context, req wire.UserBanCreateRequest) (*wire.UserBanCreateResponse, error) {
	user, err := h.userStore.User(c.Context(), req.UserID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_user", "user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	ban := &model.UserBan{
		UserID:         user.ID,
		Feature:        model.UserBanFeature(req.Feature),
		Reason:         req.Reason,
		BannedByUserID: null.StringFrom(c.Session.UserID),
		CreatedAt:      time.Now().UTC(),
	}
	if err := h.userStore.BanUser(c.Context(), ban); err != nil {
		return nil, fmt.Errorf("failed to ban user: %w", err)
	}

	h.notifier.UserBanned(user, ban)

	ban.User = user
	return wire.UserBanToWire(ban), nil
}

// HandleUserBanDelete lifts a user's ban from a feature.
func (h *UserBanHandler) HandleUserBanDelete(c *handler.Context) (*wire.UserBanDeleteResponse, error) {
	err := h.userStore.UnbanUser(c.Context(), c.Param("userID"), model.UserBanFeature(c.Param("feature")))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_user_ban", "the user isn't banned from this")
		}
		return nil, fmt.Errorf("failed to unban user: %w", err)
	}
	return &wire.UserBanDeleteResponse{}, nil
}
