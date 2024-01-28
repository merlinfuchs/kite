package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type UserHandler struct {
	store store.UserStore
}

func NewHandler(store store.UserStore) *UserHandler {
	return &UserHandler{
		store: store,
	}
}

func (h *UserHandler) HandleUserGet(c *fiber.Ctx) error {
	session := session.GetSession(c)

	userID := c.Params("userID")
	if userID == "@me" {
		userID = session.UserID
	}

	user, err := h.store.GetUser(c.Context(), userID)
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_user", "User not found")
		}
		return err
	}

	return c.JSON(wire.UserGetResponse{
		Success: true,
		Data:    wire.UserToWire(user),
	})
}
