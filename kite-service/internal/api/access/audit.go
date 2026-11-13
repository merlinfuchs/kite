package access

import (
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

// auditRoute is what a route of an app does, for its audit log.
type auditRoute struct {
	Action     string
	TargetType string
	// IDParam is the path parameter with the ID of the target, if any.
	IDParam string
}

// auditRoutes are the routes under /v1/apps/{appID} that show up in the audit
// log, by method and the rest of their path. Routes that change nothing, like
// tests, or that run all the time, like Discord API requests, are left out.
var auditRoutes = map[string]auditRoute{
	"PUT":                            {"app.update", "app", ""},
	"PUT /status":                    {"app.status_update", "app", ""},
	"PUT /credit-notices":            {"app.credit_notices_update", "app", ""},
	"PUT /token":                     {"app.token_update", "app", ""},
	"POST /collaborators":            {"collaborator.invite", "collaborator", ""},
	"PATCH /collaborators/{userID}":  {"collaborator.role_update", "collaborator", "userID"},
	"DELETE /collaborators/{userID}": {"collaborator.remove", "collaborator", "userID"},
	"DELETE /collaborators/invites/{discordUsername}": {"collaborator.invite_revoke", "collaborator", "discordUsername"},
	"POST /share-codes": {"share_code.create", "share_code", ""},

	"PUT /billing/subscriptions/{subscriptionID}/plan": {"billing.plan_update", "subscription", "subscriptionID"},

	"POST /commands":                                    {"command.create", "command", ""},
	"POST /commands/import":                             {"command.import", "command", ""},
	"POST /commands/deploy":                             {"command.deploy", "command", ""},
	"PATCH /commands/{commandID}":                       {"command.update", "command", "commandID"},
	"DELETE /commands/{commandID}":                      {"command.delete", "command", "commandID"},
	"PUT /commands/{commandID}/enabled":                 {"command.enabled_update", "command", "commandID"},
	"POST /event-listeners":                             {"event_listener.create", "event_listener", ""},
	"POST /event-listeners/import":                      {"event_listener.import", "event_listener", ""},
	"PATCH /event-listeners/{listenerID}":               {"event_listener.update", "event_listener", "listenerID"},
	"DELETE /event-listeners/{listenerID}":              {"event_listener.delete", "event_listener", "listenerID"},
	"PUT /event-listeners/{listenerID}/enabled":         {"event_listener.enabled_update", "event_listener", "listenerID"},
	"POST /event-listeners/{listenerID}/webhook-secret": {"event_listener.webhook_secret_update", "event_listener", "listenerID"},

	"POST /plugins":                   {"plugin.create", "plugin", ""},
	"PATCH /plugins/{pluginID}":       {"plugin.update", "plugin", "pluginID"},
	"DELETE /plugins/{pluginID}":      {"plugin.delete", "plugin", "pluginID"},
	"PUT /plugins/{pluginID}/enabled": {"plugin.enabled_update", "plugin", "pluginID"},

	"POST /variables":                       {"variable.create", "variable", ""},
	"POST /variables/import":                {"variable.import", "variable", ""},
	"PATCH /variables/{variableID}":         {"variable.update", "variable", "variableID"},
	"DELETE /variables/{variableID}":        {"variable.delete", "variable", "variableID"},
	"PUT /variables/{variableID}/values":    {"variable.value_set", "variable", "variableID"},
	"DELETE /variables/{variableID}/values": {"variable.value_delete", "variable", "variableID"},

	"POST /secrets":              {"secret.create", "secret", ""},
	"PATCH /secrets/{secretID}":  {"secret.update", "secret", "secretID"},
	"DELETE /secrets/{secretID}": {"secret.delete", "secret", "secretID"},

	"PUT /integrations/{integrationID}":    {"integration.connect", "integration", "integrationID"},
	"PATCH /integrations/{integrationID}":  {"integration.update", "integration", "integrationID"},
	"DELETE /integrations/{integrationID}": {"integration.remove", "integration", "integrationID"},

	"POST /messages":                                      {"message.create", "message", ""},
	"POST /messages/import":                               {"message.import", "message", ""},
	"PATCH /messages/{messageID}":                         {"message.update", "message", "messageID"},
	"DELETE /messages/{messageID}":                        {"message.delete", "message", "messageID"},
	"POST /messages/{messageID}/instances":                {"message.send", "message", "messageID"},
	"PUT /messages/{messageID}/instances/{instanceID}":    {"message.sent_update", "message", "messageID"},
	"DELETE /messages/{messageID}/instances/{instanceID}": {"message.sent_delete", "message", "messageID"},

	"POST /bundle-installs":                         {"template.install", "template", ""},
	"POST /bundle-installs/{installID}/update":      {"template.update", "template", "installID"},
	"DELETE /bundle-installs/{installID}":           {"template.uninstall", "template", "installID"},
	"PATCH /bundle-installs/{installID}/parameters": {"template.parameters_update", "template", "installID"},
	"POST /bundle-imports":                          {"bundle.import", "bundle", ""},
	"POST /bundles":                                 {"bundle.create", "bundle", ""},
	"POST /bundles/{bundleID}/versions":             {"bundle.version_create", "bundle", "bundleID"},

	"PATCH /state/user":                     {"bot.profile_update", "bot", ""},
	"DELETE /state/guilds/{guildID}":        {"guild.leave", "guild", "guildID"},
	"PATCH /state/guilds/{guildID}/profile": {"guild.profile_update", "guild", "guildID"},
}

const auditRoutePrefix = "/v1/apps/{appID}"

// auditRouteFor returns the audit route of a route pattern, like
// "PATCH /v1/apps/{appID}/commands/{commandID}".
func auditRouteFor(pattern string) (auditRoute, bool) {
	method, path, ok := strings.Cut(pattern, " ")
	if !ok || !strings.HasPrefix(path, auditRoutePrefix) {
		return auditRoute{}, false
	}

	key := method
	if rest := strings.TrimPrefix(path, auditRoutePrefix); rest != "" {
		key += " " + rest
	}

	route, ok := auditRoutes[key]
	return route, ok
}

// AuditLog adds an entry to the app's audit log for each change that
// succeeds. It goes after AppAccess, so the app is known.
func (m *AccessManager) AuditLog(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		if err := next(c); err != nil {
			return err
		}

		if c.Method() == http.MethodGet || c.App == nil || c.Session == nil {
			return nil
		}

		route, ok := auditRouteFor(c.RoutePattern())
		if !ok {
			return nil
		}

		entry := auditEntry(c, route)
		if err := m.appStore.CreateAppAuditLogEntry(c.Context(), entry); err != nil {
			slog.Error(
				"Failed to create audit log entry",
				slog.String("app_id", c.App.ID),
				slog.String("action", entry.Action),
				slog.String("error", err.Error()),
			)
		}
		return nil
	}
}

