package flow

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
)

type FlowValueType string

const (
	FlowValueTypeNull         FlowValueType = "null"
	FlowValueTypeString       FlowValueType = "string"
	FlowValueTypeBool         FlowValueType = "bool"
	FlowValueTypeFloat        FlowValueType = "number"
	FlowValueTypeInt          FlowValueType = "int"
	FlowValueTypeArray        FlowValueType = "array"
	FlowValueTypeMessage      FlowValueType = "message"
	FlowValueTypeHTTPResponse FlowValueType = "http_response"
	FlowValueTypeAny          FlowValueType = "any"
)

var FlowValueNull = FlowValue{Type: FlowValueTypeNull}

type FlowValue struct {
	Type  FlowValueType `json:"type"`
	Value interface{}   `json:"value"`
}

func NewFlowValue(v interface{}) FlowValue {
	switch v := v.(type) {
	case string:
		return NewFlowValueString(v)
	case bool:
		return NewFlowValueBool(v)
	case float64:
		return NewFlowValueNumber(v)
	case float32:
		return NewFlowValueNumber(float64(v))
	case int64:
		return NewFlowValueInt(v)
	case int32:
		return NewFlowValueInt(int64(v))
	case int16:
		return NewFlowValueInt(int64(v))
	case int8:
		return NewFlowValueInt(int64(v))
	case uint64:
		return NewFlowValueInt(int64(v))
	case uint32:
		return NewFlowValueInt(int64(v))
	case uint16:
		return NewFlowValueInt(int64(v))
	case uint8:
		return NewFlowValueInt(int64(v))
	case int:
		return NewFlowValueInt(int64(v))
	case []FlowValue:
		return NewFlowValueArray(v)
	case discord.Message:
		return NewFlowValueMessage(v)
	case http.Response:
		return NewFlowValueHTTPResponse(v)
	default:
		return NewFlowValueAny(v)
	}
}

func NewFlowValueAny(v any) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeAny,
		Value: v,
	}
}

func NewFlowValueString(s string) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeString,
		Value: s,
	}
}

func NewFlowValueBool(b bool) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeBool,
		Value: b,
	}
}

func NewFlowValueNumber(n float64) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeFloat,
		Value: n,
	}
}

func NewFlowValueInt(i int64) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeInt,
		Value: i,
	}
}

func NewFlowValueArray(a []FlowValue) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeArray,
		Value: a,
	}
}

func NewFlowValueMessage(m discord.Message) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeMessage,
		Value: m,
	}
}

func NewFlowValueHTTPResponse(r http.Response) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeHTTPResponse,
		Value: r,
	}
}

func (v *FlowValue) IsNull() bool {
	return v.Type == FlowValueTypeNull || v.Type == "" || v.Value == nil
}

func (v FlowValue) String() string {
	switch v.Type {
	case FlowValueTypeNull:
		return "null"
	case FlowValueTypeString:
		s, _ := v.Value.(string)
		return s
	case FlowValueTypeBool:
		b, _ := v.Value.(bool)
		return strconv.FormatBool(b)
	case FlowValueTypeInt:
		i, _ := v.Value.(int64)
		return fmt.Sprintf("%d", i)
	case FlowValueTypeFloat:
		n, _ := v.Value.(float64)
		if n == math.Floor(n) {
			return fmt.Sprintf("%d", int64(n))
		}

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
	case FlowValueTypeHTTPResponse:
		res, _ := v.HTTPResponse()
		return res.Status
	default:
		return fmt.Sprintf("%v", v.Value)
	}
}

func (v FlowValue) Float() float64 {
	switch v.Type {
	case FlowValueTypeNull:
		return 0
	case FlowValueTypeBool:
		b, _ := v.Value.(bool)
		if b {
			return 1
		}
		return 0
	case FlowValueTypeFloat:
		n, _ := v.Value.(float64)
		return n
	case FlowValueTypeInt:
		i, _ := v.Value.(int64)
		return float64(i)
	default:
		n, _ := strconv.ParseFloat(v.String(), 64)
		return n
	}
}

func (v FlowValue) Int() int64 {
	switch v.Type {
	case FlowValueTypeNull:
		return 0
	case FlowValueTypeBool:
		b, _ := v.Value.(bool)
		if b {
			return 1
		}
		return 0
	case FlowValueTypeInt:
		i, _ := v.Value.(int64)
		return i
	case FlowValueTypeFloat:
		n, _ := v.Value.(float64)
		return int64(n)
	default:
		n, _ := strconv.ParseInt(v.String(), 10, 64)
		return n
	}
}

func (v FlowValue) Message() (discord.Message, bool) {
	if v.Type != FlowValueTypeMessage {
		return discord.Message{}, false
	}

	return v.Value.(discord.Message), true
}

func (v FlowValue) HTTPResponse() (http.Response, bool) {
	if v.Type != FlowValueTypeHTTPResponse {
		return http.Response{}, false
	}

	return v.Value.(http.Response), true
}

func (v FlowValue) Equals(other *FlowValue) bool {
	return v.String() == other.String()
}

func (v FlowValue) EqualsStrict(other *FlowValue) bool {
	// We can't just == the values directly, as they might contain pointers.
	return reflect.DeepEqual(v, other)
}

func (v FlowValue) GreaterThan(other *FlowValue) bool {
	return v.Float() > other.Float()
}

func (v FlowValue) GreaterThanOrEqual(other *FlowValue) bool {
	return v.Float() >= other.Float()
}

func (v FlowValue) LessThan(other *FlowValue) bool {
	return v.Float() < other.Float()
}

func (v FlowValue) LessThanOrEqual(other *FlowValue) bool {
	return v.Float() <= other.Float()
}

func (v FlowValue) Contains(other *FlowValue) bool {
	return strings.Contains(v.String(), other.String())
}

func (v FlowValue) IsEmpty() bool {
	return v.String() == ""
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
	case FlowValueTypeBool:
		v.Value = false
	case FlowValueTypeInt:
		v.Value = int64(0)
	case FlowValueTypeFloat:
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

func (v FlowValue) ResolvePlaceholder(ctx context.Context) (string, error) {
	return v.String(), nil
}

func (v FlowValue) EvalEnv() any {
	switch v.Type {
	case FlowValueTypeHTTPResponse:
		resp, _ := v.HTTPResponse()
		return eval.NewHTTPResponseEnv(&resp)
	case FlowValueTypeMessage:
		msg, _ := v.Message()
		return eval.NewMessageEnv(&msg)
	default:
		return v.Value
	}
}

// FlowString is a string that can contain placeholders
// and may produce a different type of value when evaluated.
type FlowString string

func (v *FlowString) EvalTemplate(ctx context.Context, t eval.Env) (FlowValue, error) {
	res, err := eval.EvalTemplate(ctx, v.String(), t)
	if err != nil {
		return FlowValueNull, fmt.Errorf("failed to eval template: %w", err)
	}

	return NewFlowValue(res), nil
}

func (v FlowString) String() string {
	return string(v)
}

func (v FlowString) StringValue() FlowValue {
	return NewFlowValueString(string(v))
}
