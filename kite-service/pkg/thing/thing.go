package thing

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

var Null = NewAny(nil)

type Type string

const (
	AnyTypeAny            Type = "any"
	AnyTypeString         Type = "string"
	AnyTypeInt            Type = "int"
	AnyTypeFloat          Type = "float"
	AnyTypeBool           Type = "bool"
	AnyTypeDiscordMessage Type = "discord_message"
	AnyTypeDiscordUser    Type = "discord_user"
	AnyTypeDiscordMember  Type = "discord_member"
	AnyTypeDiscordChannel Type = "discord_channel"
	AnyTypeDiscordGuild   Type = "discord_guild"
	AnyTypeDiscordRole    Type = "discord_role"
	AnyTypeHTTPResponse   Type = "http_response"
	AnyTypeArray          Type = "array"
	AnyTypeObject         Type = "object"
)

// Thing is a wrapper around any with some helper methods
type Thing struct {
	Type  Type `json:"t"`
	Value any  `json:"v"`
}

func (w *Thing) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type  Type            `json:"t"`
		Value json.RawMessage `json:"v"`
	}
	if err := json.Unmarshal(data, &aux); err == nil && aux.Type != "" {
		w.Type = aux.Type

		switch aux.Type {
		case AnyTypeAny:
			w.Value, err = UnmarshalValue[any](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeString:
			w.Value, err = UnmarshalValue[string](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeInt:
			w.Value, err = UnmarshalValue[int64](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeFloat:
			w.Value, err = UnmarshalValue[float64](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeBool:
			w.Value, err = UnmarshalValue[bool](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeDiscordMessage:
			w.Value, err = UnmarshalValue[discord.Message](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeDiscordUser:
			w.Value, err = UnmarshalValue[discord.User](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeDiscordMember:
			w.Value, err = UnmarshalValue[discord.Member](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeDiscordChannel:
			w.Value, err = UnmarshalValue[discord.Channel](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeDiscordGuild:
			w.Value, err = UnmarshalValue[discord.Guild](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeDiscordRole:
			w.Value, err = UnmarshalValue[discord.Role](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeHTTPResponse:
			w.Value, err = UnmarshalValue[HTTPResponseValue](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeArray:
			w.Value, err = UnmarshalValue[[]Thing](aux.Value)
			if err != nil {
				return err
			}
		case AnyTypeObject:
			w.Value, err = UnmarshalValue[map[string]Thing](aux.Value)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown thing type: %s", aux.Type)
		}

		return nil
	}

	// This is for backwards compatibility with values missing the type field
	fmt.Println("Falling back to AnyTypeAny")
	w.Type = AnyTypeAny
	return json.Unmarshal(data, &w.Value)
}

func UnmarshalValue[T any](data []byte) (T, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return v, err
	}
	return v, nil
}

func NewAny(v any) Thing {
	if a, ok := v.(Thing); ok {
		return a
	}

	return Thing{
		Type:  AnyTypeAny,
		Value: v,
	}
}

func NewGuessType(v any) (Thing, error) {
	if a, ok := v.(Thing); ok {
		return a, nil
	}

	switch v := v.(type) {
	case string:
		return NewString(v), nil
	case int:
		return NewInt(v), nil
	case int8:
		return NewInt(v), nil
	case int16:
		return NewInt(v), nil
	case int32:
		return NewInt(v), nil
	case int64:
		return NewInt(v), nil
	case uint:
		return NewInt(v), nil
	case uint8:
		return NewInt(v), nil
	case uint16:
		return NewInt(v), nil
	case uint32:
		return NewInt(v), nil
	case uint64:
		return NewInt(v), nil
	case float32:
		return NewFloat(v), nil
	case float64:
		return NewFloat(v), nil
	case bool:
		return NewBool(v), nil
	case []byte:
		return NewString(string(v)), nil
	case []Thing:
		return NewArray(v), nil
	case discord.Message:
		return NewDiscordMessage(v), nil
	case discord.User:
		return NewDiscordUser(v), nil
	case discord.Member:
		return NewDiscordMember(v), nil
	case discord.Channel:
		return NewDiscordChannel(v), nil
	case discord.Guild:
		return NewDiscordGuild(v), nil
	case discord.Role:
		return NewDiscordRole(v), nil
	case HTTPResponseValue:
		return NewHTTPResponse(v), nil
	case map[string]Thing:
		return NewObject(v), nil
	case nil:
		return Null, nil
	default:
		slog.Error("unable to guess type for value", "type", reflect.TypeOf(v))
		return Null, fmt.Errorf("unable to guess type for value: %T", v)
	}
}

func NewGuessTypeWithFallback(v any) Thing {
	res, err := NewGuessType(v)
	if err != nil {
		return NewAny(v)
	}
	return res
}

func NewString(v string) Thing {
	return Thing{
		Type:  AnyTypeString,
		Value: v,
	}
}

func NewInt[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](v T) Thing {
	return Thing{
		Type:  AnyTypeInt,
		Value: int64(v),
	}
}

func NewFloat[T float32 | float64](v T) Thing {
	return Thing{
		Type:  AnyTypeFloat,
		Value: float64(v),
	}
}

func NewBool(v bool) Thing {
	return Thing{
		Type:  AnyTypeBool,
		Value: v,
	}
}

func NewDiscordMessage(v discord.Message) Thing {
	return Thing{
		Type:  AnyTypeDiscordMessage,
		Value: v,
	}
}

func NewDiscordUser(v discord.User) Thing {
	return Thing{
		Type:  AnyTypeDiscordUser,
		Value: v,
	}
}

func NewDiscordChannel(v discord.Channel) Thing {
	return Thing{
		Type:  AnyTypeDiscordChannel,
		Value: v,
	}
}

func NewDiscordMember(v discord.Member) Thing {
	return Thing{
		Type:  AnyTypeDiscordMember,
		Value: v,
	}
}

func NewDiscordGuild(v discord.Guild) Thing {
	return Thing{
		Type:  AnyTypeDiscordGuild,
		Value: v,
	}
}

func NewDiscordRole(v discord.Role) Thing {
	return Thing{
		Type:  AnyTypeDiscordRole,
		Value: v,
	}
}

func NewHTTPResponse(v HTTPResponseValue) Thing {
	return Thing{
		Type:  AnyTypeHTTPResponse,
		Value: v,
	}
}

func NewFromHTTPResponse(v *http.Response) (Thing, error) {
	val, err := NewHTTPResponseValue(v)
	if err != nil {
		return Thing{}, fmt.Errorf("failed to create http response value: %w", err)
	}

	return NewHTTPResponse(val), nil
}

func NewArray(v []Thing) Thing {
	return Thing{
		Type:  AnyTypeArray,
		Value: v,
	}
}

func NewObject(v map[string]Thing) Thing {
	return Thing{
		Type:  AnyTypeObject,
		Value: v,
	}
}

func (w Thing) String() string {
	switch w.Type {
	case AnyTypeString:
		return w.Value.(string)
	case AnyTypeInt:
		return strconv.FormatInt(w.Value.(int64), 10)
	case AnyTypeFloat:
		return strconv.FormatFloat(w.Value.(float64), 'f', -1, 64)
	case AnyTypeBool:
		return strconv.FormatBool(w.Value.(bool))
	case AnyTypeDiscordMessage:
		return w.Value.(discord.Message).ID.String()
	case AnyTypeDiscordUser:
		return w.Value.(discord.User).ID.String()
	case AnyTypeDiscordMember:
		return w.Value.(discord.Member).User.ID.String()
	case AnyTypeDiscordChannel:
		return w.Value.(discord.Channel).ID.String()
	case AnyTypeDiscordGuild:
		return w.Value.(discord.Guild).ID.String()
	case AnyTypeDiscordRole:
		return w.Value.(discord.Role).ID.String()
	case AnyTypeHTTPResponse:
		return string(w.Value.(HTTPResponseValue).Body)
	case AnyTypeArray:
		return fmt.Sprintf("%v", w.Value)
	case AnyTypeObject:
		return fmt.Sprintf("%v", w.Value)
	default:
		return fmt.Sprintf("%v", w.Value)
	}
}

func (w Thing) Int() int64 {
	switch w.Type {
	case AnyTypeInt:
		return w.Value.(int64)
	case AnyTypeFloat:
		return int64(w.Float())
	case AnyTypeString:
		i, _ := strconv.ParseInt(w.Value.(string), 10, 64)
		return i
	case AnyTypeBool:
		if w.Value.(bool) {
			return 1
		}
		return 0
	case AnyTypeArray:
		return int64(len(w.Array()))
	case AnyTypeObject:
		return int64(len(w.Object()))
	case AnyTypeDiscordMessage:
		return int64(w.DiscordMessage().ID)
	case AnyTypeHTTPResponse:
		return int64(w.HTTPResponse().StatusCode)
	default:
		return 0
	}
}

func (w Thing) Float() float64 {
	switch w.Type {
	case AnyTypeInt:
		return float64(w.Value.(int64))
	case AnyTypeFloat:
		return w.Value.(float64)
	case AnyTypeString:
		f, _ := strconv.ParseFloat(w.Value.(string), 64)
		return f
	case AnyTypeBool:
		if w.Value.(bool) {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func (w Thing) Bool() bool {
	switch w.Type {
	case AnyTypeBool:
		return w.Value.(bool)
	case AnyTypeInt:
		return w.Value.(int64) != 0
	case AnyTypeFloat:
		return w.Value.(float64) != 0
	case AnyTypeString:
		return w.Value.(string) != ""
	case AnyTypeDiscordMessage:
		return w.Value != nil
	case AnyTypeHTTPResponse:
		return w.Value != nil
	default:
		return false
	}
}

func (w Thing) Object() map[string]Thing {
	switch w.Type {
	case AnyTypeObject:
		return w.Value.(map[string]Thing)
	default:
		return nil
	}
}

func (w Thing) Array() []Thing {
	switch w.Type {
	case AnyTypeArray:
		return w.Value.([]Thing)
	default:
		return nil
	}
}

func (w Thing) DiscordMessage() discord.Message {
	if w.Type == AnyTypeDiscordMessage {
		return w.Value.(discord.Message)
	}
	return discord.Message{}
}

func (w Thing) HTTPResponse() HTTPResponseValue {
	if w.Type == AnyTypeHTTPResponse {
		return w.Value.(HTTPResponseValue)
	}
	return HTTPResponseValue{}
}

func (w Thing) Equals(other *Thing) bool {
	return reflect.DeepEqual(w.Value, other.Value)
}

func (w Thing) GreaterThan(other *Thing) bool {
	return w.Float() > other.Float()
}

func (w Thing) GreaterThanOrEqual(other *Thing) bool {
	return w.Float() >= other.Float()
}

func (w Thing) LessThan(other *Thing) bool {
	return w.Float() < other.Float()
}

func (w Thing) LessThanOrEqual(other *Thing) bool {
	return w.Float() <= other.Float()
}

func (w Thing) Contains(other *Thing) bool {
	// TODO: handle arrays and objects?
	return strings.Contains(w.String(), other.String())
}

func (w Thing) IsEmpty() bool {
	return w.String() == ""
}

func (w Thing) IsNil() bool {
	return w.Value == nil
}

func (w Thing) Append(other Thing) Thing {
	// TODO: implement for arrays
	return NewString(w.String() + other.String())
}

func (w Thing) Add(other Thing) Thing {
	return NewFloat(w.Float() + other.Float())
}

func (w Thing) Sub(other Thing) Thing {
	return NewFloat(w.Float() - other.Float())
}

func Cast[T any](v Thing) (T, bool) {
	if t, ok := v.Value.(T); ok {
		return t, true
	}
	return *new(T), false
}
