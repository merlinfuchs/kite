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
	"github.com/kitecloud/kite/kite-service/pkg/placeholder"
)

type FlowValueType string

const (
	FlowValueTypeNull         FlowValueType = "null"
	FlowValueTypeString       FlowValueType = "string"
	FlowValueTypeNumber       FlowValueType = "number"
	FlowValueTypeArray        FlowValueType = "array"
	FlowValueTypeMessage      FlowValueType = "message"
	FlowValueTypeHTTPResponse FlowValueType = "http_response"
)

var FlowValueNull = FlowValue{Type: FlowValueTypeNull}

// TODO: do we need this or can we just have all values be strings?
type FlowValue struct {
	Type  FlowValueType `json:"type"`
	Value interface{}   `json:"value"`
}

func NewFlowValueString(s string) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeString,
		Value: s,
	}
}

func NewFlowValueNumber(n float64) FlowValue {
	return FlowValue{
		Type:  FlowValueTypeNumber,
		Value: n,
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

func (v *FlowValue) ContainsPlaceholder() bool {
	if v.Type != FlowValueTypeString {
		return false
	}

	return placeholder.ContainsPlaceholder(v.String())
}

func (v *FlowValue) FillPlaceholders(ctx context.Context, t *placeholder.Engine) error {
	res, err := t.Fill(ctx, v.String())
	if err != nil {
		return fmt.Errorf("failed to fill placeholders: %w", err)
	}

	*v = FlowValue{
		Type:  FlowValueTypeString,
		Value: res,
	}

	return nil
}

func (v *FlowValue) IsNull() bool {
	return v.Type == FlowValueTypeNull || v.Type == "" || v.Value == nil
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

func (v *FlowValue) HTTPResponse() (http.Response, bool) {
	if v.Type != FlowValueTypeMessage {
		return http.Response{}, false
	}

	return v.Value.(http.Response), true
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

func (v FlowValue) GetPlaceholder(ctx context.Context, key string) (placeholder.Provider, error) {
	// TODO: implement for some types
	return nil, placeholder.ErrNotFound
}

func (v FlowValue) ResolvePlaceholder(ctx context.Context) (string, error) {
	return v.String(), nil
}

// FlowString is a string that can contain placeholders.
type FlowString string

func (v FlowString) ContainsPlaceholder() bool {
	return placeholder.ContainsPlaceholder(string(v))
}

func (v *FlowString) FillPlaceholders(ctx context.Context, t *placeholder.Engine) (FlowString, error) {
	res, err := t.Fill(ctx, v.String())
	if err != nil {
		return "", fmt.Errorf("failed to fill placeholders: %w", err)
	}

	return FlowString(res), nil
}

func (v FlowString) Float() float64 {
	n, _ := strconv.ParseFloat(string(v), 64)
	return n
}

func (v FlowString) Int() int64 {
	n, _ := strconv.ParseInt(string(v), 10, 64)
	return n
}

func (v FlowString) String() string {
	return string(v)
}

func (v FlowString) Equals(other *FlowString) bool {
	return v.String() == other.String()
}

func (v FlowString) EqualsStrict(other *FlowString) bool {
	// We can't just == the values directly, as they might contain pointers.
	return reflect.DeepEqual(v, other)
}

func (v FlowString) GreaterThan(other *FlowString) bool {
	return v.Float() > other.Float()
}

func (v FlowString) GreaterThanOrEqual(other *FlowString) bool {
	return v.Float() >= other.Float()
}

func (v FlowString) LessThan(other *FlowString) bool {
	return v.Float() < other.Float()
}

func (v FlowString) LessThanOrEqual(other *FlowString) bool {
	return v.Float() <= other.Float()
}

func (v FlowString) Contains(other *FlowString) bool {
	return strings.Contains(v.String(), other.String())
}
