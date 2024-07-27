package flow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/diamondburned/arikawa/v3/discord"
)

type FlowValueType string

const (
	FlowValueTypeNull    FlowValueType = "null"
	FlowValueTypeString  FlowValueType = "string"
	FlowValueTypeNumber  FlowValueType = "number"
	FlowValueTypeMessage FlowValueType = "message"
)

var FlowValueNull = FlowValue{Type: FlowValueTypeNull}

type FlowValue struct {
	Type  FlowValueType
	Value interface{}
}

func (v *FlowValue) HasVariables() bool {
	return strings.Contains(v.String(), "{{")
}

func (v *FlowValue) ResolveVariables() {
	if !v.HasVariables() {
		return
	}

	// TODO: find and replace any variables in the value
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
	case FlowValueTypeMessage:
		m, _ := v.Value.(discord.Message)
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
	if len(data) == 0 {
		v.Type = FlowValueTypeNull
		return nil
	}

	if data[0] == '"' {
		v.Type = FlowValueTypeString
		v.Value = ""
		return json.Unmarshal(data, &v.Value)
	}

	if unicode.IsDigit(rune(data[0])) {
		v.Type = FlowValueTypeNumber
		v.Value = float64(0)
		return json.Unmarshal(data, &v.Value)
	}

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
	case FlowValueTypeMessage:
		v.Value = discord.Message{}
	}

	if err := json.Unmarshal(aux.Value, &v.Value); err != nil {
		return err
	}

	return nil
}

func (v FlowValue) MarshalJSON() ([]byte, error) {
	if v.Type == FlowValueTypeNull {
		return []byte("null"), nil
	}
	if v.Type == FlowValueTypeString {
		return json.Marshal(v.Value)
	}
	if v.Type == FlowValueTypeNumber {
		return json.Marshal(v.Value)
	}

	type Aux FlowValue
	return json.Marshal(Aux(v))
}
