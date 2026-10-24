package wire

import (
	"encoding/json"
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
	Parameters    map[string]string   `json:"parameters"`
	Items         []BundleInstallItem `json:"items"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// BundleInstallItem is a command, event listener or message the install
// created.
type BundleInstallItem struct {
	Key  string `json:"key"`
	Type string `json:"type"`
	ID   string `json:"id"`
	// Edited is whether the user changed it since it was installed or updated.
	Edited bool `json:"edited"`
	// Deleted is whether the user deleted it. An update doesn't add it again.
	Deleted bool `json:"deleted"`
	// Parameters are the values its flows have now, which can differ from the
	// install's after a change in the editor.
	Parameters map[string]string `json:"parameters"`
}

type BundleInstallListResponse = []*BundleInstall

// BundleInstallCreateRequest installs items of a bundle. The service fills in
// the parameters, so an install is always the bundle at that version.
type BundleInstallCreateRequest struct {
	BundleID string `json:"bundle_id"`
	// BundleVersion is the version the user chose the items and parameters for.
	BundleVersion int               `json:"bundle_version"`
	Parameters    map[string]string `json:"parameters"`
	// Items are the keys of the commands, event listeners and messages to add.
	Items []string `json:"items"`
}

func (req BundleInstallCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.BundleID, validation.Required, validation.Length(1, 100)),
		validation.Field(&req.BundleVersion, validation.Required, validation.Min(1)),
		validation.Field(&req.Parameters, validation.By(validateBundleParameters)),
		validation.Field(&req.Items,
			validation.Required,
			validation.Length(1, 100),
			validation.Each(validation.Required, validation.Length(1, 100)),
			validation.By(func(any) error {
				keys := make(map[string]bool, len(req.Items))
				for _, key := range req.Items {
					if keys[key] {
						return fmt.Errorf("item %s is listed twice", key)
					}
					keys[key] = true
				}
				return nil
			}),
		),
	)
}

func validateBundleParameters(value any) error {
	params, _ := value.(map[string]string)
	if len(params) > 50 {
		return errors.New("must have at most 50 parameters")
	}
	for key, v := range params {
		if len(key) == 0 || len(key) > 100 {
			return errors.New("keys must be between 1 and 100 characters")
		}
		if len(v) > 4000 {
			return fmt.Errorf("value of %s must be at most 4000 characters", key)
		}
	}
	return nil
}

type BundleInstallCreateResponse = BundleInstall

// BundleInstallUpdateRequest updates an install to the current version of its
// bundle. Items the user deleted stay deleted, and edited ones are kept unless
// they're in Overwrite.
type BundleInstallUpdateRequest struct {
	// BundleVersion is the version the user saw the changes of.
	BundleVersion int `json:"bundle_version"`
	// Parameters are values for parameters the new version needs, the install
	// keeps the values of the others.
	Parameters map[string]string `json:"parameters"`
	// Overwrite are the keys of edited items to replace with the new version.
	Overwrite []string `json:"overwrite"`
}

func (req BundleInstallUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.BundleVersion, validation.Required, validation.Min(1)),
		validation.Field(&req.Parameters, validation.By(validateBundleParameters)),
		validation.Field(&req.Overwrite,
			validation.Length(0, 100),
			validation.Each(validation.Required, validation.Length(1, 100)),
		),
	)
}

type BundleInstallUpdateResponse = BundleInstall

type BundleInstallDeleteResponse = Empty

// BundleInstallParametersUpdateRequest sets parameters of an install in every
// flow it created that uses them, also where they were changed in the editor.
type BundleInstallParametersUpdateRequest struct {
	Parameters map[string]string `json:"parameters"`
}

func (req BundleInstallParametersUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Parameters, validation.Required, validation.By(validateBundleParameters)),
	)
}

type BundleInstallParametersUpdateResponse = BundleInstall

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
		Parameters:    install.Parameters,
		Items:         items,
		CreatedAt:     install.CreatedAt,
		UpdatedAt:     install.UpdatedAt,
	}
}

// BundleExportRequest exports items of the app as a bundle. The message
// templates and stored variables they use are exported with them.
type BundleExportRequest struct {
	CommandIDs       []string `json:"command_ids"`
	EventListenerIDs []string `json:"event_listener_ids"`
	MessageIDs       []string `json:"message_ids"`
}

func (req BundleExportRequest) Validate() error {
	ids := validation.Each(validation.Required, validation.Length(1, 100))
	return validation.ValidateStruct(&req,
		validation.Field(&req.CommandIDs, validation.Length(0, 100), ids),
		validation.Field(&req.EventListenerIDs, validation.Length(0, 100), ids),
		validation.Field(&req.MessageIDs, validation.Length(0, 100), ids),
	)
}

type BundleExportResponse struct {
	// Content is the bundle, whose parameters are the ones of its flows.
	Content json.RawMessage `json:"content"`
}

// BundleImportRequest imports items of an exported bundle with the parameters
// filled in. The items are copies, they aren't linked to the bundle.
type BundleImportRequest struct {
	Content    json.RawMessage   `json:"content"`
	Parameters map[string]string `json:"parameters"`
	// Items are the keys of the commands, event listeners, messages and
	// variables to add.
	Items []string `json:"items"`
}

func (req BundleImportRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Content, validation.Required, validation.Length(1, 512*1024)),
		validation.Field(&req.Parameters, validation.By(validateBundleParameters)),
		validation.Field(&req.Items,
			validation.Required,
			validation.Length(1, 400),
			validation.Each(validation.Required, validation.Length(1, 100)),
		),
	)
}

type BundleImportResponse struct {
	// Items are what the import created, keyed like the bundle. Messages and
	// variables the app already had aren't created again.
	Items []BundleInstallItem `json:"items"`
}
