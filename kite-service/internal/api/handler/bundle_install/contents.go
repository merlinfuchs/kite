package bundleinstall

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// installedContents are the commands, event listeners, messages and variables
// of the app that installs created, by their ID.
type installedContents struct {
	commands       map[string]*model.Command
	eventListeners map[string]*model.EventListener
	messages       map[string]*model.Message
	items          map[string]installedItem
}

type installedItem struct {
	installID string
	itemType  model.BundleInstallItemType
	// hash is the hash of what the item is now.
	hash string
}

// appContents returns all commands, event listeners, messages and variables
// of the app.
func (h *BundleInstallHandler) appContents(c *handler.Context) (store.BundleInstallContents, error) {
	var res store.BundleInstallContents
	var err error
	if res.Commands, err = h.commandStore.CommandsByApp(c.Context(), c.App.ID); err != nil {
		return res, fmt.Errorf("failed to get commands: %w", err)
	}
	if res.EventListeners, err = h.eventListenerStore.EventListenersByApp(c.Context(), c.App.ID); err != nil {
		return res, fmt.Errorf("failed to get event listeners: %w", err)
	}
	if res.Messages, err = h.messageStore.MessagesByApp(c.Context(), c.App.ID); err != nil {
		return res, fmt.Errorf("failed to get messages: %w", err)
	}
	if res.Variables, err = h.variableStore.VariablesByAppWithoutTotals(c.Context(), c.App.ID); err != nil {
		return res, fmt.Errorf("failed to get variables: %w", err)
	}
	return res, nil
}

func (h *BundleInstallHandler) installedContents(c *handler.Context) (*installedContents, error) {
	all, err := h.bundleInstallStore.InstalledContentsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installed contents: %w", err)
	}

	res := &installedContents{
		commands:       map[string]*model.Command{},
		eventListeners: map[string]*model.EventListener{},
		messages:       map[string]*model.Message{},
		items:          map[string]installedItem{},
	}
	if err := res.replace(all); err != nil {
		return nil, err
	}
	return res, nil
}

// replace puts the contents in, which installs created.
func (ic *installedContents) replace(contents store.BundleInstallContents) error {
	for _, command := range contents.Commands {
		hash, err := command.FlowSource.ContentHash()
		if err != nil {
			return fmt.Errorf("failed to hash command flow: %w", err)
		}
		ic.commands[command.ID] = command
		ic.items[command.ID] = installedItem{command.BundleInstallID.String, model.BundleInstallItemTypeCommand, hash}
	}
	for _, listener := range contents.EventListeners {
		hash, err := listener.FlowSource.ContentHash()
		if err != nil {
			return fmt.Errorf("failed to hash event listener flow: %w", err)
		}
		ic.eventListeners[listener.ID] = listener
		ic.items[listener.ID] = installedItem{listener.BundleInstallID.String, model.BundleInstallItemTypeEventListener, hash}
	}
	for _, msg := range contents.Messages {
		hash, err := msg.ContentHash()
		if err != nil {
			return fmt.Errorf("failed to hash message: %w", err)
		}
		ic.messages[msg.ID] = msg
		ic.items[msg.ID] = installedItem{msg.BundleInstallID.String, model.BundleInstallItemTypeMessage, hash}
	}
	for _, variable := range contents.Variables {
		ic.items[variable.ID] = installedItem{variable.BundleInstallID.String, model.BundleInstallItemTypeVariable, ""}
	}
	return nil
}

// hash returns the hash of what the item of the install is now, or false if
// the user deleted it.
func (ic *installedContents) hash(installID string, item model.BundleInstallItem) (string, bool) {
	i, ok := ic.items[item.ID]
	if !ok || i.installID != installID || i.itemType != item.Type {
		return "", false
	}
	return i.hash, true
}

// bundleIDs returns the IDs of the messages and variables of the install that
// the app still has.
func (ic *installedContents) bundleIDs(install *model.BundleInstall) bundleIDs {
	ids := bundleIDs{messages: map[string]string{}, variables: map[string]string{}}
	for _, item := range install.Items {
		if _, ok := ic.hash(install.ID, item); !ok {
			continue
		}
		switch item.Type {
		case model.BundleInstallItemTypeMessage:
			ids.messages[item.Key] = item.ID
		case model.BundleInstallItemTypeVariable:
			ids.variables[item.Key] = item.ID
		}
	}
	return ids
}

