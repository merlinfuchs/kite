package thing

import (
	"encoding/json"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/stretchr/testify/assert"
)

func TestGuessType(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected Type
	}{
		{name: "string", value: "test", expected: TypeString},
		{name: "int", value: 1, expected: TypeInt},
		{name: "float", value: 1.1, expected: TypeFloat},
		{name: "bool", value: true, expected: TypeBool},
		{name: "array", value: []Thing{NewInt(1), NewInt(2), NewInt(3)}, expected: TypeArray},
		{name: "object", value: map[string]Thing{"a": NewInt(1), "b": NewInt(2)}, expected: TypeObject},
		{name: "discord_message", value: discord.Message{ID: 123}, expected: TypeDiscordMessage},
		{name: "discord_user", value: discord.User{ID: 123}, expected: TypeDiscordUser},
		{name: "discord_member", value: discord.Member{User: discord.User{ID: 123}}, expected: TypeDiscordMember},
		{name: "discord_channel", value: discord.Channel{ID: 123}, expected: TypeDiscordChannel},
		{name: "discord_guild", value: discord.Guild{ID: 123}, expected: TypeDiscordGuild},
		{name: "discord_role", value: discord.Role{ID: 123}, expected: TypeDiscordRole},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			thing := NewGuessTypeWithFallback(test.value)
			assert.Equal(t, test.expected, thing.Type)
		})
	}
}

func TestMarshalUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name       string
		value      any
		equalValue func(t *testing.T, old, new Thing)
	}{
		{name: "string", value: "test"},
		{name: "int", value: 1},
		{name: "float", value: 1.0},
		{name: "bool", value: true},
		{name: "array", value: []Thing{NewInt(1), NewInt(2), NewObject(map[string]Thing{"a": NewInt(3)})}},
		{name: "object", value: map[string]Thing{"a": NewInt(1), "b": NewObject(map[string]Thing{"c": NewInt(2)})}},
		{name: "discord_message", value: discord.Message{ID: 123}, equalValue: func(t *testing.T, old, new Thing) {
			assert.Equal(t, old.DiscordMessage().ID, new.DiscordMessage().ID)
		}},
		{name: "discord_user", value: discord.User{ID: 123}, equalValue: func(t *testing.T, old, new Thing) {
			assert.Equal(t, old.DiscordUser().ID, new.DiscordUser().ID)
		}},
		{name: "discord_member", value: discord.Member{User: discord.User{ID: 123}}, equalValue: func(t *testing.T, old, new Thing) {
			assert.Equal(t, old.DiscordMember().User.ID, new.DiscordMember().User.ID)
		}},
		{name: "discord_channel", value: discord.Channel{ID: 123}, equalValue: func(t *testing.T, old, new Thing) {
			assert.Equal(t, old.DiscordChannel().ID, new.DiscordChannel().ID)
		}},
		{name: "discord_guild", value: discord.Guild{ID: 123}, equalValue: func(t *testing.T, old, new Thing) {
			assert.Equal(t, old.DiscordGuild().ID, new.DiscordGuild().ID)
		}},
		{name: "discord_role", value: discord.Role{ID: 123}, equalValue: func(t *testing.T, old, new Thing) {
			assert.Equal(t, old.DiscordRole().ID, new.DiscordRole().ID)
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			thing := NewGuessTypeWithFallback(test.value)
			raw, err := json.Marshal(thing)
			if err != nil {
				t.Errorf("error marshalling thing: %v", err)
			}
			var unmarshalled Thing
			err = json.Unmarshal(raw, &unmarshalled)
			if err != nil {
				t.Errorf("error unmarshalling thing: %v", err)
			}

			assert.Equal(t, thing.Type, unmarshalled.Type)
			if test.equalValue != nil {
				test.equalValue(t, thing, unmarshalled)
			} else {
				assert.Equal(t, thing.Value, unmarshalled.Value)
			}
		})
	}

}

func TestUnmarshalFallback(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected Type
	}{
		{name: "string", value: "1", expected: TypeInt},
		{name: "string", value: "1.1", expected: TypeFloat},
		{name: "string", value: "true", expected: TypeBool},
		{name: "string", value: "false", expected: TypeBool},
		{name: "string", value: "[]", expected: TypeAny},
		{name: "string", value: "{}", expected: TypeAny},
		{name: "string", value: "123", expected: TypeInt},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var thing Thing
			err := json.Unmarshal([]byte(test.value), &thing)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, thing.Type)
		})
	}
}

