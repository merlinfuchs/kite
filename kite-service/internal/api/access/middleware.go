package access

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (m *AccessManager) AppAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		appID := c.Param("appID")
		if c.APIToken != nil && !c.APIToken.CanUseApp(appID) {
			return handler.ErrForbidden("missing_access", "The API token can't use this app")
		}

		app, err := m.appStore.App(c.Context(), appID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_app", "App not found")
			}
			return err
		}

		member, err := m.sharedAppMember(c, app)
		if err != nil {
			return err
		}
		decided, err := sharedAppAccess(member)
		switch {
		case err != nil:
			return err
		case decided:
			// Everyone with Manage Server has the same say.
			c.UserAppRole = model.AppCollaboratorRoleOwner
			c.SharedAppMember = member
		case app.OwnerUserID != c.Session.UserID:
			collaborator, err := m.appStore.Collaborator(c.Context(), appID, c.Session.UserID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return handler.ErrForbidden("missing_access", "Access to app missing")
				}
				return err
			}

			c.UserAppRole = collaborator.Role
		default:
			c.UserAppRole = model.AppCollaboratorRoleOwner
		}

		features, err := m.planManager.AppFeatures(c.Context(), appID)
		if err != nil {
			return err
		}

		c.App = app
		c.Features = features
		return next(c)
	}
}

// sharedAppMember is the user in the server of an app on the Kite bot, nil
// for other apps. Who can change those follows Manage Server in the server,
// which Discord is asked about. Without the bot in the server it can't be, so
// the app's owner and collaborators can still open it then. Nothing of it
// runs without the bot, and adding the bot again needs Manage Server.
func (m *AccessManager) sharedAppMember(c *handler.Context, app *model.App) (*model.SharedAppMember, error) {
	// Tokens are refused for these apps anyway, see CustomBotOnly.
	if !app.IsShared() || m.sharedMembers == nil || c.APIToken != nil {
		return nil, nil
	}

	user, err := m.userStore.User(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	member, err := m.sharedMembers.Member(c.Context(), app, user.DiscordID)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// sharedAppAccess decides the access of a member of the server of an app on
// the Kite bot. It's left to the app's owner and collaborators when Discord
// can't say, see sharedAppMember.
func sharedAppAccess(member *model.SharedAppMember) (decided bool, err error) {
	switch {
	case member == nil || !member.BotInServer:
		return false, nil
	case member.ManagesServer:
		return true, nil
	default:
		return true, handler.ErrForbidden("missing_access", "You need the Manage Server permission in the server of this app.")
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

		// Scoped in the query so another app's message is simply not found.
		message, err := m.messageStore.Message(c.Context(), appID, messageID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_message", "Message not found")
			}
			return err
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

func (m *AccessManager) PluginInstanceAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		pluginID := c.Param("pluginID")
		appID := c.Param("appID")

		pluginInstance, err := m.pluginInstanceStore.PluginInstance(c.Context(), appID, pluginID)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return handler.ErrNotFound("unknown_plugin_instance", "Plugin instance not found")
			}
			return err
		}

		// We assume that app access has already been checked
		if pluginInstance.AppID != appID {
			return handler.ErrForbidden("missing_access", "Access to plugin instance missing")
		}

		c.PluginInstance = pluginInstance
		return next(c)
	}
}

// CustomBotOnly refuses changes to apps on the Kite bot, which only run the
// official templates they add, unless the route allows them, see
// HandlerGroup.AllowSharedApp. Reading stays allowed. API tokens, and so MCP,
// can't use these apps at all yet. See design/kite-lite.md.
func (m *AccessManager) CustomBotOnly(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		if c.App == nil || !c.App.IsShared() {
			return next(c)
		}
		if c.APIToken != nil {
			return handler.ErrForbidden("kite_bot_without_api", "Apps on the Kite bot can't be used with API tokens or MCP.")
		}
		if c.Method() != http.MethodGet && !c.SharedAppAllowed() {
			return handler.ErrForbidden("needs_custom_bot", "Apps on the Kite bot only run templates. Create an app with a custom bot to build your own.")
		}
		return next(c)
	}
}
