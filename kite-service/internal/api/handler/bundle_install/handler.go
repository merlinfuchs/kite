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
	variablehandler "github.com/kitecloud/kite/kite-service/internal/api/handler/variable"
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
	variableStore        store.VariableStore
	commandHandler       *commandhandler.CommandHandler
	eventListenerHandler *eventlistener.EventListenerHandler
	messageHandler       *messagehandler.MessageHandler
	variableHandler      *variablehandler.VariableHandler
	appStateManager      store.AppStateManager
}

func NewBundleInstallHandler(
	bundleInstallStore store.BundleInstallStore,
	bundleStore store.BundleStore,
	commandStore store.CommandStore,
	eventListenerStore store.EventListenerStore,
	messageStore store.MessageStore,
	variableStore store.VariableStore,
	commandHandler *commandhandler.CommandHandler,
	eventListenerHandler *eventlistener.EventListenerHandler,
	messageHandler *messagehandler.MessageHandler,
	variableHandler *variablehandler.VariableHandler,
	appStateManager store.AppStateManager,
) *BundleInstallHandler {
	return &BundleInstallHandler{
		bundleInstallStore:   bundleInstallStore,
		bundleStore:          bundleStore,
		commandStore:         commandStore,
		eventListenerStore:   eventListenerStore,
		messageStore:         messageStore,
		variableStore:        variableStore,
		commandHandler:       commandHandler,
		eventListenerHandler: eventListenerHandler,
		messageHandler:       messageHandler,
		variableHandler:      variableHandler,
		appStateManager:      appStateManager,
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

	// Blocks can also use the messages and variables of earlier installs.
	contents, err := h.installedContents(c)
	if err != nil {
		return nil, err
	}
	installs, err := h.bundleInstallStore.BundleInstallsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installs: %w", err)
	}

	installID := util.UniqueID()
	created, items, err := h.prepareContents(c, sel, contents.bundleIDs(installs, b.ID), null.StringFrom(installID))
	if err != nil {
		return nil, err
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
	}, created)
	if err != nil {
		return nil, fmt.Errorf("failed to create install: %w", err)
	}

	return wire.BundleInstallToWire(install), nil
}

