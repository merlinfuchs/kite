package wire

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
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

// MaxVariableValueLength is the maximum size of a value entered in the
// dashboard, the same as a value entered in a block.
const MaxVariableValueLength = 100_000

// MaxVariableScopeLength is the maximum length of a scope entered in the
// dashboard.
const MaxVariableScopeLength = 1000

// MaxVariableValuePreviewLength is how much of a value is returned in a list,
// flows can store values of several megabytes.
const MaxVariableValuePreviewLength = 10_000

// VariableValueType is how a value is shown and edited in the dashboard.
type VariableValueType string

const (
	VariableValueTypeString VariableValueType = "string"
	VariableValueTypeNumber VariableValueType = "number"
	VariableValueTypeBool   VariableValueType = "bool"
	// VariableValueTypeJSON covers lists, objects and null.
	VariableValueTypeJSON VariableValueType = "json"
)

type VariableValue struct {
	// Scope is null for the value of an unscoped variable.
	Scope null.String       `json:"scope"`
	Type  VariableValueType `json:"type"`
	// Value is the text for strings, numbers and booleans, and JSON otherwise.
	Value string `json:"value"`
	// ReadOnly values hold something that can't be entered as text, like a
	// Discord user stored by a flow. They can only be deleted.
	ReadOnly bool `json:"read_only"`
	// Truncated values were cut to MaxVariableValuePreviewLength.
	Truncated bool      `json:"truncated"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VariableValueListResponse struct {
	Values []*VariableValue `json:"values"`
	// Total is the number of values matching the search, not only this page.
	Total int `json:"total"`
}

// VariableValueSetRequest creates the value for a scope or overwrites it. An
// empty scope is the unscoped value.
type VariableValueSetRequest struct {
	Scope string            `json:"scope"`
	Type  VariableValueType `json:"type"`
	Value string            `json:"value"`
}

func (req *VariableValueSetRequest) Sanitize() {
	req.Scope = strings.TrimSpace(req.Scope)
}

func (req VariableValueSetRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Scope, validation.RuneLength(0, MaxVariableScopeLength)),
		validation.Field(&req.Type, validation.Required, validation.In(
			VariableValueTypeString,
			VariableValueTypeNumber,
			VariableValueTypeBool,
			VariableValueTypeJSON,
		)),
		validation.Field(&req.Value, validation.Length(0, MaxVariableValueLength)),
	)
}

// Thing parses the entered value. Invalid input is an error instead of
// becoming text, so a typo in a number doesn't silently change its type.
func (req VariableValueSetRequest) Thing() (thing.Thing, error) {
	switch req.Type {
	case VariableValueTypeString:
		return thing.NewString(req.Value), nil
	case VariableValueTypeNumber:
		raw := strings.TrimSpace(req.Value)
		if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
			return thing.NewInt(i), nil
		}
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return thing.Null, errors.New("must be a number")
		}
		return thing.NewFloat(f), nil
	case VariableValueTypeBool:
		switch strings.TrimSpace(req.Value) {
		case "true":
			return thing.NewBool(true), nil
		case "false":
			return thing.NewBool(false), nil
		}
		return thing.Null, errors.New("must be true or false")
	case VariableValueTypeJSON:
		// UseNumber so integers above 2^53, like Discord IDs, stay exact.
		dec := json.NewDecoder(strings.NewReader(req.Value))
		dec.UseNumber()

		var v any
		if err := dec.Decode(&v); err != nil {
			return thing.Null, errors.New("must be valid JSON")
		}
		if _, err := dec.Token(); err != io.EOF {
			return thing.Null, errors.New("must be a single JSON value")
		}
		return jsonToThing(v)
	}
	return thing.Null, errors.New("unknown type")
}

func jsonToThing(v any) (thing.Thing, error) {
	switch v := v.(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return thing.NewInt(i), nil
		}
		f, err := v.Float64()
		if err != nil {
			return thing.Null, errors.New("contains a number that is too large")
		}
		return thing.NewFloat(f), nil
	case map[string]any:
		res := make(map[string]thing.Thing, len(v))
		for key, item := range v {
			t, err := jsonToThing(item)
			if err != nil {
				return thing.Null, err
			}
			res[key] = t
		}
		return thing.NewObject(res), nil
	case []any:
		res := make([]thing.Thing, len(v))
		for i, item := range v {
			t, err := jsonToThing(item)
			if err != nil {
				return thing.Null, err
			}
			res[i] = t
		}
		return thing.NewArray(res), nil
	default:
		return thing.NewGuessTypeWithFallback(v), nil
	}
}

type VariableValueSetResponse = VariableValue

type VariableValueDeleteResponse = Empty

// isPlainThing reports whether the value survives being edited as text or
// JSON, which is not the case for Discord objects and the like.
func isPlainThing(t thing.Thing) bool {
	switch v := t.Value.(type) {
	case nil:
		return true
	case string:
		return t.Type == thing.TypeString
	case int64:
		return t.Type == thing.TypeInt
	case float64:
		return t.Type == thing.TypeFloat
	case bool:
		return t.Type == thing.TypeBool
	case []thing.Thing:
		if t.Type != thing.TypeArray {
			return false
		}
		for _, item := range v {
			if !isPlainThing(item) {
				return false
			}
		}
		return true
	case map[string]thing.Thing:
		if t.Type != thing.TypeObject {
			return false
		}
		for _, item := range v {
			if !isPlainThing(item) {
				return false
			}
		}
		return true
	}
	return false
}

func marshalIndent(v any) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	// The value is shown as text, not embedded in HTML.
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return ""
	}
	return strings.TrimSuffix(buf.String(), "\n")
}

func VariableValueToWire(value *model.VariableValue) *VariableValue {
	if value == nil {
		return nil
	}

	res := &VariableValue{
		Scope:     value.Scope,
		ReadOnly:  !isPlainThing(value.Data),
		CreatedAt: value.CreatedAt,
		UpdatedAt: value.UpdatedAt,
	}

	switch {
	case res.ReadOnly && value.Data.IsDiscordEntity():
		// The whole object, JSONValue would reduce it to its ID.
		res.Type = VariableValueTypeJSON
		res.Value = marshalIndent(value.Data.Value)
	case res.ReadOnly || value.Data.Value == nil:
		res.Type = VariableValueTypeJSON
		res.Value = marshalIndent(value.Data.JSONValue())
	case value.Data.Type == thing.TypeString:
		res.Type = VariableValueTypeString
		res.Value = value.Data.String()
	case value.Data.Type == thing.TypeInt || value.Data.Type == thing.TypeFloat:
		res.Type = VariableValueTypeNumber
		res.Value = value.Data.String()
	case value.Data.Type == thing.TypeBool:
		res.Type = VariableValueTypeBool
		res.Value = value.Data.String()
	default:
		res.Type = VariableValueTypeJSON
		res.Value = marshalIndent(value.Data.JSONValue())
	}

	if len(res.Value) > MaxVariableValuePreviewLength {
		// Cutting at a byte offset can split a character.
		res.Value = strings.ToValidUTF8(res.Value[:MaxVariableValuePreviewLength], "")
		res.Truncated = true
	}

	return res
}
