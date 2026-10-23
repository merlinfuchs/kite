package wire

import (
	"errors"
	"fmt"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

type BundleInstall struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Enabled       bool                `json:"enabled"`
	AppID         string              `json:"app_id"`
	BundleID      string              `json:"bundle_id"`
	BundleVersion int                 `json:"bundle_version"`
	Settings      map[string]string   `json:"settings"`
	Items         []BundleInstallItem `json:"items"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// BundleInstallItem is a command or event listener the install created. It
// may have been deleted since.
type BundleInstallItem struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	ID   string `json:"id"`
}

type BundleInstallListResponse = []*BundleInstall

// BundleInstallCreateRequest installs a bundle: it creates the install
// together with its commands, event listeners and messages.
type BundleInstallCreateRequest struct {
	BundleID       string                                    `json:"bundle_id"`
	BundleVersion  int                                       `json:"bundle_version"`
	Name           string                                    `json:"name"`
	Description    string                                    `json:"description"`
	Settings       map[string]string                         `json:"settings"`
	Commands       []BundleInstallCommandCreateRequest       `json:"commands"`
	EventListeners []BundleInstallEventListenerCreateRequest `json:"event_listeners"`
	// Messages are created as message templates. Blocks that use one have its
	// key as their message_template_id, which is replaced with the ID of the
	// created template.
	Messages []BundleInstallMessageCreateRequest `json:"messages"`
}

func (req *BundleInstallCreateRequest) Sanitize() {
	for i := range req.Messages {
		req.Messages[i].Message.Sanitize()
	}
}

func (req BundleInstallCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.BundleID, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.BundleVersion, validation.Required, validation.Min(1)),
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Description, validation.Length(0, 1000)),
		validation.Field(&req.Settings, validation.By(validateBundleInstallSettings)),
		validation.Field(&req.EventListeners),
		validation.Field(&req.Messages),
		validation.Field(&req.Commands, validation.By(func(any) error {
			if len(req.Commands)+len(req.EventListeners)+len(req.Messages) == 0 {
				return errors.New("must contain at least one command, event listener or message")
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
			for _, m := range req.Messages {
				if keys[m.Key] {
					return fmt.Errorf("key %s is used twice", m.Key)
				}
				keys[m.Key] = true
			}
			return nil
		})),
	)
}

func validateBundleInstallSettings(value any) error {
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

type BundleInstallCommandCreateRequest struct {
	// Key identifies the command in the bundle.
	Key     string               `json:"key"`
	Command CommandCreateRequest `json:"command"`
}

func (req BundleInstallCommandCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Key, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Command),
	)
}

type BundleInstallEventListenerCreateRequest struct {
	// Key identifies the event listener in the bundle.
	Key           string                     `json:"key"`
	EventListener EventListenerCreateRequest `json:"event_listener"`
}

func (req BundleInstallEventListenerCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Key, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.EventListener),
	)
}

type BundleInstallMessageCreateRequest struct {
	// Key identifies the message in the bundle.
	Key     string               `json:"key"`
	Message MessageCreateRequest `json:"message"`
}

func (req BundleInstallMessageCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Key, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.Message),
	)
}

type BundleInstallCreateResponse = BundleInstall

func BundleInstallToWire(install *model.BundleInstall) *BundleInstall {
	if install == nil {
		return nil
	}

	items := make([]BundleInstallItem, len(install.Items))
	for i, item := range install.Items {
		items[i] = BundleInstallItem{
			Key:  item.Key,
			Type: string(item.Type),
			ID:   item.ID,
		}
	}

	return &BundleInstall{
		ID:            install.ID,
		Name:          install.Name,
		Description:   install.Description,
		Enabled:       install.Enabled,
		AppID:         install.AppID,
		BundleID:      install.BundleID,
		BundleVersion: install.BundleVersion,
		Settings:      install.Settings,
		Items:         items,
		CreatedAt:     install.CreatedAt,
		UpdatedAt:     install.UpdatedAt,
	}
}
