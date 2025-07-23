package thing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

var Null = NewAny(nil)

type Type string

const (
	TypeAny            Type = "any"
	TypeString         Type = "string"
	TypeInt            Type = "int"
	TypeFloat          Type = "float"
	TypeBool           Type = "bool"
	TypeDiscordMessage Type = "discord_message"
	TypeDiscordUser    Type = "discord_user"
	TypeDiscordMember  Type = "discord_member"
	TypeDiscordChannel Type = "discord_channel"
	TypeDiscordGuild   Type = "discord_guild"
	TypeDiscordRole    Type = "discord_role"
	TypeRobloxUser     Type = "roblox_user"
	TypeHTTPResponse   Type = "http_response"
	TypeArray          Type = "array"
	TypeObject         Type = "object"
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
		case TypeAny:
			w.Value, err = UnmarshalValue[any](aux.Value)
			if err != nil {
				return err
			}
		case TypeString:
			w.Value, err = UnmarshalValue[string](aux.Value)
			if err != nil {
				return err
			}
		case TypeInt:
			w.Value, err = UnmarshalValue[int64](aux.Value)
			if err != nil {
				return err
			}
		case TypeFloat:
			w.Value, err = UnmarshalValue[float64](aux.Value)
			if err != nil {
				return err
			}
		case TypeBool:
			w.Value, err = UnmarshalValue[bool](aux.Value)
			if err != nil {
				return err
			}
		case TypeDiscordMessage:
			w.Value, err = UnmarshalValue[discord.Message](aux.Value)
			if err != nil {
				return err
			}
		case TypeDiscordUser:
			w.Value, err = UnmarshalValue[discord.User](aux.Value)
			if err != nil {
				return err
			}
		case TypeDiscordMember:
			w.Value, err = UnmarshalValue[discord.Member](aux.Value)
			if err != nil {
				return err
			}
		case TypeDiscordChannel:
			w.Value, err = UnmarshalValue[discord.Channel](aux.Value)
			if err != nil {
				return err
			}
		case TypeDiscordGuild:
			w.Value, err = UnmarshalValue[discord.Guild](aux.Value)
			if err != nil {
				return err
			}
		case TypeDiscordRole:
			w.Value, err = UnmarshalValue[discord.Role](aux.Value)
			if err != nil {
				return err
			}
		case TypeRobloxUser:
			w.Value, err = UnmarshalValue[RobloxUserValue](aux.Value)
			if err != nil {
				return err
			}
		case TypeHTTPResponse:
			w.Value, err = UnmarshalValue[HTTPResponseValue](aux.Value)
			if err != nil {
				return err
			}
		case TypeArray:
			w.Value, err = UnmarshalValue[[]Thing](aux.Value)
			if err != nil {
				return err
			}
		case TypeObject:
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
	// We try to guess the type from the value and fallback to TypeAny if we can't
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	t := NewGuessTypeWithFallback(value)

	w.Type = t.Type
	w.Value = t.Value
	return nil
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
		Type:  TypeAny,
		Value: v,
	}
}

func NewGuessType(v any) (Thing, error) {
	if a, ok := v.(Thing); ok {
		return a, nil
	}

	if t, ok := v.(ToThing); ok {
		return t.Thing(), nil
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
		if v == float32(int32(v)) {
			return NewInt(int32(v)), nil
		}
		return NewFloat(v), nil
	case float64:
		if v == float64(int64(v)) {
			return NewInt(int64(v)), nil
		}
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
	case RobloxUserValue:
		return NewRobloxUser(v), nil
	case HTTPResponseValue:
		return NewHTTPResponse(v), nil
	case map[string]Thing:
		return NewObject(v), nil
	case nil:
		return Null, nil
	default:
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
		Type:  TypeString,
		Value: v,
	}
}

func NewInt[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64](v T) Thing {
	return Thing{
		Type:  TypeInt,
		Value: int64(v),
	}
}

func NewFloat[T float32 | float64](v T) Thing {
	return Thing{
		Type:  TypeFloat,
		Value: float64(v),
	}
}

func NewBool(v bool) Thing {
	return Thing{
		Type:  TypeBool,
		Value: v,
	}
}

func NewDiscordMessage(v discord.Message) Thing {
	return Thing{
		Type:  TypeDiscordMessage,
		Value: v,
	}
}

