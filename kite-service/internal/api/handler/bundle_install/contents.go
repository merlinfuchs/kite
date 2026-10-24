package bundleinstall

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// installedContents are the commands, event listeners and messages of the app
// that installs created, by their ID.
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

	res := &installedContents{
		commands:       map[string]*model.Command{},
		eventListeners: map[string]*model.EventListener{},
		messages:       map[string]*model.Message{},
		items:          map[string]installedItem{},
	}
	if err := res.replace(store.BundleInstallContents{
		Commands:       commands,
		EventListeners: eventListeners,
		Messages:       messages,
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

// bundleMessageIDs are the IDs of the messages of the bundle the app has, by
// their key in the bundle, from all installs of it. Blocks of the bundle use
// them by their key.
func (ic *installedContents) bundleMessageIDs(installs []*model.BundleInstall, bundleID string) map[string]string {
	ids := map[string]string{}
	for _, install := range installs {
		if install.BundleID != bundleID {
			continue
		}
		for _, item := range install.Items {
			if item.Type != model.BundleInstallItemTypeMessage {
				continue
			}
			if _, ok := ic.hash(install.ID, item); ok {
				ids[item.Key] = item.ID
			}
		}
	}
	return ids
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
