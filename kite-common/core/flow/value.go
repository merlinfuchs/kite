package flow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-common/core/template"
)

type FlowValueType string

const (
	FlowValueTypeNull    FlowValueType = "null"
	FlowValueTypeString  FlowValueType = "string"
	FlowValueTypeNumber  FlowValueType = "number"
	FlowValueTypeArray   FlowValueType = "array"
	FlowValueTypeMessage FlowValueType = "message"
)

var FlowValueNull = FlowValue{Type: FlowValueTypeNull}

type FlowValue struct {
	Type  FlowValueType `json:"type"`
	Value interface{}   `json:"value"`
}

func (v *FlowValue) ContainsVariable() bool {
	if v.Type != FlowValueTypeString {
		return false
	}

	return template.ContainsVariable(v.String())
}

func (v *FlowValue) ResolveVariables(t FlowContextVariables) error {
	if !v.ContainsVariable() {
		return nil
	}

	res, err := t.ParseAndExecute(v.String())
	if err != nil {
		return fmt.Errorf("failed to resolve variables: %w", err)
	}

	*v = FlowValue{
		Type:  FlowValueTypeString,
		Value: res,
	}

	return nil
}

func (v *FlowValue) String() string {
	switch v.Type {
	case FlowValueTypeNull:
		return "null"
	case FlowValueTypeString:
		s, _ := v.Value.(string)
		return s
	case FlowValueTypeNumber:
		n, _ := v.Value.(float64)
		return fmt.Sprintf("%f", n)
	case FlowValueTypeArray:
		a, _ := v.Value.([]FlowValue)
		var res []string
		for _, v := range a {
			res = append(res, v.String())
		}
		return strings.Join(res, ", ")
	case FlowValueTypeMessage:
		m, _ := v.Message()
		return m.ID.String()
	}

	return ""
}

func (v *FlowValue) Number() float64 {
	switch v.Type {
	case FlowValueTypeNull:
		return 0
	case FlowValueTypeNumber:
		n, _ := v.Value.(float64)
		return n
	default:
		n, _ := strconv.ParseFloat(v.String(), 64)
		return n
	}
}

func (v *FlowValue) Message() (discord.Message, bool) {
	if v.Type != FlowValueTypeMessage {
		return discord.Message{}, false
	}

	return v.Value.(discord.Message), true
}

func (v *FlowValue) Equals(other *FlowValue) bool {
	return v.String() == other.String()
}

func (v *FlowValue) EqualsStrict(other *FlowValue) bool {
	// We can't just == the values directly, as they might contain pointers.
	return reflect.DeepEqual(v, other)
}

func (v *FlowValue) GreaterThan(other *FlowValue) bool {
	return v.Number() > other.Number()
}

func (v *FlowValue) GreaterThanOrEqual(other *FlowValue) bool {
	return v.Number() >= other.Number()
}

func (v *FlowValue) LessThan(other *FlowValue) bool {
	return v.Number() < other.Number()
}

func (v *FlowValue) LessThanOrEqual(other *FlowValue) bool {
	return v.Number() <= other.Number()
}

func (v *FlowValue) Contains(other *FlowValue) bool {
	return strings.Contains(v.String(), other.String())
}

func (v *FlowValue) UnmarshalJSON(data []byte) error {
	aux := struct {
		Type  FlowValueType
		Value json.RawMessage
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	v.Type = aux.Type

	switch v.Type {
	case FlowValueTypeNull:
		v.Value = nil
		return nil
	case FlowValueTypeString:
		v.Value = ""
	case FlowValueTypeNumber:
		v.Value = float64(0)
	case FlowValueTypeArray:
		v.Value = []FlowValue{}
	case FlowValueTypeMessage:
		v.Value = discord.Message{}
	}

	if err := json.Unmarshal(aux.Value, &v.Value); err != nil {
		return err
	}

	return nil
}
