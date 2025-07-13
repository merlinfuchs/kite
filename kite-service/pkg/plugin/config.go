package plugin

import (
	"encoding/json"
)

type ConfigValues map[string]json.RawMessage

func NewConfigValues() ConfigValues {
	return make(ConfigValues)
}

func (c ConfigValues) GetString(key string) string {
	return UnmarshalConfigValue[string](c[key])
}

func (c ConfigValues) GetInt(key string) int {
	return UnmarshalConfigValue[int](c[key])
}

func (c ConfigValues) GetBool(key string) bool {
	return UnmarshalConfigValue[bool](c[key])
}

func (c ConfigValues) GetFloat(key string) float64 {
	return UnmarshalConfigValue[float64](c[key])
}

func (c ConfigValues) GetStringArray(key string) []string {
	return UnmarshalConfigValue[[]string](c[key])
}

func (c ConfigValues) GetIntArray(key string) []int {
	return UnmarshalConfigValue[[]int](c[key])
}

func (c ConfigValues) GetFloatArray(key string) []float64 {
	return UnmarshalConfigValue[[]float64](c[key])
}

func (c ConfigValues) GetBoolArray(key string) []bool {
	return UnmarshalConfigValue[[]bool](c[key])
}

func UnmarshalConfigValue[T any](v json.RawMessage) T {
	var t T
	_ = json.Unmarshal(v, &t)
	return t
}

type Config struct {
	Sections []ConfigSection `json:"sections"`
}

type ConfigSection struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Fields      []ConfigField `json:"fields"`
}

type ConfigField struct {
	Key         string          `json:"key"`
	Type        ConfigFieldType `json:"type"`
	ItemType    ConfigFieldType `json:"item_type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Required    bool            `json:"required"`
}

type ConfigFieldType string

const (
	ConfigFieldTypeString ConfigFieldType = "string"
	ConfigFieldTypeInt    ConfigFieldType = "int"
	ConfigFieldTypeBool   ConfigFieldType = "bool"
	ConfigFieldTypeFloat  ConfigFieldType = "float"
	ConfigFieldTypeArray  ConfigFieldType = "array"
)
