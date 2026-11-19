package bundleinstall

import (
	"cmp"
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// checkSharedParameters refuses channels and roles of other servers for an app
// on the Kite bot. Its flows can only reach its own server, see the sharedbot
// package, so they would fail when they run. Roles also have to be below the
// user's highest role, like Discord requires to give them, so Manage Server
// doesn't let them hand out roles through the bot they couldn't give.
func (h *BundleInstallHandler) checkSharedParameters(
	c *handler.Context,
	values map[string]string,
	parameterOf func(key string) (flow.FlowParameterType, string),
) error {
	if !c.App.IsShared() {
		return nil
	}

	// The server's channels and roles with their positions, loaded once
	// they're needed.
	ids := map[string]map[string]int{}
	for key, value := range values {
		if value == "" {
			continue
		}

		typ, label := parameterOf(key)
		var kind string
		switch typ {
		case flow.FlowParameterTypeChannel, flow.FlowParameterTypeCategory:
			kind = "channel"
		case flow.FlowParameterTypeRole:
			kind = "role"
		default:
			continue
		}

		if ids[kind] == nil {
			list, err := h.sharedGuildIDs(c, kind)
			if err != nil {
				return err
			}
			ids[kind] = list
		}
		position, ok := ids[kind][value]
		if !ok {
			return handler.ErrBadRequest("invalid_parameter", fmt.Sprintf("%s isn't a %s of this server", label, kind))
		}
		if member := c.SharedAppMember; kind == "role" && member != nil && !member.CanGiveRole(position) {
			return handler.ErrBadRequest("invalid_parameter", fmt.Sprintf("%s is a role at or above your highest role, so you can't give it", label))
		}
	}
	return nil
}

// bundleParameterOf looks up the type and label of the bundle's parameters.
func bundleParameterOf(b *bundle.Bundle) func(key string) (flow.FlowParameterType, string) {
	return func(key string) (flow.FlowParameterType, string) {
		for _, p := range b.Parameters {
			if p.Key == key {
				return p.Type, cmp.Or(p.Label, p.Key)
			}
		}
		return "", key
	}
}

// sharedGuildIDs returns the channels or roles of the app's server, by ID,
// with their positions.
func (h *BundleInstallHandler) sharedGuildIDs(c *handler.Context, kind string) (map[string]int, error) {
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	guildID := c.App.DiscordGuildID.String
	ids := map[string]int{}
	if kind == "role" {
		list, err := state.AppGuildRoles(c.Context(), guildID)
		if err != nil {
			return nil, fmt.Errorf("failed to get roles: %w", err)
		}
		for _, role := range list {
			ids[role.ID.String()] = role.Position
		}
		return ids, nil
	}

	list, err := state.AppGuildChannels(c.Context(), guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get channels: %w", err)
	}
	for _, ch := range list {
		ids[ch.ID.String()] = ch.Position
	}
	return ids, nil
}
