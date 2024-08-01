package access

import (
	"errors"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (m *AccessManager) AppAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		appID := c.Param("appID")

		app, err := m.appStore.App(c.Context(), appID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_app", "App not found")
			}
			return err
		}

		if app.OwnerUserID != c.Session.UserID {
			return handler.ErrForbidden("missing_access", "Access to app missing")
		}

		c.App = app
		return next(c)
	}
}

func (m *AccessManager) CommandAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		commanID := c.Param("commandID")
		appID := c.Param("appID")

		command, err := m.commandStore.Command(c.Context(), commanID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_command", "Command not found")
			}
			return err
		}

		// We assume that app access has already been checked
		if command.AppID != appID {
			return handler.ErrForbidden("missing_access", "Access to command missing")
		}

		c.Command = command
		return next(c)
	}
}
