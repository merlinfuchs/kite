package wire

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

type Module struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Enabled         bool              `json:"enabled"`
	AppID           string            `json:"app_id"`
	TemplateID      string            `json:"template_id"`
	TemplateVersion int               `json:"template_version"`
	Settings        map[string]string `json:"settings"`
	Items           []ModuleItem      `json:"items"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// ModuleItem is a command or event listener the module created. It may have
// been deleted since.
type ModuleItem struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ModuleListResponse = []*Module

// ModuleCreateRequest installs a template: it creates the module together with
// its commands and event listeners.
type ModuleCreateRequest struct {
	TemplateID      string                             `json:"template_id"`
	TemplateVersion int                                `json:"template_version"`
	Name            string                             `json:"name"`
	Description     string                             `json:"description"`
	Settings        map[string]string                  `json:"settings"`
	Commands        []ModuleCommandCreateRequest       `json:"commands"`
	EventListeners  []ModuleEventListenerCreateRequest `json:"event_listeners"`
}

func (req ModuleCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.TemplateID, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.TemplateVersion, validation.Required, validation.Min(1)),
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Description, validation.Length(0, 1000)),
		validation.Field(&req.Settings, validation.By(validateModuleSettings)),
		validation.Field(&req.EventListeners),
		validation.Field(&req.Commands, validation.By(func(any) error {
			if len(req.Commands)+len(req.EventListeners) == 0 {
				return errors.New("must contain at least one command or event listener")
			}
			keys := make(map[string]bool)
			for _, c := range req.Commands {
				if keys[c.Key] {
					return fmt.Errorf("key %s is used twice", c.Key)
				}
				keys[c.Key] = true
			}
			for _, l := range req.EventListeners {
				if keys[l.Key] {
					return fmt.Errorf("key %s is used twice", l.Key)
				}
				keys[l.Key] = true
			}
			return nil
		})),
	)
}

func validateModuleSettings(value any) error {
	settings, _ := value.(map[string]string)
	if len(settings) > 50 {
		return errors.New("must have at most 50 settings")
	}
	for key, v := range settings {
		if len(key) == 0 || len(key) > 100 {
			return errors.New("keys must be between 1 and 100 characters")
		}
		if len(v) > 4000 {
			return fmt.Errorf("value of %s must be at most 4000 characters", key)
		}
	}
	return nil
}

type ModuleCommandCreateRequest struct {
	// Key identifies the command in the template.
	Key     string               `json:"key"`
	Command CommandCreateRequest `json:"command"`
}

func (req ModuleCommandCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Key, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Command),
	)
}

type ModuleEventListenerCreateRequest struct {
	// Key identifies the event listener in the template.
	Key           string                     `json:"key"`
	EventListener EventListenerCreateRequest `json:"event_listener"`
}

func (req ModuleEventListenerCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Key, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.EventListener),
	)
}

type ModuleCreateResponse = Module

func ModuleToWire(module *model.Module) *Module {
	if module == nil {
		return nil
	}

	items := make([]ModuleItem, len(module.Items))
	for i, item := range module.Items {
		items[i] = ModuleItem{
			Key:  item.Key,
			Type: string(item.Type),
			ID:   item.ID,
		}
	}

	return &Module{
		ID:              module.ID,
		Name:            module.Name,
		Description:     module.Description,
		Enabled:         module.Enabled,
		AppID:           module.AppID,
		TemplateID:      module.TemplateID,
		TemplateVersion: module.TemplateVersion,
		Settings:        module.Settings,
		Items:           items,
		CreatedAt:       module.CreatedAt,
		UpdatedAt:       module.UpdatedAt,
	}
}
