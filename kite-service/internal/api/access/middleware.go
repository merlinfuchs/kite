package access

import (
	"errors"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
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
			collaborator, err := m.appStore.Collaborator(c.Context(), appID, c.Session.UserID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return handler.ErrForbidden("missing_access", "Access to app missing")
				}
				return err
			}

			c.UserAppRole = collaborator.Role
		} else {
			c.UserAppRole = model.AppCollaboratorRoleOwner
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

func (m *AccessManager) VariableAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		variableID := c.Param("variableID")
		appID := c.Param("appID")

		variable, err := m.variableStore.Variable(c.Context(), variableID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_variable", "Variable not found")
			}
			return err
		}

		// We assume that app access has already been checked
		if variable.AppID != appID {
			return handler.ErrForbidden("missing_access", "Access to variable missing")
		}

		c.Variable = variable
		return next(c)
	}
}

func (m *AccessManager) MessageAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		messageID := c.Param("messageID")
		appID := c.Param("appID")

		message, err := m.messageStore.Message(c.Context(), messageID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_message", "Message not found")
			}
			return err
		}

		// We assume that app access has already been checked
		if message.AppID != appID {
			return handler.ErrForbidden("missing_access", "Access to message missing")
		}

		c.Message = message
		return next(c)
	}
}

func (m *AccessManager) EventListenerAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		listenerID := c.Param("listenerID")
		appID := c.Param("appID")

		eventListener, err := m.eventListenerStore.EventListener(c.Context(), listenerID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_event_listener", "Event listener not found")
			}
			return err
		}

		// We assume that app access has already been checked
		if eventListener.AppID != appID {
			return handler.ErrForbidden("missing_access", "Access to event listener missing")
		}

		c.EventListener = eventListener
		return next(c)
	}
}