func auditEntry(c *handler.Context, route auditRoute) *model.AppAuditLogEntry {
	entry := &model.AppAuditLogEntry{
		AppID:      c.App.ID,
		UserID:     null.StringFrom(c.Session.UserID),
		Action:     route.Action,
		TargetType: route.TargetType,
		CreatedAt:  time.Now().UTC(),
	}
	if route.IDParam != "" {
		entry.TargetID = c.Param(route.IDParam)
	}
	if c.APIToken != nil {
		entry.APITokenID = null.StringFrom(c.APIToken.ID)
	}

	// The name before the change, from the access middleware of the item.
	switch {
	case c.Command != nil:
		entry.TargetName = c.Command.Name
	case c.EventListener != nil:
		entry.TargetName = eventListenerName(c.EventListener.Description, string(c.EventListener.Type))
	case c.Message != nil:
		entry.TargetName = c.Message.Name
	case c.Variable != nil:
		entry.TargetName = c.Variable.Name
	case c.PluginInstance != nil:
		entry.TargetName = c.PluginInstance.PluginID
	case route.TargetType == "app":
		entry.TargetID = c.App.ID
		entry.TargetName = c.App.Name
	}

	// The response has the new item, like the one that was created.
	if resp, ok := readAuditResponse(c.Response()); ok {
		if entry.TargetID == "" {
			entry.TargetID = resp.ID
		}
		if resp.Name != "" {
			entry.TargetName = resp.Name
		} else if resp.Type != "" && route.TargetType == "event_listener" {
			entry.TargetName = eventListenerName(resp.Description, resp.Type)
		}
		// Enabling and disabling are clearer than "updated enabled".
		if resp.Enabled != nil && strings.HasSuffix(route.Action, ".enabled_update") {
			base := strings.TrimSuffix(route.Action, ".enabled_update")
			if *resp.Enabled {
				entry.Action = base + ".enable"
			} else {
				entry.Action = base + ".disable"
			}
		}
	}

	if target := c.AuditTarget(); target != nil {
		if target.Type != "" {
			entry.TargetType = target.Type
		}
		if target.ID != "" {
			entry.TargetID = target.ID
		}
		if target.Name != "" {
			entry.TargetName = target.Name
		}
		if target.Action != "" {
			entry.Action = target.Action
		}
	}

	// These are clear without a name: the app is the one of the log, and
	// who left is who did it.
	if (entry.TargetType == "app" && entry.Action != "app.update") || entry.Action == "collaborator.leave" {
		entry.TargetName = ""
	}

	return entry
}

func eventListenerName(description string, typ string) string {
	if description != "" {
		return description
	}
	return typ
}

// auditResponse is what the audit log reads of the item in a response.
type auditResponse struct {
	ID          string
	Name        string
	Type        string
	Description string
	Enabled     *bool
}

// readAuditResponse reads the fields of a response that's a struct by their
// names, without encoding the whole item, which can have a big flow. Fields
// of other types are left empty.
func readAuditResponse(resp any) (auditResponse, bool) {
	v := reflect.ValueOf(resp)
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return auditResponse{}, false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return auditResponse{}, false
	}

	str := func(name string) string {
		if f := v.FieldByName(name); f.Kind() == reflect.String {
			return f.String()
		}
		return ""
	}

	res := auditResponse{
		ID:          str("ID"),
		Name:        str("Name"),
		Type:        str("Type"),
		Description: str("Description"),
	}
	if f := v.FieldByName("Enabled"); f.Kind() == reflect.Bool {
		enabled := f.Bool()
		res.Enabled = &enabled
	}
	return res, true
}

// RequireSettingsAccess only lets roles through that can change the app's
// settings, not developers. It goes after AppAccess.
func (m *AccessManager) RequireSettingsAccess(next handler.HandlerFunc) handler.HandlerFunc {
	return func(c *handler.Context) error {
		if !c.UserAppRole.CanManageSettings() {
			return handler.ErrForbidden("missing_permissions", "Only admins and the owner of the app can do this")
		}
		return next(c)
	}
}
