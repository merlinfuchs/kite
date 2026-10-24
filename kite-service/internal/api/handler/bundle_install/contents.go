package bundleinstall

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
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
	variables      map[string]*model.Variable
	items          map[string]installedItem
}

type installedItem struct {
	installID string
	itemType  model.BundleInstallItemType
	// hash is the hash of what the item is now.
	hash string
}

func (h *BundleInstallHandler) installedContents(c *handler.Context) (*installedContents, error) {
	commands, err := h.commandStore.CommandsByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get commands: %w", err)
	}
	eventListeners, err := h.eventListenerStore.EventListenersByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event listeners: %w", err)
	}
	messages, err := h.messageStore.MessagesByApp(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	variables, err := h.variableStore.VariablesByAppWithoutTotals(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variables: %w", err)
	}

	res := &installedContents{
		commands:       map[string]*model.Command{},
		eventListeners: map[string]*model.EventListener{},
		messages:       map[string]*model.Message{},
		variables:      map[string]*model.Variable{},
		items:          map[string]installedItem{},
	}
	if err := res.replace(store.BundleInstallContents{
		Commands:       commands,
		EventListeners: eventListeners,
		Messages:       messages,
		Variables:      variables,
	}); err != nil {
		return nil, err
	}
	return res, nil
}

// replace puts the contents in, skipping the ones no install created.
func (ic *installedContents) replace(contents store.BundleInstallContents) error {
	for _, command := range contents.Commands {
		if !command.BundleInstallID.Valid {
			continue
		}
		hash, err := command.FlowSource.ContentHash()
		if err != nil {
			return fmt.Errorf("failed to hash command flow: %w", err)
		}
		ic.commands[command.ID] = command
		ic.items[command.ID] = installedItem{command.BundleInstallID.String, model.BundleInstallItemTypeCommand, hash}
	}
	for _, listener := range contents.EventListeners {
		if !listener.BundleInstallID.Valid {
			continue
		}
		hash, err := listener.FlowSource.ContentHash()
		if err != nil {
			return fmt.Errorf("failed to hash event listener flow: %w", err)
		}
		ic.eventListeners[listener.ID] = listener
		ic.items[listener.ID] = installedItem{listener.BundleInstallID.String, model.BundleInstallItemTypeEventListener, hash}
	}
	for _, msg := range contents.Messages {
		if !msg.BundleInstallID.Valid {
			continue
		}
		hash, err := msg.ContentHash()
		if err != nil {
			return fmt.Errorf("failed to hash message: %w", err)
		}
		ic.messages[msg.ID] = msg
		ic.items[msg.ID] = installedItem{msg.BundleInstallID.String, model.BundleInstallItemTypeMessage, hash}
	}
	for _, variable := range contents.Variables {
		if !variable.BundleInstallID.Valid {
			continue
		}
		ic.variables[variable.ID] = variable
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

// bundleIDs returns the IDs of the messages and variables of the bundle the
// app has, from all installs of it.
func (ic *installedContents) bundleIDs(installs []*model.BundleInstall, bundleID string) bundleIDs {
	ids := bundleIDs{messages: map[string]string{}, variables: map[string]string{}}
	for _, install := range installs {
		if install.BundleID != bundleID {
			continue
		}
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
	}
	return ids
}

// bundleIDs are the IDs of messages and variables of a bundle in the app, by
// their key in the bundle. Blocks of the bundle use them by their key.
type bundleIDs struct {
	messages  map[string]string
	variables map[string]string
}

// remap points the blocks that use a message or variable of the bundle at the
// one in the app. Blocks can't use others, they don't exist in the app.
func (ids bundleIDs) remap(flowSource *flow.FlowData) error {
	for i, node := range flowSource.Nodes {
		if key := node.Data.MessageTemplateID; key != "" {
			id, ok := ids.messages[key]
			if !ok {
				return handler.ErrBadRequest("unknown_message", fmt.Sprintf("block %s uses message %s of the bundle, which the app doesn't have", node.ID, key))
			}
			flowSource.Nodes[i].Data.MessageTemplateID = id
		}
		if key := node.Data.VariableID; key != "" {
			id, ok := ids.variables[key]
			if !ok {
				return handler.ErrBadRequest("unknown_variable", fmt.Sprintf("block %s uses variable %s of the bundle, which the app doesn't have", node.ID, key))
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
	}
	return res
}
