package bundleinstall

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	commandhandler "github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	eventlistener "github.com/kitecloud/kite/kite-service/internal/api/handler/event_listener"
	messagehandler "github.com/kitecloud/kite/kite-service/internal/api/handler/message"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type BundleInstallHandler struct {
	bundleInstallStore   store.BundleInstallStore
	bundleStore          store.BundleStore
	commandStore         store.CommandStore
	eventListenerStore   store.EventListenerStore
	messageStore         store.MessageStore
	commandHandler       *commandhandler.CommandHandler
	eventListenerHandler *eventlistener.EventListenerHandler
	messageHandler       *messagehandler.MessageHandler
}

func NewBundleInstallHandler(
	bundleInstallStore store.BundleInstallStore,
	bundleStore store.BundleStore,
	commandStore store.CommandStore,
	eventListenerStore store.EventListenerStore,
	messageStore store.MessageStore,
	commandHandler *commandhandler.CommandHandler,
	eventListenerHandler *eventlistener.EventListenerHandler,
	messageHandler *messagehandler.MessageHandler,
) *BundleInstallHandler {
	return &BundleInstallHandler{
		bundleInstallStore:   bundleInstallStore,
		bundleStore:          bundleStore,
		commandStore:         commandStore,
		eventListenerStore:   eventListenerStore,
		messageStore:         messageStore,
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

	contents, err := h.installedContents(c)
	if err != nil {
		return nil, err
	}

	res := make([]*wire.BundleInstall, len(installs))
	for i, install := range installs {
		res[i] = contents.toWire(install)
	}

	return &res, nil
}

// HandleBundleInstallCreate installs items of a bundle with the settings
// filled in. They go through the same checks as an import, and are created
// together with the install in one transaction.
func (h *BundleInstallHandler) HandleBundleInstallCreate(c *handler.Context, req wire.BundleInstallCreateRequest) (*wire.BundleInstallCreateResponse, error) {
	b, err := h.currentBundle(c, req.BundleID)
	if err != nil {
		return nil, err
	}
	if req.BundleVersion != b.Version {
		return nil, handler.ErrBadRequest("bundle_version_changed", "the bundle changed, reload the page and try again")
	}

	settings := req.Settings
	if settings == nil {
		settings = map[string]string{}
	}

	sel, err := b.Select(req.Items, settings)
	if err != nil {
		var invalid *bundle.InvalidError
		if errors.As(err, &invalid) {
			return nil, handler.ErrBadRequest("invalid_bundle_install", invalid.Error())
		}
		return nil, fmt.Errorf("failed to select bundle items: %w", err)
	}

	commandReqs := make([]wire.CommandCreateRequest, len(sel.Commands))
	for i, cmd := range sel.Commands {
		commandReqs[i] = wire.CommandCreateRequest{FlowSource: cmd.FlowSource, Enabled: true}
		// The settings can make a block invalid, like a text that's too long.
		if err := commandReqs[i].Validate(); err != nil {
			return nil, err
		}
	}
	commands, err := h.commandHandler.PrepareImport(c, commandReqs)
	if err != nil {
		return nil, err
	}

	eventListenerReqs := make([]wire.EventListenerCreateRequest, len(sel.EventListeners))
	for i, listener := range sel.EventListeners {
		eventListenerReqs[i] = wire.EventListenerCreateRequest{Source: listener.Source, FlowSource: listener.FlowSource, Enabled: true}
		if err := eventListenerReqs[i].Validate(); err != nil {
			return nil, err
		}
	}
	eventListeners, err := h.eventListenerHandler.PrepareImport(c, eventListenerReqs)
	if err != nil {
		return nil, err
	}

	messageReqs := make([]wire.MessageCreateRequest, len(sel.Messages))
	for i, msg := range sel.Messages {
		messageReqs[i] = wire.MessageCreateRequest{
			Name:        msg.Name,
			Description: msg.Description,
			Data:        msg.Data,
			FlowSources: msg.FlowSources,
		}
		messageReqs[i].Sanitize()
		if err := messageReqs[i].Validate(); err != nil {
			return nil, err
		}
	}
	messages, err := h.messageHandler.PrepareImport(c, messageReqs)
	if err != nil {
		return nil, err
	}

	// Blocks point at the messages of the bundle by their key.
	messageIDs := make(map[string]string, len(messages))
	for i, msg := range messages {
		messageIDs[sel.Messages[i].ID] = msg.ID
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
			Key:  sel.Commands[i].ID,
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
			Key:  sel.EventListeners[i].ID,
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
			Key:  sel.Messages[i].ID,
			Type: model.BundleInstallItemTypeMessage,
			ID:   msg.ID,
			Hash: hash,
		})
	}

	install, err := h.bundleInstallStore.CreateBundleInstall(c.Context(), &model.BundleInstall{
		ID:            installID,
		Name:          b.Name,
		Description:   b.Description,
		Enabled:       true,
		AppID:         c.App.ID,
		CreatorUserID: c.Session.UserID,
		BundleID:      b.ID,
		BundleVersion: b.Version,
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
			return handler.ErrBadRequest("unknown_message", fmt.Sprintf("block %s uses message %s of the bundle, which the app doesn't have", node.ID, key))
		}
		flowSource.Nodes[i].Data.MessageTemplateID = id
	}
	return nil
}

