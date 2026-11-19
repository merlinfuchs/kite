package bundleinstall

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// checkSharedParameters refuses channels and roles of other servers for an app
// on the Kite bot. Its flows can only reach its own server, see the sharedbot
// package, so they would fail when they run.
func (h *BundleInstallHandler) checkSharedParameters(
	c *handler.Context,
	values map[string]string,
	parameterOf func(key string) (flow.FlowParameterType, string),
) error {
	if !c.App.IsShared() || h.appStateManager == nil {
		return nil
	}

	var channels, roles map[string]bool
	for key, value := range values {
		if value == "" {
			continue
		}

		typ, label := parameterOf(key)
		switch typ {
		case flow.FlowParameterTypeChannel, flow.FlowParameterTypeCategory:
			if channels == nil {
				ids, err := h.sharedGuildIDs(c, false)
				if err != nil {
					return err
				}
				channels = ids
			}
			if !channels[value] {
				return handler.ErrBadRequest("invalid_parameter", fmt.Sprintf("%s isn't a channel of this server", label))
			}
		case flow.FlowParameterTypeRole:
			if roles == nil {
				ids, err := h.sharedGuildIDs(c, true)
				if err != nil {
					return err
				}
				roles = ids
			}
			if !roles[value] {
				return handler.ErrBadRequest("invalid_parameter", fmt.Sprintf("%s isn't a role of this server", label))
			}
		}
	}
	return nil
}

// bundleParameterOf looks up the type and label of the bundle's parameters.
func bundleParameterOf(b *bundle.Bundle) func(key string) (flow.FlowParameterType, string) {
	return func(key string) (flow.FlowParameterType, string) {
		for _, p := range b.Parameters {
			if p.Key == key {
				return p.Type, p.Label
			}
		}
		return "", key
	}
}

// sharedGuildIDs returns the IDs of the channels or roles of the app's server.
func (h *BundleInstallHandler) sharedGuildIDs(c *handler.Context, roles bool) (map[string]bool, error) {
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	guildID := c.App.DiscordGuildID.String
	ids := map[string]bool{}
	if roles {
		list, err := state.AppGuildRoles(c.Context(), guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to get roles: %w", err)
		}
		for _, role := range list {
			ids[role.ID.String()] = true
		}
		return ids, nil
	}

	list, err := state.AppGuildChannels(c.Context(), guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	for _, ch := range list {
		ids[ch.ID.String()] = true
	}
	return ids, nil
}