func NewDiscordUser(v discord.User) Thing {
	return Thing{
		Type:  TypeDiscordUser,
		Value: v,
	}
}

func NewDiscordChannel(v discord.Channel) Thing {
	return Thing{
		Type:  TypeDiscordChannel,
		Value: v,
	}
}

func NewDiscordMember(v discord.Member) Thing {
	return Thing{
		Type:  TypeDiscordMember,
		Value: v,
	}
}

func NewDiscordGuild(v discord.Guild) Thing {
	return Thing{
		Type:  TypeDiscordGuild,
		Value: v,
	}
}

func NewDiscordRole(v discord.Role) Thing {
	return Thing{
		Type:  TypeDiscordRole,
		Value: v,
	}
}

func NewRobloxUser(v RobloxUserValue) Thing {
	return Thing{
		Type:  TypeRobloxUser,
		Value: v,
	}
}

func NewHTTPResponse(v HTTPResponseValue) Thing {
	return Thing{
		Type:  TypeHTTPResponse,
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
		Type:  TypeArray,
		Value: v,
	}
}

func NewObject(v map[string]Thing) Thing {
	return Thing{
		Type:  TypeObject,
		Value: v,
	}
}

func (w Thing) String() string {
	switch w.Type {
	case TypeString:
		return w.Value.(string)
	case TypeInt:
		return strconv.FormatInt(w.Value.(int64), 10)
	case TypeFloat:
		return strconv.FormatFloat(w.Value.(float64), 'f', -1, 64)
	case TypeBool:
		return strconv.FormatBool(w.Value.(bool))
	case TypeDiscordMessage:
		return w.Value.(discord.Message).Content
	case TypeDiscordUser:
		return w.Value.(discord.User).Mention()
	case TypeDiscordMember:
		return w.Value.(discord.Member).Mention()
	case TypeDiscordChannel:
		return w.Value.(discord.Channel).Mention()
	case TypeDiscordGuild:
		return w.Value.(discord.Guild).Name
	case TypeDiscordRole:
		return w.Value.(discord.Role).Mention()
	case TypeRobloxUser:
		return w.Value.(RobloxUserValue).Name
	case TypeHTTPResponse:
		return string(w.Value.(HTTPResponseValue).Body)
	case TypeArray:
		return fmt.Sprintf("%v", w.Value)
	case TypeObject:
		return fmt.Sprintf("%v", w.Value)
	default:
		return fmt.Sprintf("%v", w.Value)
	}
}

func (w Thing) Snowflake() discord.Snowflake {
	switch w.Type {
	case TypeString:
		id, _ := strconv.ParseInt(w.Value.(string), 10, 64)
		return discord.Snowflake(id)
	case TypeInt:
		return discord.Snowflake(w.Value.(int64))
	case TypeFloat:
		return discord.Snowflake(int64(w.Value.(float64)))
	case TypeBool:
		if w.Value.(bool) {
			return 1
		}
		return 0
	case TypeDiscordMessage:
		return discord.Snowflake(w.Value.(discord.Message).ID)
	case TypeDiscordUser:
		return discord.Snowflake(w.Value.(discord.User).ID)
	case TypeDiscordMember:
		return discord.Snowflake(w.Value.(discord.Member).User.ID)
	case TypeDiscordChannel:
		return discord.Snowflake(w.Value.(discord.Channel).ID)
	case TypeDiscordGuild:
		return discord.Snowflake(w.Value.(discord.Guild).ID)
	case TypeDiscordRole:
		return discord.Snowflake(w.Value.(discord.Role).ID)
	case TypeRobloxUser:
		return discord.Snowflake(w.Value.(RobloxUserValue).ID)
	case TypeHTTPResponse:
		return discord.Snowflake(int64(w.Value.(HTTPResponseValue).StatusCode))
	default:
		return discord.NullSnowflake
	}
}

