package model

// DeletedEntityType is the table a deleted entity was deleted from.
type DeletedEntityType string

const (
	DeletedEntityTypeCommand        DeletedEntityType = "commands"
	DeletedEntityTypeEventListener  DeletedEntityType = "event_listeners"
	DeletedEntityTypePluginInstance DeletedEntityType = "plugin_instances"
)

// DeletedEntity is the tombstone of a deleted command, event listener or
// plugin instance.
type DeletedEntity struct {
	ID    string
	Type  DeletedEntityType
	AppID string
}