// bundleIDs are the IDs of messages and variables of a bundle in the app, by
// their key in the bundle. Blocks of the bundle use them by their key.
type bundleIDs struct {
	messages  map[string]string
	variables map[string]string
}

// remapAll is remap for the flows of a message.
func (ids bundleIDs) remapAll(flowSources map[string]flow.FlowData) error {
	for _, flowSource := range flowSources {
		// The copy shares its nodes with the map, which remap changes.
		if err := ids.remap(&flowSource); err != nil {
			return err
		}
	}
	return nil
}

// missing returns the keys of the messages and variables of the bundle that
// the flows use and the app doesn't have, also through the buttons of those
// messages.
func (ids bundleIDs) missing(b *bundle.Bundle, flows []flow.FlowData) ([]string, error) {
	var keys []string
	add := func(key string, have map[string]string, inBundle bool) bool {
		if key == "" || have[key] != "" || !inBundle || slices.Contains(keys, key) {
			return false
		}
		keys = append(keys, key)
		return true
	}

	for i := 0; i < len(flows); i++ {
		for _, node := range flows[i].Nodes {
			key := node.Data.MessageTemplateID
			msg, ok := b.Message(key)
			if add(key, ids.messages, ok) {
				for _, raw := range msg.FlowSources {
					var flowSource flow.FlowData
					if err := json.Unmarshal(raw, &flowSource); err != nil {
						return nil, fmt.Errorf("failed to decode message flow: %w", err)
					}
					flows = append(flows, flowSource)
				}
			}
			key = node.Data.VariableID
			_, ok = b.Variable(key)
			add(key, ids.variables, ok)
		}
	}
	return keys, nil
}

// remap points the blocks that use a message or variable of the bundle at the
// one in the app. Blocks can't use others, they don't exist in the app.
func (ids bundleIDs) remap(flowSource *flow.FlowData) error {
	for i, node := range flowSource.Nodes {
		if key := node.Data.MessageTemplateID; key != "" {
			id, ok := ids.messages[key]
			if !ok {
				return handler.ErrBadRequest("unknown_message", fmt.Sprintf("block %s uses message %s of the template, which the app doesn't have", node.ID, key))
			}
			flowSource.Nodes[i].Data.MessageTemplateID = id
		}
		if key := node.Data.VariableID; key != "" {
			id, ok := ids.variables[key]
			if !ok {
				return handler.ErrBadRequest("unknown_variable", fmt.Sprintf("block %s uses variable %s of the template, which the app doesn't have", node.ID, key))
			}
			flowSource.Nodes[i].Data.VariableID = id
		}
	}
	return nil
}

func (ic *installedContents) toWire(install *model.BundleInstall) *wire.BundleInstall {
	res := wire.BundleInstallToWire(install)
	for i, item := range install.Items {
		hash, ok := ic.hash(install.ID, item)
		res.Items[i].Deleted = !ok
		res.Items[i].Edited = ok && hash != item.Hash
		if ok {
			res.Items[i].Parameters = ic.parameterValues(item)
		}
	}
	return res
}

// flows returns the flows of the item, the ones of a message's buttons for a
// message and none for a variable.
func (ic *installedContents) flows(item model.BundleInstallItem) []flow.FlowData {
	switch item.Type {
	case model.BundleInstallItemTypeCommand:
		return []flow.FlowData{ic.commands[item.ID].FlowSource}
	case model.BundleInstallItemTypeEventListener:
		return []flow.FlowData{ic.eventListeners[item.ID].FlowSource}
	case model.BundleInstallItemTypeMessage:
		return slices.Collect(maps.Values(ic.messages[item.ID].FlowSources))
	}
	return nil
}

// parameterValues returns the values the flows of the item have now.
func (ic *installedContents) parameterValues(item model.BundleInstallItem) map[string]string {
	values := map[string]string{}
	for _, f := range ic.flows(item) {
		maps.Copy(values, f.ParameterValues())
	}
	return values
}
