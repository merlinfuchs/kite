package bundleinstall

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

type BundleInstallHandler struct {
	bundleInstallStore   store.BundleInstallStore
	commandHandler       *commandhandler.CommandHandler
	eventListenerHandler *eventlistener.EventListenerHandler
}

func NewBundleInstallHandler(
	bundleInstallStore store.BundleInstallStore,
	commandHandler *commandhandler.CommandHandler,
	eventListenerHandler *eventlistener.EventListenerHandler,
) *BundleInstallHandler {
	return &BundleInstallHandler{
		bundleInstallStore:   bundleInstallStore,
		commandHandler:       commandHandler,
		eventListenerHandler: eventListenerHandler,
	}
}

func (h *BundleInstallHandler) HandleBundleInstallList(c *handler.Context) (*wire.BundleInstallListResponse, error) {
	installs, err := h.bundleInstallStore.BundleInstallsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installs: %w", err)
	}

	res := make([]*wire.BundleInstall, len(installs))
	for i, install := range installs {
		res[i] = wire.BundleInstallToWire(install)
	}

	return &res, nil
}

// HandleBundleInstallCreate installs a bundle. Its commands and event
// listeners go through the same checks as an import, and are created together
// with the install in one transaction.
func (h *BundleInstallHandler) HandleBundleInstallCreate(c *handler.Context, req wire.BundleInstallCreateRequest) (*wire.BundleInstallCreateResponse, error) {
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

	installID := util.UniqueID()
	items := make([]model.BundleInstallItem, 0, len(commands)+len(eventListeners))

	for i, command := range commands {
		hash, err := command.FlowSource.ContentHash()
		if err != nil {
			return nil, fmt.Errorf("failed to hash command flow: %w", err)
		}
		command.BundleInstallID = null.StringFrom(installID)
		items = append(items, model.BundleInstallItem{
			Key:  req.Commands[i].Key,
			Type: model.BundleInstallItemTypeCommand,
			ID:   command.ID,
			Hash: hash,
		})
	}

	for i, listener := range eventListeners {
		hash, err := listener.FlowSource.ContentHash()
		if err != nil {
			return nil, fmt.Errorf("failed to hash event listener flow: %w", err)
		}
		listener.BundleInstallID = null.StringFrom(installID)
		items = append(items, model.BundleInstallItem{
			Key:  req.EventListeners[i].Key,
			Type: model.BundleInstallItemTypeEventListener,
			ID:   listener.ID,
			Hash: hash,
		})
	}

	settings := req.Settings
	if settings == nil {
		settings = map[string]string{}
	}

	install, err := h.bundleInstallStore.CreateBundleInstall(c.Context(), &model.BundleInstall{
		ID:            installID,
		Name:          req.Name,
		Description:   req.Description,
		Enabled:       true,
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		BundleID:      req.BundleID,
		BundleVersion: req.BundleVersion,
		Settings:      settings,
		Items:         items,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}, store.BundleInstallContents{
		Commands:       commands,
		EventListeners: eventListeners,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create install: %w", err)
	}

	return wire.BundleInstallToWire(install), nil
}