func TestPartialGuessType(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected Thing
	}{
		{
			name:     "object_with_untyped_values",
			value:    `{"t": "object", "v": {"a": 1, "b": 2}}`,
			expected: NewObject(map[string]Thing{"a": NewInt(1), "b": NewInt(2)}),
		},
		{
			name:     "array_with_untyped_values",
			value:    `{"t": "array", "v": [1, 2, 3]}`,
			expected: NewArray([]Thing{NewInt(1), NewInt(2), NewInt(3)}),
		},
		{
			name:     "nested_object",
			value:    `{"t": "object", "v": {"a": {"b": 1}}}`,
			expected: NewObject(map[string]Thing{"a": NewAny(map[string]any{"b": float64(1)})}),
		},
		{
			name:     "string",
			value:    `{"t": "string", "v": "test"}`,
			expected: NewString("test"),
		},
		{name: "string", value: `{"t": "string", "v": "test"}`, expected: NewString("test")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var thing Thing
			err := json.Unmarshal([]byte(test.value), &thing)
			assert.NoError(t, err)
			assert.Equal(t, test.expected.Type, thing.Type)
			assert.Equal(t, test.expected.Value, thing.Value)
		})
	}
}

func TestIntConversions(t *testing.T) {
	tests := []struct {
		name     string
		value    Thing
		expected int64
	}{
		{name: "int", value: NewInt(1), expected: 1},
		{name: "valid_string", value: NewString("5"), expected: 5},
		{name: "invalid_string", value: NewString("test"), expected: 0},
		{name: "float", value: NewFloat(1.1), expected: 1},
		{name: "message", value: NewDiscordMessage(discord.Message{ID: 123}), expected: 123},
		{name: "bool", value: NewBool(true), expected: 1},
		{name: "bool", value: NewBool(false), expected: 0},
		{name: "array", value: NewArray([]Thing{NewInt(1), NewInt(2), NewInt(3)}), expected: 3},
		{name: "object", value: NewObject(map[string]Thing{"a": NewInt(1), "b": NewInt(2)}), expected: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.value.Int())
		})
	}
}

func TestFloatConversions(t *testing.T) {
	tests := []struct {
		name     string
		value    Thing
		expected float64
	}{
		{name: "int", value: NewInt(1), expected: 1},
		{name: "valid_string", value: NewString("5.5"), expected: 5.5},
		{name: "invalid_string", value: NewString("test"), expected: 0},
		{name: "float", value: NewFloat(1.1), expected: 1.1},
		{name: "message", value: NewDiscordMessage(discord.Message{ID: 123}), expected: 123},
		{name: "bool", value: NewBool(true), expected: 1},
		{name: "bool", value: NewBool(false), expected: 0},
		{name: "array", value: NewArray([]Thing{NewInt(1), NewInt(2), NewInt(3)}), expected: 3},
		{name: "object", value: NewObject(map[string]Thing{"a": NewInt(1), "b": NewInt(2)}), expected: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.value.Float())
		})
	}
}

func TestBoolConversions(t *testing.T) {
	tests := []struct {
		name     string
		value    Thing
		expected bool
	}{
		{name: "int", value: NewInt(1), expected: true},
		{name: "float", value: NewFloat(1.1), expected: true},
		{name: "string", value: NewString("true"), expected: true},
		{name: "string", value: NewString("false"), expected: false},
		{name: "bool", value: NewBool(true), expected: true},
		{name: "bool", value: NewBool(false), expected: false},
		{name: "array", value: NewArray([]Thing{NewInt(1), NewInt(2), NewInt(3)}), expected: true},
		{name: "object", value: NewObject(map[string]Thing{"a": NewInt(1), "b": NewInt(2)}), expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.value.Bool())
		})
	}
}

func TestStringConversions(t *testing.T) {
	tests := []struct {
		name     string
		value    Thing
		expected string
	}{
		{name: "int", value: NewInt(1), expected: "1"},
		{name: "float", value: NewFloat(1.1), expected: "1.1"},
		{name: "string", value: NewString("test"), expected: "test"},
		{name: "bool", value: NewBool(true), expected: "true"},
		{name: "bool", value: NewBool(false), expected: "false"},
		{name: "message", value: NewDiscordMessage(discord.Message{ID: 123}), expected: "123"},
		{name: "user", value: NewDiscordUser(discord.User{ID: 123}), expected: "123"},
		{name: "member", value: NewDiscordMember(discord.Member{User: discord.User{ID: 123}}), expected: "123"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, test.value.String())
		})
	}
}
