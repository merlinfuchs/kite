package wire

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"gopkg.in/guregu/null.v4"
)

// FlowFunction is a function of an app. It isn't named Function, which would
// shadow the built-in type of the same name in TypeScript.
type FlowFunction struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	AppID         string        `json:"app_id"`
	ModuleID      null.String   `json:"module_id"`
	CreatorUserID string        `json:"creator_user_id"`
	FlowSource    flow.FlowData `json:"flow_source"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type FunctionGetResponse = FlowFunction

type FunctionListResponse = []*FlowFunction

type FunctionCreateRequest struct {
	FlowSource flow.FlowData `json:"flow_source"`
}

func (req FunctionCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.FlowSource, validation.Required),
	)
}

type FunctionCreateResponse = FlowFunction

// FunctionsImportRequest creates several functions at once, e.g. the ones a
// shared command uses.
type FunctionsImportRequest struct {
	Functions []FunctionCreateRequest `json:"functions"`
}

func (req FunctionsImportRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Functions, validation.Required, validation.Length(1, 100)),
	)
}

type FunctionsImportResponse = []*FlowFunction

type FunctionUpdateRequest struct {
	FlowSource flow.FlowData `json:"flow_source"`
}

func (req FunctionUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.FlowSource, validation.Required),
	)
}

type FunctionUpdateResponse = FlowFunction

type FunctionDeleteResponse = Empty

func FunctionToWire(function *model.Function) *FlowFunction {
	if function == nil {
		return nil
	}

	return &FlowFunction{
		ID:            function.ID,
		Name:          function.Name,
		Description:   function.Description,
		AppID:         function.AppID,
		ModuleID:      function.ModuleID,
		CreatorUserID: function.CreatorUserID,
		FlowSource:    function.FlowSource,
		CreatedAt:     function.CreatedAt,
		UpdatedAt:     function.UpdatedAt,
	}
}
