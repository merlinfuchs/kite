package module

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	commandhandler "github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	eventlistener "github.com/kitecloud/kite/kite-service/internal/api/handler/event_listener"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

type ModuleHandler struct {
	moduleStore          store.ModuleStore
	commandHandler       *commandhandler.CommandHandler
	eventListenerHandler *eventlistener.EventListenerHandler
}

func NewModuleHandler(
	moduleStore store.ModuleStore,
	commandHandler *commandhandler.CommandHandler,
	eventListenerHandler *eventlistener.EventListenerHandler,
) *ModuleHandler {
	return &ModuleHandler{
		moduleStore:          moduleStore,
		commandHandler:       commandHandler,
		eventListenerHandler: eventListenerHandler,
	}
}

func (h *ModuleHandler) HandleModuleList(c *handler.Context) (*wire.ModuleListResponse, error) {
	modules, err := h.moduleStore.ModulesByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get modules: %w", err)
	}

	res := make([]*wire.Module, len(modules))
	for i, module := range modules {
		res[i] = wire.ModuleToWire(module)
	}

	return &res, nil
}

// HandleModuleCreate installs a template. Its commands and event listeners go
// through the same checks as an import, and are only created if all of them
// pass.
func (h *ModuleHandler) HandleModuleCreate(c *handler.Context, req wire.ModuleCreateRequest) (*wire.ModuleCreateResponse, error) {
	commandReqs := make([]wire.CommandCreateRequest, len(req.Commands))
	for i, cmd := range req.Commands {
		commandReqs[i] = cmd.Command
	}
	commands, err := h.commandHandler.PrepareImport(c, commandReqs)
	if err != nil {
		return nil, err
	}

	eventListenerReqs := make([]wire.EventListenerCreateRequest, len(req.EventListeners))
	for i, listener := range req.EventListeners {
		eventListenerReqs[i] = listener.EventListener
	}
	eventListeners, err := h.eventListenerHandler.PrepareImport(c, eventListenerReqs)
	if err != nil {
		return nil, err
	}

	moduleID := util.UniqueID()
	items := make([]model.ModuleItem, 0, len(commands)+len(eventListeners))

	for i, command := range commands {
		hash, err := command.FlowSource.ContentHash()
		if err != nil {
			return nil, fmt.Errorf("failed to hash command flow: %w", err)
		}
		command.ModuleID = null.StringFrom(moduleID)
		items = append(items, model.ModuleItem{
			Key:  req.Commands[i].Key,
			Type: model.ModuleItemTypeCommand,
			ID:   command.ID,
			Hash: hash,
		})
	}

	for i, listener := range eventListeners {
		hash, err := listener.FlowSource.ContentHash()
		if err != nil {
			return nil, fmt.Errorf("failed to hash event listener flow: %w", err)
		}
		listener.ModuleID = null.StringFrom(moduleID)
		items = append(items, model.ModuleItem{
			Key:  req.EventListeners[i].Key,
			Type: model.ModuleItemTypeEventListener,
			ID:   listener.ID,
			Hash: hash,
		})
	}

	settings := req.Settings
	if settings == nil {
		settings = map[string]string{}
	}

	module, err := h.moduleStore.CreateModule(c.Context(), &model.Module{
		ID:              moduleID,
		Name:            req.Name,
		Description:     req.Description,
		Enabled:         true,
		AppID:           c.App.ID,
		CreatorUserID:   c.Session.UserID,
		TemplateID:      req.TemplateID,
		TemplateVersion: req.TemplateVersion,
		Settings:        settings,
		Items:           items,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create module: %w", err)
	}

	if err := h.commandHandler.CreateCommands(c, commands); err != nil {
		return nil, err
	}
	if err := h.eventListenerHandler.CreateEventListeners(c, eventListeners); err != nil {
		return nil, err
	}

	return wire.ModuleToWire(module), nil
}
