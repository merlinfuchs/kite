package plugin

import (
	"strings"

	"github.com/diamondburned/arikawa/v3/utils/ws"
)

type Registry struct {
	plugins map[string]Plugin
	// order keeps the plugins in the order they were registered, so they are
	// listed in a stable order instead of the random order of the map.
	order []string
}

func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

func (r *Registry) Register(plugins ...Plugin) {
	for _, p := range plugins {
		if _, exists := r.plugins[p.ID()]; !exists {
			r.order = append(r.order, p.ID())
		}
		r.plugins[p.ID()] = p
	}
}

func (r *Registry) Plugins() []Plugin {
	plugins := make([]Plugin, 0, len(r.order))
	for _, id := range r.order {
		plugins = append(plugins, r.plugins[id])
	}
	return plugins
}

func (r *Registry) Plugin(id string) Plugin {
	return r.plugins[id]
}

// EventTypesForResources resolves "plugin_id:resource_id" pairs to the Discord
// event types those plugin resources subscribe to.
//
// Unknown plugins and resource IDs that don't name an event are skipped: a
// resource may be a command rather than an event, and a plugin may have been
// removed from the registry while instances of it still exist in the database.
func (r *Registry) EventTypesForResources(resources []string) []ws.EventType {
	seen := make(map[ws.EventType]struct{})
	var types []ws.EventType

	for _, resource := range resources {
		pluginID, resourceID, ok := strings.Cut(resource, ":")
		if !ok {
			continue
		}

		plugin := r.plugins[pluginID]
		if plugin == nil {
			continue
		}

		for _, event := range plugin.Events() {
			if event.ID != resourceID {
				continue
			}

			eventType := event.Type.DiscordEventType()
			if _, dup := seen[eventType]; dup {
				continue
			}

			seen[eventType] = struct{}{}
			types = append(types, eventType)
		}
	}

	return types
}