func (w Thing) Int() int64 {
	switch w.Type {
	case TypeInt:
		return w.Value.(int64)
	case TypeFloat:
		return int64(w.Float())
	case TypeString:
		i, _ := strconv.ParseInt(w.Value.(string), 10, 64)
		return i
	case TypeBool:
		if w.Value.(bool) {
			return 1
		}
		return 0
	case TypeArray:
		return int64(len(w.Array()))
	case TypeObject:
		return int64(len(w.Object()))
	case TypeDiscordMessage:
		return int64(w.DiscordMessage().ID)
	case TypeDiscordUser:
		return int64(w.DiscordUser().ID)
	case TypeDiscordMember:
		return int64(w.DiscordMember().User.ID)
	case TypeDiscordChannel:
		return int64(w.DiscordChannel().ID)
	case TypeDiscordGuild:
		return int64(w.DiscordGuild().ID)
	case TypeDiscordRole:
		return int64(w.DiscordRole().ID)
	case TypeRobloxUser:
		return int64(w.RobloxUser().ID)
	case TypeHTTPResponse:
		return int64(w.HTTPResponse().StatusCode)
	default:
		str := w.String()
		if str == "" {
			return 0
		}
		i, _ := strconv.ParseInt(str, 10, 64)
		return i
	}
}

func (w Thing) Float() float64 {
	switch w.Type {
	case TypeInt:
		return float64(w.Value.(int64))
	case TypeFloat:
		return w.Value.(float64)
	case TypeString:
		f, _ := strconv.ParseFloat(w.Value.(string), 64)
		return f
	case TypeBool:
		if w.Value.(bool) {
			return 1
		}
		return 0
	case TypeArray:
		return float64(len(w.Array()))
	case TypeObject:
		return float64(len(w.Object()))
	case TypeDiscordMessage:
		return float64(w.DiscordMessage().ID)
	case TypeDiscordUser:
		return float64(w.DiscordUser().ID)
	case TypeDiscordMember:
		return float64(w.DiscordMember().User.ID)
	case TypeDiscordChannel:
		return float64(w.DiscordChannel().ID)
	case TypeDiscordGuild:
		return float64(w.DiscordGuild().ID)
	case TypeDiscordRole:
		return float64(w.DiscordRole().ID)
	case TypeRobloxUser:
		return float64(w.RobloxUser().ID)
	case TypeHTTPResponse:
		return float64(w.HTTPResponse().StatusCode)
	default:
		str := w.String()
		if str == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(str, 64)
		return f
	}
}

func (w Thing) Bool() bool {
	switch w.Type {
	case TypeBool:
		return w.Value.(bool)
	case TypeInt:
		return w.Value.(int64) != 0
	case TypeFloat:
		return w.Value.(float64) != 0
	case TypeString:
		v := w.Value.(string)
		return v != "" && v != "null" && v != "undefined" && v != "0" && v != "false"
	case TypeArray:
		return len(w.Array()) > 0
	case TypeObject:
		return len(w.Object()) > 0
	default:
		return w.Value != nil
	}
}

func (w Thing) Object() map[string]Thing {
	switch w.Type {
	case TypeObject:
		return w.Value.(map[string]Thing)
	default:
		return nil
	}
}

func (w Thing) Array() []Thing {
	switch w.Type {
	case TypeArray:
		return w.Value.([]Thing)
	default:
		return nil
	}
}

func (w Thing) DiscordMessage() discord.Message {
	if w.Type == TypeDiscordMessage {
		return w.Value.(discord.Message)
	}
	return discord.Message{}
}

func (w Thing) DiscordUser() discord.User {

	if w.Type == TypeDiscordUser {
		return w.Value.(discord.User)
	}
	return discord.User{}
}

func (w Thing) DiscordMember() discord.Member {

	if w.Type == TypeDiscordMember {
		return w.Value.(discord.Member)
	}
	return discord.Member{}
}

func (w Thing) DiscordChannel() discord.Channel {

	if w.Type == TypeDiscordChannel {
		return w.Value.(discord.Channel)
	}
	return discord.Channel{}
}

func (w Thing) DiscordGuild() discord.Guild {

	if w.Type == TypeDiscordGuild {
		return w.Value.(discord.Guild)
	}
	return discord.Guild{}
}

func (w Thing) DiscordRole() discord.Role {

	if w.Type == TypeDiscordRole {
		return w.Value.(discord.Role)
	}
	return discord.Role{}
}

func (w Thing) RobloxUser() RobloxUserValue {
	if w.Type == TypeRobloxUser {
		return w.Value.(RobloxUserValue)
	}
	return RobloxUserValue{}
}

func (w Thing) HTTPResponse() HTTPResponseValue {
	if w.Type == TypeHTTPResponse {
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
	return w.String() == "" || w.String() == "0"
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

type ToThing interface {
	Thing() Thing
}
