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
	maxBundlesPerUser    int
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
	maxBundlesPerUser int,
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
		maxBundlesPerUser:    maxBundlesPerUser,
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

	ids := make([]string, len(installs))
	for i, install := range installs {
		ids[i] = install.BundleID
	}
	bundles, err := h.bundleStore.BundleSummaries(c.Context(), ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundles: %w", err)
	}
	bundleByID := make(map[string]*model.Bundle, len(bundles))
	for _, b := range bundles {
		bundleByID[b.ID] = b
	}

	res := make([]*wire.BundleInstall, len(installs))
	for i, install := range installs {
		res[i] = contents.toWire(install)
		setInstallSource(res[i], bundleByID[install.BundleID], c.Session.UserID)
	}

	return &res, nil
}

// setInstallSource says who made the bundle of the install and which version
// the app can update to. A deleted bundle is nil and leaves it all empty.
func setInstallSource(res *wire.BundleInstall, b *model.Bundle, userID string) {
	if b == nil {
		return
	}
	res.Official = !b.OwnerUserID.Valid
	res.Yours = b.OwnedBy(userID)
	res.Removed = b.RemovedAt.Valid
	if b.Owner != nil {
		res.Owner = wire.BundleOwnerToWire(b.Owner)
	}
	if b.Listing != nil {
		res.Icon = b.Listing.Icon
	}
	if b.VisibleTo(userID) {
		res.AvailableVersion = null.IntFrom(int64(b.InstallVersion()))
	}
}

// HandleBundleInstallCreate installs items of a bundle with the parameters
// filled in. They go through the same checks as an import, and are created
// together with the install in one transaction. An app that has the bundle
// already gets them added to its install, which has to be up to date.
func (h *BundleInstallHandler) HandleBundleInstallCreate(c *handler.Context, req wire.BundleInstallCreateRequest) (*wire.BundleInstallCreateResponse, error) {
	b, err := h.currentBundle(c, req.BundleID)
	if err != nil {
		return nil, err
	}
	if req.BundleVersion != b.Version {
		return nil, handler.ErrBadRequest("bundle_version_changed", "the template changed, reload the page and try again")
	}

	install, err := h.installOf(c, b.ID)
	if err != nil {
		return nil, err
	}
	isNew := install == nil
	if isNew {
		install = &model.BundleInstall{
			ID:            util.UniqueID(),
			Name:          b.Name,
			Description:   b.Description,
			Enabled:       true,
			AppID:         c.App.ID,
			CreatorUserID: c.Session.UserID,
			BundleID:      b.ID,
			BundleVersion: b.Version,
			CreatedAt:     time.Now().UTC(),
		}
	} else if install.BundleVersion < b.Version {
		return nil, handler.ErrBadRequest("bundle_install_outdated", fmt.Sprintf("update the template to version %d first", b.Version))
	}

	// The items it adds get the values the install has, and the new ones.
	params := map[string]string{}
	maps.Copy(params, install.Parameters)
	maps.Copy(params, req.Parameters)

	sel, err := selectItems(b, req.Items, params, "invalid_bundle_install")
	if err != nil {
		return nil, err
	}
	if err := sel.Customize(req.Messages); err != nil {
		return nil, handler.ErrBadRequest("invalid_bundle_install", err.Error())
	}

	// Blocks can also use the messages and variables the install has.
	contents, err := h.installedContents(c)
	if err != nil {
		return nil, err
	}

	// Items the install still has can't be added again, deleted ones can.
	added := util.IDSet(req.Items)
	kept := make([]model.BundleInstallItem, 0, len(install.Items))
	for _, item := range install.Items {
		if _, ok := added[item.Key]; !ok {
			kept = append(kept, item)
			continue
		}
		if _, ok := contents.hash(install.ID, item); ok {
			return nil, handler.ErrBadRequest("already_installed", fmt.Sprintf("the app already has %s of the template", item.Key))
		}
	}

	created, items, err := h.prepareContents(c, sel, contents.bundleIDs(install), null.StringFrom(install.ID))
	if err != nil {
		return nil, err
	}
	install.Items = append(kept, items...)
	install.Parameters = params

	if isNew {
		install.UpdatedAt = install.CreatedAt
		install, err = h.bundleInstallStore.CreateBundleInstall(c.Context(), install, created)
		if err != nil {
			if errors.Is(err, store.ErrAlreadyExists) {
				return nil, handler.ErrBadRequest("bundle_install_changed", "the template was added at the same time, reload the page and try again")
			}
			return nil, fmt.Errorf("failed to create install: %w", err)
		}
	} else {
		install, err = h.updateInstall(c, install, store.BundleInstallChanges{Created: created})
		if err != nil {
			return nil, err
		}
	}
	if err := contents.replace(created); err != nil {
		return nil, err
	}
	return contents.toWire(install), nil
}

