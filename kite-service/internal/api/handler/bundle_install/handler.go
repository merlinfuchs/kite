package bundleinstall

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	commandhandler "github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	eventlistener "github.com/kitecloud/kite/kite-service/internal/api/handler/event_listener"
	messagehandler "github.com/kitecloud/kite/kite-service/internal/api/handler/message"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type BundleInstallHandler struct {
	bundleInstallStore   store.BundleInstallStore
	commandHandler       *commandhandler.CommandHandler
	eventListenerHandler *eventlistener.EventListenerHandler
	messageHandler       *messagehandler.MessageHandler
}

func NewBundleInstallHandler(
	bundleInstallStore store.BundleInstallStore,
	commandHandler *commandhandler.CommandHandler,
	eventListenerHandler *eventlistener.EventListenerHandler,
	messageHandler *messagehandler.MessageHandler,
) *BundleInstallHandler {
	return &BundleInstallHandler{
		bundleInstallStore:   bundleInstallStore,
		commandHandler:       commandHandler,
		eventListenerHandler: eventListenerHandler,
		messageHandler:       messageHandler,
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

// HandleBundleInstallCreate installs a bundle. Its commands, event listeners
// and messages go through the same checks as an import, and are created
// together with the install in one transaction.
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

	messageReqs := make([]wire.MessageCreateRequest, len(req.Messages))
	for i, msg := range req.Messages {
		messageReqs[i] = msg.Message
	}
	messages, err := h.messageHandler.PrepareImport(c, messageReqs)
	if err != nil {
		return nil, err
	}

	// Blocks point at the messages of the bundle by their key.
	messageIDs := make(map[string]string, len(messages))
	for i, msg := range messages {
		messageIDs[req.Messages[i].Key] = msg.ID
	}

	installID := util.UniqueID()
	items := make([]model.BundleInstallItem, 0, len(commands)+len(eventListeners)+len(messages))

	for i, command := range commands {
		if err := remapMessageTemplates(&command.FlowSource, messageIDs); err != nil {
			return nil, err
		}
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
		if err := remapMessageTemplates(&listener.FlowSource, messageIDs); err != nil {
			return nil, err
		}
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

	for i, msg := range messages {
		for key, flowSource := range msg.FlowSources {
			if err := remapMessageTemplates(&flowSource, messageIDs); err != nil {
				return nil, err
			}
			msg.FlowSources[key] = flowSource
		}
		hash, err := msg.ContentHash()
		if err != nil {
			return nil, fmt.Errorf("failed to hash message: %w", err)
		}
		msg.BundleInstallID = null.StringFrom(installID)
		items = append(items, model.BundleInstallItem{
			Key:  req.Messages[i].Key,
			Type: model.BundleInstallItemTypeMessage,
			ID:   msg.ID,
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
		Messages:       messages,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create install: %w", err)
	}

	return wire.BundleInstallToWire(install), nil
}

// remapMessageTemplates points the blocks that use a message of the bundle at
// the message template created for it. Blocks can't use templates outside of
// the bundle, they don't exist in the app.
func remapMessageTemplates(flowSource *flow.FlowData, messageIDs map[string]string) error {
	for i, node := range flowSource.Nodes {
		key := node.Data.MessageTemplateID
		if key == "" {
			continue
		}
		id, ok := messageIDs[key]
		if !ok {
			return handler.ErrBadRequest("unknown_message", fmt.Sprintf("block %s uses message %s, which isn't part of the bundle", node.ID, key))
		}
		flowSource.Nodes[i].Data.MessageTemplateID = id
	}
	return nil
}