// HandleBundleInstallUpdate updates the items of an install to the current
// version of its bundle. Items the user didn't change are replaced, edited ones
// only if the user chose to, and deleted ones aren't added again.
func (h *BundleInstallHandler) HandleBundleInstallUpdate(c *handler.Context, req wire.BundleInstallUpdateRequest) (*wire.BundleInstallUpdateResponse, error) {
	install, err := h.bundleInstall(c)
	if err != nil {
		return nil, err
	}

	b, err := h.currentBundle(c, install.BundleID)
	if err != nil {
		return nil, err
	}
	if req.BundleVersion != b.Version {
		return nil, handler.ErrBadRequest("bundle_version_changed", "the bundle changed, reload the page and try again")
	}
	if install.BundleVersion >= b.Version {
		return nil, handler.ErrBadRequest("bundle_install_up_to_date", "the install already has the latest version")
	}

	settings := make(map[string]string, len(install.Settings)+len(req.Settings))
	maps.Copy(settings, install.Settings)
	maps.Copy(settings, req.Settings)

	contents, err := h.installedContents(c)
	if err != nil {
		return nil, err
	}

	overwrite := make(map[string]bool, len(req.Overwrite))
	for _, key := range req.Overwrite {
		overwrite[key] = true
	}

	items := slices.Clone(install.Items)
	itemIndex := make(map[string]int, len(items))
	var keys []string
	for i, item := range items {
		hash, ok := contents.hash(install.ID, item)
		// Items that were removed from the bundle stay as they are.
		if !ok || !b.HasItem(item.Key) || (hash != item.Hash && !overwrite[item.Key]) {
			continue
		}
		itemIndex[item.Key] = i
		keys = append(keys, item.Key)
	}

	sel, err := b.Select(keys, settings)
	if err != nil {
		var invalid *bundle.InvalidError
		if errors.As(err, &invalid) {
			return nil, handler.ErrBadRequest("invalid_bundle_install", invalid.Error())
		}
		return nil, fmt.Errorf("failed to select bundle items: %w", err)
	}

	installs, err := h.bundleInstallStore.BundleInstallsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installs: %w", err)
	}
	messageIDs := contents.bundleMessageIDs(installs, install.BundleID)

	var changed store.BundleInstallContents

	for _, cmd := range sel.Commands {
		item := &items[itemIndex[cmd.ID]]
		current, ok := contents.commands[item.ID]
		if !ok {
			continue
		}
		if err := (wire.CommandUpdateRequest{FlowSource: cmd.FlowSource}).Validate(); err != nil {
			return nil, err
		}
		if err := remapMessageTemplates(&cmd.FlowSource, messageIDs); err != nil {
			return nil, err
		}
		if item.Hash, err = cmd.FlowSource.ContentHash(); err != nil {
			return nil, fmt.Errorf("failed to hash command flow: %w", err)
		}
		if contents.items[item.ID].hash == item.Hash {
			continue
		}
		updated, err := h.commandHandler.PrepareUpdate(c, current, cmd.FlowSource)
		if err != nil {
			return nil, err
		}
		changed.Commands = append(changed.Commands, updated)
	}

	for _, listener := range sel.EventListeners {
		item := &items[itemIndex[listener.ID]]
		current, ok := contents.eventListeners[item.ID]
		if !ok {
			continue
		}
		if err := (wire.EventListenerUpdateRequest{FlowSource: listener.FlowSource}).Validate(); err != nil {
			return nil, err
		}
		if err := remapMessageTemplates(&listener.FlowSource, messageIDs); err != nil {
			return nil, err
		}
		if item.Hash, err = listener.FlowSource.ContentHash(); err != nil {
			return nil, fmt.Errorf("failed to hash event listener flow: %w", err)
		}
		if contents.items[item.ID].hash == item.Hash {
			continue
		}
		updated, err := h.eventListenerHandler.PrepareUpdate(c, current, listener.FlowSource)
		if err != nil {
			return nil, err
		}
		changed.EventListeners = append(changed.EventListeners, updated)
	}

	for _, msg := range sel.Messages {
		item := &items[itemIndex[msg.ID]]
		current, ok := contents.messages[item.ID]
		if !ok {
			continue
		}
		// The name isn't part of the content, so the user's name is kept.
		msgReq := wire.MessageUpdateRequest{
			Name:        current.Name,
			Description: current.Description,
			Data:        msg.Data,
			FlowSources: msg.FlowSources,
		}
		msgReq.Sanitize()
		if err := msgReq.Validate(); err != nil {
			return nil, err
		}
		for key, flowSource := range msgReq.FlowSources {
			if err := remapMessageTemplates(&flowSource, messageIDs); err != nil {
				return nil, err
			}
			msgReq.FlowSources[key] = flowSource
		}
		updated := *current
		updated.Data = msgReq.Data
		updated.FlowSources = msgReq.FlowSources
		updated.UpdatedAt = time.Now().UTC()
		if item.Hash, err = updated.ContentHash(); err != nil {
			return nil, fmt.Errorf("failed to hash message: %w", err)
		}
		if contents.items[item.ID].hash == item.Hash {
			continue
		}
		changed.Messages = append(changed.Messages, &updated)
	}

	install.BundleVersion = b.Version
	install.Settings = settings
	install.Items = items
	install.UpdatedAt = time.Now().UTC()

	install, err = h.bundleInstallStore.UpdateBundleInstall(c.Context(), install, changed)
	if err != nil {
		return nil, fmt.Errorf("failed to update install: %w", err)
	}

	if err := contents.replace(changed); err != nil {
		return nil, err
	}
	return contents.toWire(install), nil
}

// HandleBundleInstallDelete deletes the install with everything it created
// that the app still has, including what the user changed.
func (h *BundleInstallHandler) HandleBundleInstallDelete(c *handler.Context) (*wire.BundleInstallDeleteResponse, error) {
	install, err := h.bundleInstall(c)
	if err != nil {
		return nil, err
	}

	if err := h.bundleInstallStore.DeleteBundleInstall(c.Context(), install.ID); err != nil {
		return nil, fmt.Errorf("failed to delete install: %w", err)
	}

	return &wire.BundleInstallDeleteResponse{}, nil
}

// currentBundle returns the version of the bundle apps install.
func (h *BundleInstallHandler) currentBundle(c *handler.Context, id string) (*bundle.Bundle, error) {
	_, version, err := h.bundleStore.BundleCurrentVersion(c.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle", "bundle not found")
		}
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	return bundle.Parse(version.Content)
}

func (h *BundleInstallHandler) bundleInstall(c *handler.Context) (*model.BundleInstall, error) {
	install, err := h.bundleInstallStore.BundleInstall(c.Context(), c.App.ID, c.Param("installID"))
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_bundle_install", "install not found")
		}
		return nil, fmt.Errorf("failed to get install: %w", err)
	}
	return install, nil
}