// installOf returns the install of the bundle in the app, or nil if it has
// none. An app has each bundle at most once.
func (h *BundleInstallHandler) installOf(c *handler.Context, bundleID string) (*model.BundleInstall, error) {
	installs, err := h.bundleInstallStore.BundleInstallsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installs: %w", err)
	}
	for _, install := range installs {
		if install.BundleID == bundleID {
			return install, nil
		}
	}
	return nil, nil
}

// updateInstall saves the install and the changes to its items, unless it
// was changed since it was loaded.
func (h *BundleInstallHandler) updateInstall(c *handler.Context, install *model.BundleInstall, changes store.BundleInstallChanges) (*model.BundleInstall, error) {
	from := install.UpdatedAt
	install.UpdatedAt = time.Now().UTC()
	install, err := h.bundleInstallStore.UpdateBundleInstall(c.Context(), install, from, changes, c.Session.UserID)
	if err != nil {
		if errors.Is(err, store.ErrOutdated) {
			return nil, handler.ErrBadRequest("bundle_install_changed", "the install changed, reload the page and try again")
		}
		return nil, fmt.Errorf("failed to update install: %w", err)
	}
	return install, nil
}

// prepareContents checks the selected items like an import and returns them
// ready to be created, pointing at the messages and variables in ids and the
// ones it adds. installID is null for imports, which aren't linked to the
// bundle.
func (h *BundleInstallHandler) prepareContents(c *handler.Context, sel *bundle.Selection, ids bundleIDs, installID null.String) (store.BundleInstallContents, []model.BundleInstallItem, error) {
	commandReqs := make([]wire.CommandCreateRequest, len(sel.Commands))
	for i, cmd := range sel.Commands {
		commandReqs[i] = wire.CommandCreateRequest{FlowSource: cmd.FlowSource, Enabled: true}
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
		if err := ids.remapAll(msg.FlowSources); err != nil {
			return store.BundleInstallContents{}, nil, err
		}
		// A message the user changed while adding it counts as changed, so
		// updates of the bundle keep it.
		hashed := *msg
		if def := sel.Messages[i].Default; def != nil {
			hashed.Data = *def
		}
		hash, err := hashed.ContentHash()
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
		return nil, handler.ErrBadRequest("bundle_version_changed", "the template changed, reload the page and try again")
	}
	if install.BundleVersion >= b.Version {
		return nil, handler.ErrBadRequest("bundle_install_up_to_date", "the install already has the latest version")
	}

	// Parameters the new version doesn't have anymore are dropped.
	params := make(map[string]string, len(b.Parameters))
	for _, p := range b.Parameters {
		if value, ok := req.Parameters[p.Key]; ok {
			params[p.Key] = value
		} else if value, ok := install.Parameters[p.Key]; ok {
			params[p.Key] = value
		}
	}

	contents, err := h.installedContents(c)
	if err != nil {
		return nil, err
	}

	overwrite := util.IDSet(req.Overwrite)

	items := slices.Clone(install.Items)
	itemIndex := make(map[string]int, len(items))
	var keys []string
	for i, item := range items {
		hash, ok := contents.hash(install.ID, item)
		_, overwritten := overwrite[item.Key]
		// Items that were removed from the bundle stay as they are, and so do
		// variables, which only have a name.
		if !ok || !b.HasItem(item.Key) || item.Type == model.BundleInstallItemTypeVariable || (hash != item.Hash && !overwritten) {
			continue
		}
		itemIndex[item.Key] = i
		keys = append(keys, item.Key)
	}

	sel, err := selectItems(b, keys, params, "invalid_bundle_install")
	if err != nil {
		return nil, err
	}

	ids := contents.bundleIDs(install)

	// The new version can use messages and variables the app doesn't have
	// yet, which are added to the install.
	missing, err := ids.missing(b, sel.Flows())
	if err != nil {
		return nil, err
	}
	var created store.BundleInstallContents
	before := len(items)
	if len(missing) > 0 {
		used, err := selectItems(b, missing, params, "invalid_bundle_install")
		if err != nil {
			return nil, err
		}
		var createdItems []model.BundleInstallItem
		if created, createdItems, err = h.prepareContents(c, used, ids, null.StringFrom(install.ID)); err != nil {
			return nil, err
		}
		items = append(items, createdItems...)
	}

	var changed store.BundleInstallContents

	for _, cmd := range sel.Commands {
		item := &items[itemIndex[cmd.ID]]
		current := contents.commands[item.ID]
		next, ok, err := contents.nextFlow(item, cmd.FlowSource, current.FlowSource, req.Parameters, ids, func(f flow.FlowData) error {
			return wire.CommandUpdateRequest{FlowSource: f}.Validate()
		})
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		updated, err := h.commandHandler.PrepareUpdate(c, current, next)
		if err != nil {
			return nil, err
		}
		changed.Commands = append(changed.Commands, updated)
	}

	for _, listener := range sel.EventListeners {
		item := &items[itemIndex[listener.ID]]
		current := contents.eventListeners[item.ID]
		next, ok, err := contents.nextFlow(item, listener.FlowSource, current.FlowSource, req.Parameters, ids, func(f flow.FlowData) error {
			return wire.EventListenerUpdateRequest{FlowSource: f}.Validate()
		})
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		updated, err := h.eventListenerHandler.PrepareUpdate(c, current, next)
		if err != nil {
			return nil, err
		}
		changed.EventListeners = append(changed.EventListeners, updated)
	}

	for _, msg := range sel.Messages {
		item := &items[itemIndex[msg.ID]]
		current := contents.messages[item.ID]
		for key, flowSource := range msg.FlowSources {
			msg.FlowSources[key] = keepParameterValues(flowSource, current.FlowSources[key], req.Parameters)
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
		if err := ids.remapAll(msgReq.FlowSources); err != nil {
			return nil, err
		}
		updated := *current
		updated.Data = msgReq.Data
		updated.FlowSources = msgReq.FlowSources
		updated.UpdatedAt = time.Now().UTC()
		if item.Hash, err = updated.ContentHash(); err != nil {
			return nil, fmt.Errorf("failed to hash message: %w", err)
		}
		if contents.items[item.ID].hash == item.Hash && sameParameters(updated.FlowSources, current.FlowSources) {
			continue
		}
		changed.Messages = append(changed.Messages, &updated)
	}

	install.BundleVersion = b.Version
	install.Parameters = params
	// The messages and variables it created replace the ones the user deleted.
	kept := slices.DeleteFunc(items[:before], func(item model.BundleInstallItem) bool {
		return slices.Contains(missing, item.Key)
	})
	install.Items = slices.Concat(kept, items[before:])

	changes := store.BundleInstallChanges{Created: created, Updated: changed}
	install, err = h.updateInstall(c, install, changes)
	if err != nil {
		return nil, err
	}

	if err := contents.replace(changes.Created); err != nil {
		return nil, err
	}
	if err := contents.replace(changes.Updated); err != nil {
		return nil, err
	}
	return contents.toWire(install), nil
}

// nextFlow returns the flow of the new version of a command or event listener
// ready to be saved, and false if it's the one the app has. It sets the hash
// of the item.
func (ic *installedContents) nextFlow(item *model.BundleInstallItem, next, current flow.FlowData, set map[string]string, ids bundleIDs, validate func(flow.FlowData) error) (flow.FlowData, bool, error) {
	next = keepParameterValues(next, current, set)
	if err := validate(next); err != nil {
		return next, false, err
	}
	if err := ids.remap(&next); err != nil {
		return next, false, err
	}
	hash, err := next.ContentHash()
	if err != nil {
		return next, false, fmt.Errorf("failed to hash flow: %w", err)
	}
	item.Hash = hash
	return next, ic.items[item.ID].hash != hash || !slices.Equal(next.Parameters, current.Parameters), nil
}

// HandleBundleInstallDelete deletes the install with everything it created
// that the app still has, including what the user changed. Message templates
// and variables that other flows of the app use are kept.
func (h *BundleInstallHandler) HandleBundleInstallDelete(c *handler.Context) (*wire.BundleInstallDeleteResponse, error) {
	install, err := h.bundleInstall(c)
	if err != nil {
		return nil, err
	}

	all, err := h.appContents(c)
	if err != nil {
		return nil, err
	}
	used := usedByOthers(all, install.ID)
	var keep []string
	for _, item := range install.Items {
		if used[item.ID] {
			keep = append(keep, item.ID)
		}
	}

	if err := h.bundleInstallStore.DeleteBundleInstall(c.Context(), install.ID, keep, c.Session.UserID); err != nil {
		return nil, fmt.Errorf("failed to delete install: %w", err)
	}

	return &wire.BundleInstallDeleteResponse{}, nil
}

// usedByOthers returns the IDs of the message templates and variables that
// flows the install didn't create use, also through the buttons of its
// messages they use.
func usedByOthers(all store.BundleInstallContents, installID string) map[string]bool {
	used := map[string]bool{}
	add := func(flowSource flow.FlowData) {
		for _, node := range flowSource.Nodes {
			if id := node.Data.MessageTemplateID; id != "" {
				used[id] = true
			}
			if id := node.Data.VariableID; id != "" {
				used[id] = true
			}
		}
	}
	other := func(id null.String) bool { return id.String != installID }

	for _, cmd := range all.Commands {
		if other(cmd.BundleInstallID) {
			add(cmd.FlowSource)
		}
	}
	for _, listener := range all.EventListeners {
		if other(listener.BundleInstallID) {
			add(listener.FlowSource)
		}
	}
	own := map[string]*model.Message{}
	for _, msg := range all.Messages {
		if !other(msg.BundleInstallID) {
			own[msg.ID] = msg
			continue
		}
		for _, flowSource := range msg.FlowSources {
			add(flowSource)
		}
	}

	// Its messages that are kept also keep what their buttons use.
	for changed := true; changed; {
		changed = false
		for id, msg := range own {
			if !used[id] {
				continue
			}
			delete(own, id)
			for _, flowSource := range msg.FlowSources {
				add(flowSource)
			}
			changed = true
		}
	}
	return used
}

// keepParameterValues returns the new version of a flow with the values the
// user gave its parameters, except the ones the update sets and ones that
// don't fit a type the parameter changed to.
func keepParameterValues(updated flow.FlowData, current flow.FlowData, set map[string]string) flow.FlowData {
	values := current.ParameterValues()
	for key := range set {
		delete(values, key)
	}
	for _, p := range updated.Parameters {
		if value, ok := values[p.Key]; ok && p.CheckValue(value) != nil {
			delete(values, p.Key)
		}
	}
	updated, _ = updated.WithParameterValues(values)
	return updated
}

// sameParameters reports whether the flows of two versions of a message have
// the same parameters. The content hash leaves them out, so a new label or
// type alone is still an update.
func sameParameters(a map[string]flow.FlowData, b map[string]flow.FlowData) bool {
	for key, flowSource := range a {
		if !slices.Equal(flowSource.Parameters, b[key].Parameters) {
			return false
		}
	}
	return true
}

// selectItems is Select of the bundle, with items or parameters that can't be
// installed as a bad request with the code.
func selectItems(b *bundle.Bundle, keys []string, params map[string]string, code string) (*bundle.Selection, error) {
	sel, err := b.Select(keys, params)
	if err != nil {
		var invalid *bundle.InvalidError
		if errors.As(err, &invalid) {
			return nil, handler.ErrBadRequest(code, invalid.Error())
		}
		return nil, fmt.Errorf("failed to select bundle items: %w", err)
	}
	return sel, nil
}

// currentBundle returns the version of the bundle apps install, if the user
// can see the bundle. The app a bundle is made from has the originals of its
// items, so it can't install it.
func (h *BundleInstallHandler) currentBundle(c *handler.Context, id string) (*bundle.Bundle, error) {
	b, err := h.bundleStore.Bundle(c.Context(), id)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	if err != nil || !b.VisibleTo(c.Session.UserID) {
		return nil, handler.ErrNotFound("unknown_bundle", "template not found")
	}
	if b.SourceAppID.Valid && b.SourceAppID.String == c.App.ID {
		return nil, handler.ErrBadRequest("bundle_source_app", "the template is made from this app, which has its items already")
	}
	return bundle.Parse(b.Content)
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
