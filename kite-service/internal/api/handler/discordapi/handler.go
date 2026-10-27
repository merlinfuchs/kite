package discordapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

type DiscordAPIHandler struct {
	appStateManager store.AppStateManager
	logStore        store.LogStore
}

func NewDiscordAPIHandler(appStateManager store.AppStateManager, logStore store.LogStore) *DiscordAPIHandler {
	return &DiscordAPIHandler{
		appStateManager: appStateManager,
		logStore:        logStore,
	}
}

// HandleDiscordAPIRequest sends a request to the Discord API with the app's
// bot token, for API tokens with the discord scope of the endpoint's access.
// Changes are logged so the app's admins see them.
func (h *DiscordAPIHandler) HandleDiscordAPIRequest(c *handler.Context, req wire.DiscordAPIRequestRequest) (*wire.DiscordAPIRequestResponse, error) {
	// The dashboard has blocks for this, and tokens are what scopes limit.
	if c.APIToken == nil {
		return nil, handler.ErrForbidden("api_token_required", "Only API tokens can use this endpoint")
	}

	apiReq, err := flow.DiscordAPIRequestOutsideFlow(req.Operation, req.PathParams, req.Query, req.Body)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_discord_api_request", err.Error())
	}

	access := model.APITokenAccessForMethod(apiReq.Method)
	if err := handler.RequireAPITokenAccess(c.APIToken, model.APITokenResourceDiscord, access); err != nil {
		return nil, err
	}

	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrBadRequest("app_not_running", "The app isn't running, enable it first")
		}
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	apiReq.Reason = api.AuditLogReason(req.Reason)
	if apiReq.Reason == "" {
		apiReq.Reason = api.AuditLogReason(fmt.Sprintf("Kite API token %q", c.APIToken.Name))
	}

	res, err := state.AppDiscordAPIRequest(c.Context(), apiReq)
	if access == model.APITokenAccessWrite {
		h.logRequest(c, req.Operation, apiReq.Method+" "+apiReq.Path, err)
	}
	if err != nil {
		// APIHandler turns errors from Discord into 400s.
		return nil, err
	}

	if len(res) == 0 {
		res = json.RawMessage("null")
	}
	return &wire.DiscordAPIRequestResponse{Result: res}, nil
}

func (h *DiscordAPIHandler) logRequest(c *handler.Context, operation string, request string, reqErr error) {
	entry := model.LogEntry{
		AppID:     c.App.ID,
		Level:     model.LogLevelInfo,
		Message:   fmt.Sprintf("Discord API request %s (%s) by API token %q", operation, request, c.APIToken.Name),
		CreatedAt: time.Now().UTC(),
	}
	if reqErr != nil {
		entry.Level = model.LogLevelWarn
		entry.Message += fmt.Sprintf(" failed: %s", reqErr)
	}

	if err := h.logStore.CreateLogEntry(c.Context(), entry); err != nil {
		slog.Error(
			"Failed to create Discord API request log entry",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
	}
}
