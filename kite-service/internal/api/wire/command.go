package wire

import (
	"encoding/json"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

type Command struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Enabled        bool          `json:"enabled"`
	AppID          string        `json:"app_id"`
	ModuleID       null.String   `json:"module_id"`
	CreatorUserID  string        `json:"creator_user_id"`
	FlowSource     flow.FlowData `json:"flow_source"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	LastDeployedAt null.Time     `json:"last_deployed_at"`
}

type CommandGetResponse = Command

type CommandListResponse = []*Command

type CommandCreateRequest struct {
	FlowSource flow.FlowData `json:"flow_source"`
	Enabled    bool          `json:"enabled"`
}

func (req CommandCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.FlowSource, validation.Required),
	)
}

type CommandCreateResponse = Command

type CommandsImportRequest struct {
	Commands []CommandCreateRequest `json:"commands"`
}

func (req CommandsImportRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Commands, validation.Required),
	)
}

type CommandsImportResponse = []*Command

type CommandUpdateRequest struct {
	FlowSource flow.FlowData `json:"flow_source"`
	Enabled    bool          `json:"enabled"`
}

func (req CommandUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.FlowSource, validation.Required),
	)
}

type CommandUpdateResponse = Command

type CommandUpdateEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

func (req CommandUpdateEnabledRequest) Validate() error {
	return nil
}

type CommandUpdateEnabledResponse = Command

type CommandDeleteResponse = Empty

type CommandsDeployResponse struct {
	Deployed bool            `json:"deployed"`
	Error    json.RawMessage `json:"error,omitempty"`
}

func CommandToWire(command *model.Command) *Command {
	if command == nil {
		return nil
	}

	return &Command{
		ID:             command.ID,
		Name:           command.Name,
		Description:    command.Description,
		Enabled:        command.Enabled,
		AppID:          command.AppID,
		ModuleID:       command.ModuleID,
		CreatorUserID:  command.CreatorUserID,
		FlowSource:     command.FlowSource,
		CreatedAt:      command.CreatedAt,
		UpdatedAt:      command.UpdatedAt,
		LastDeployedAt: command.LastDeployedAt,
	}
}