// prepareContents checks the selected items like an import and returns them
// ready to be created, pointing at the messages and variables in ids and the
// ones it adds. installID is null for imports, which aren't linked to the
// bundle.
func (h *BundleInstallHandler) prepareContents(c *handler.Context, sel *bundle.Selection, ids bundleIDs, installID null.String) (store.BundleInstallContents, []model.BundleInstallItem, error) {
	commandReqs := make([]wire.CommandCreateRequest, len(sel.Commands))
	for i, cmd := range sel.Commands {
		commandReqs[i] = wire.CommandCreateRequest{FlowSource: cmd.FlowSource, Enabled: true}
		// The settings can make a block invalid, like a text that's too long.
		if err := commandReqs[i].Validate(); err != nil {
			return store.BundleInstallContents{}, nil, err
		}
	}
	commands, err := h.commandHandler.PrepareImport(c, commandReqs)
	if err != nil {
		return store.BundleInstallContents{}, nil, err
	}

	eventListenerReqs := make([]wire.EventListenerCreateRequest, len(sel.EventListeners))
	for i, listener := range sel.EventListeners {
		eventListenerReqs[i] = wire.EventListenerCreateRequest{Source: listener.Source, FlowSource: listener.FlowSource, Enabled: true}
		if err := eventListenerReqs[i].Validate(); err != nil {
			return store.BundleInstallContents{}, nil, err
		}
	}
	eventListeners, err := h.eventListenerHandler.PrepareImport(c, eventListenerReqs)
	if err != nil {
		return store.BundleInstallContents{}, nil, err
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
			return store.BundleInstallContents{}, nil, err
		}
	}
	messages, err := h.messageHandler.PrepareImport(c, messageReqs)
	if err != nil {
		return store.BundleInstallContents{}, nil, err
	}

	variableReqs := make([]wire.VariableCreateRequest, len(sel.Variables))
	for i, v := range sel.Variables {
		variableReqs[i] = wire.VariableCreateRequest{Name: v.Name, Scoped: v.Scoped}
	}
	variables, err := h.variableHandler.PrepareImport(c, variableReqs)
	if err != nil {
		return store.BundleInstallContents{}, nil, err
	}

	for i, msg := range messages {
		ids.messages[sel.Messages[i].ID] = msg.ID
	}
	for i, variable := range variables {
		ids.variables[sel.Variables[i].ID] = variable.ID
	}

	items := make([]model.BundleInstallItem, 0, len(commands)+len(eventListeners)+len(messages)+len(variables))

	for i, command := range commands {
		if err := ids.remap(&command.FlowSource); err != nil {
			return store.BundleInstallContents{}, nil, err
		}
		hash, err := command.FlowSource.ContentHash()
		if err != nil {
			return store.BundleInstallContents{}, nil, fmt.Errorf("failed to hash command flow: %w", err)
		}
		command.BundleInstallID = installID
		items = append(items, model.BundleInstallItem{
			Key:  sel.Commands[i].ID,
			Type: model.BundleInstallItemTypeCommand,
			ID:   command.ID,
			Hash: hash,
		})
	}

	for i, listener := range eventListeners {
		if err := ids.remap(&listener.FlowSource); err != nil {
			return store.BundleInstallContents{}, nil, err
		}
		hash, err := listener.FlowSource.ContentHash()
		if err != nil {
			return store.BundleInstallContents{}, nil, fmt.Errorf("failed to hash event listener flow: %w", err)
		}
		listener.BundleInstallID = installID
		items = append(items, model.BundleInstallItem{
			Key:  sel.EventListeners[i].ID,
			Type: model.BundleInstallItemTypeEventListener,
			ID:   listener.ID,
			Hash: hash,
		})
	}

	for i, msg := range messages {
		for key, flowSource := range msg.FlowSources {
			if err := ids.remap(&flowSource); err != nil {
				return store.BundleInstallContents{}, nil, err
			}
			msg.FlowSources[key] = flowSource
		}
		hash, err := msg.ContentHash()
		if err != nil {
			return store.BundleInstallContents{}, nil, fmt.Errorf("failed to hash message: %w", err)
		}
		msg.BundleInstallID = installID
		items = append(items, model.BundleInstallItem{
			Key:  sel.Messages[i].ID,
			Type: model.BundleInstallItemTypeMessage,
			ID:   msg.ID,
			Hash: hash,
		})
	}

	for i, variable := range variables {
		variable.BundleInstallID = installID
		items = append(items, model.BundleInstallItem{
			Key:  sel.Variables[i].ID,
			Type: model.BundleInstallItemTypeVariable,
			ID:   variable.ID,
		})
	}

	return store.BundleInstallContents{
		Commands:       commands,
		EventListeners: eventListeners,
		Messages:       messages,
		Variables:      variables,
	}, items, nil
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
		// Items that were removed from the bundle stay as they are, and so do
		// variables, which only have a name.
		if !ok || !b.HasItem(item.Key) || item.Type == model.BundleInstallItemTypeVariable || (hash != item.Hash && !overwrite[item.Key]) {
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
	ids := contents.bundleIDs(installs, install.BundleID)

	// The new version can use variables the app doesn't have yet.
	created, err := h.missingVariables(c, b, sel, ids)
	if err != nil {
		return nil, err
	}
	for _, v := range created {
		v.variable.BundleInstallID = null.StringFrom(install.ID)
		ids.variables[v.key] = v.variable.ID
		items = append(items, model.BundleInstallItem{
			Key:  v.key,
			Type: model.BundleInstallItemTypeVariable,
			ID:   v.variable.ID,
		})
	}

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
		if err := ids.remap(&cmd.FlowSource); err != nil {
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
		if err := ids.remap(&listener.FlowSource); err != nil {
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
			if err := ids.remap(&flowSource); err != nil {
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

	changes := store.BundleInstallChanges{Updated: changed}
	for _, v := range created {
		changes.Created.Variables = append(changes.Created.Variables, v.variable)
	}
	install, err = h.bundleInstallStore.UpdateBundleInstall(c.Context(), install, changes)
	if err != nil {
		return nil, fmt.Errorf("failed to update install: %w", err)
	}

	if err := contents.replace(changes.Created); err != nil {
		return nil, err
	}
	if err := contents.replace(changes.Updated); err != nil {
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

type createdVariable struct {
	key      string
	variable *model.Variable
}

// missingVariables returns the variables of the bundle that the selected items
// use and the app doesn't have, ready to be created.
func (h *BundleInstallHandler) missingVariables(c *handler.Context, b *bundle.Bundle, sel *bundle.Selection, ids bundleIDs) ([]createdVariable, error) {
	flows := make([]flow.FlowData, 0, len(sel.Commands)+len(sel.EventListeners))
	for _, cmd := range sel.Commands {
		flows = append(flows, cmd.FlowSource)
	}
	for _, listener := range sel.EventListeners {
		flows = append(flows, listener.FlowSource)
	}
	for _, msg := range sel.Messages {
		for _, flowSource := range msg.FlowSources {
			flows = append(flows, flowSource)
		}
	}

	var keys []string
	var reqs []wire.VariableCreateRequest
	for _, flowSource := range flows {
		for _, node := range flowSource.Nodes {
			key := node.Data.VariableID
			if key == "" || ids.variables[key] != "" || slices.Contains(keys, key) {
				continue
			}
			// Blocks that use one the bundle doesn't have fail on remapping.
			v, ok := b.Variable(key)
			if !ok {
				continue
			}
			keys = append(keys, key)
			reqs = append(reqs, wire.VariableCreateRequest{Name: v.Name, Scoped: v.Scoped})
		}
	}

	variables, err := h.variableHandler.PrepareImport(c, reqs)
	if err != nil {
		return nil, err
	}
	res := make([]createdVariable, len(variables))
	for i, variable := range variables {
		res[i] = createdVariable{key: keys[i], variable: variable}
	}
	return res, nil
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
