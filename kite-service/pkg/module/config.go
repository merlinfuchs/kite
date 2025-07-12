package module

type ConfigValues map[string]interface{}

func NewConfigValues() ConfigValues {
	return make(ConfigValues)
}

func (c ConfigValues) GetString(key string) string {
	v, _ := c[key].(string)
	return v
}

func (c ConfigValues) GetInt(key string) int {
	v, _ := c[key].(int)
	return v
}

func (c ConfigValues) GetBool(key string) bool {
	v, _ := c[key].(bool)
	return v
}

func (c ConfigValues) GetFloat(key string) float64 {
	v, _ := c[key].(float64)
	return v
}

func (c ConfigValues) GetStringArray(key string) []string {
	v, _ := c[key].([]string)
	return v
}

func (c ConfigValues) GetIntArray(key string) []int {
	v, _ := c[key].([]int)
	return v
}

func (c ConfigValues) GetFloatArray(key string) []float64 {
	v, _ := c[key].([]float64)
	return v
}

func (c ConfigValues) GetBoolArray(key string) []bool {
	v, _ := c[key].([]bool)
	return v
}

type Config struct {
	Sections []ConfigSection
}

type ConfigSection struct {
	Name        string
	Description string
	Fields      []ConfigField
}

type ConfigField struct {
	Key         string
	Type        ConfigFieldType
	ItemType    ConfigFieldType
	Name        string
	Description string
}

type ConfigFieldType string

const (
	ConfigFieldTypeString ConfigFieldType = "string"
	ConfigFieldTypeInt    ConfigFieldType = "int"
	ConfigFieldTypeBool   ConfigFieldType = "bool"
	ConfigFieldTypeFloat  ConfigFieldType = "float"
	ConfigFieldTypeArray  ConfigFieldType = "array"
)
