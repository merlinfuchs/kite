package wire

import (
	"regexp"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

var variableNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type Variable struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Scoped      bool        `json:"scoped"`
	AppID       string      `json:"app_id"`
	ModuleID    null.String `json:"module_id"`
	TotalValues null.Int    `json:"total_values"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type VariableGetResponse = Variable

type VariableListResponse = []*Variable

type VariableCreateRequest struct {
	Name   string `json:"name"`
	Scoped bool   `json:"scoped"`
}

func (req VariableCreateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.Name,
			validation.Required,
			validation.Length(1, 100),
			validation.Match(variableNameRegex).
				Error("must only consist of letters, numbers, and underscores"),
		),
	)
}

type VariableCreateResponse = Variable

type VariablesImportRequest struct {
	Variables []VariableCreateRequest `json:"variables"`
}

func (req VariablesImportRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Variables, validation.Required),
	)
}

type VariablesImportResponse = []*Variable

type VariableUpdateRequest struct {
	Name   string `json:"name"`
	Scoped bool   `json:"scoped"`
}

func (req VariableUpdateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.Name,
			validation.Required,
			validation.Length(1, 100),
			validation.Match(variableNameRegex).
				Error("must only consist of letters, numbers, and underscores"),
		),
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
		Name:        variable.Name,
		Scoped:      variable.Scoped,
		AppID:       variable.AppID,
		ModuleID:    variable.ModuleID,
		CreatedAt:   variable.CreatedAt,
		UpdatedAt:   variable.UpdatedAt,
		TotalValues: variable.TotalValues,
	}
}
