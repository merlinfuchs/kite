package appsetup

import (
	"errors"
	"fmt"

	arikawastore "github.com/diamondburned/arikawa/v3/state/store"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// AppSetupHandler reports how far a new app is through the setup checklist.
type AppSetupHandler struct {
	appStateManager store.AppStateManager
	appStore        store.AppStore
	usageStore      store.UsageStore
}

func NewAppSetupHandler(
	appStateManager store.AppStateManager,
	appStore store.AppStore,
	usageStore store.UsageStore,
) *AppSetupHandler {
	return &AppSetupHandler{
		appStateManager: appStateManager,
		appStore:        appStore,
		usageStore:      usageStore,
	}
}

func (h *AppSetupHandler) HandleAppSetupGet(c *handler.Context) (*wire.AppSetupGetResponse, error) {
	invited, err := h.invited(c)
	if err != nil {
		return nil, err
	}

	hasFlow, err := h.appStore.HasAppFlows(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check flows: %w", err)
	}

	ranFlow, err := h.usageStore.HasAppUsage(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check usage: %w", err)
	}

	return &wire.AppSetupGetResponse{
		Invited: invited,
		HasFlow: hasFlow,
		RanFlow: ranFlow,
	}, nil
}

// invited reports whether the bot is in any server. Only the gateway knows,
// so a bot that isn't connected counts as not invited.
func (h *AppSetupHandler) invited(c *handler.Context) (bool, error) {
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if errors.Is(err, store.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get app state: %w", err)
	}

	guilds, err := state.AppGuilds(c.Context())
	if errors.Is(err, arikawastore.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to get guilds: %w", err)
	}
	return len(guilds) > 0, nil
}
