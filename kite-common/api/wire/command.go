package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-common/core/flow"
	"github.com/kitecloud/kite/kite-common/model"
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

type CommandDeleteResponse = Empty

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
