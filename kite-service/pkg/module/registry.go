package module

type Registry struct {
	modules map[string]Module
}

func NewRegistry() *Registry {
	return &Registry{
		modules: make(map[string]Module),
	}
}

func (r *Registry) Register(modules ...Module) {
	for _, m := range modules {
		r.modules[m.ID()] = m
	}
}

func (r *Registry) Modules() []Module {
	modules := make([]Module, 0, len(r.modules))
	for _, module := range r.modules {
		modules = append(modules, module)
	}
	return modules
}

func (r *Registry) Module(id string) Module {
	return r.modules[id]
}
