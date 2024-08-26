package wire

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

var variableScopes = []interface{}{"global", "guild", "channel", "user", "member"}
var variableTypes = []interface{}{"string", "integer", "float", "boolean"}
var variableNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type Variable struct {
	ID          string      `json:"id"`
	Scope       string      `json:"scope"`
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	AppID       string      `json:"app_id"`
	ModuleID    null.String `json:"module_id"`
	TotalValues null.Int    `json:"total_values"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type VariableGetResponse = Variable

type VariableListResponse = []*Variable

type VariableCreateRequest struct {
	Scope string `json:"scope"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

func (req VariableCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Scope, validation.Required, validation.In(variableScopes...)),
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100), validation.Match(variableNameRegex)),
		validation.Field(&req.Type, validation.Required, validation.In(variableTypes...)),
	)
}

type VariableCreateResponse = Variable

type VariableUpdateRequest struct {
	Scope string `json:"scope"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

func (req VariableUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Scope, validation.Required, validation.In(variableScopes...)),
		validation.Field(&req.Name, validation.Required, validation.Length(1, 100), validation.Match(variableNameRegex)),
		validation.Field(&req.Type, validation.Required, validation.In(variableTypes...)),
	)
}

type VariableUpdateResponse = Variable

type VariableDeleteResponse = Empty

func VariableToWire(variable *model.Variable) *Variable {
	if variable == nil {
		return nil
	}

	return &Variable{
		ID:          variable.ID,
		Scope:       string(variable.Scope),
		Name:        variable.Name,
		Type:        variable.Type,
		AppID:       variable.AppID,
		ModuleID:    variable.ModuleID,
		CreatedAt:   variable.CreatedAt,
		UpdatedAt:   variable.UpdatedAt,
		TotalValues: variable.TotalValues,
	}
}
