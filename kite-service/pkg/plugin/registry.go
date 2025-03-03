package plugin

type Registry struct {
	plugins map[string]Plugin
}

func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]Plugin),
	}
}

func (r *Registry) Register(plugin Plugin) {
	r.plugins[plugin.ID()] = plugin
}

func (r *Registry) Plugins() []Plugin {
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	return plugins
}

func (r *Registry) Plugin(id string) Plugin {
	return r.plugins[id]
}
