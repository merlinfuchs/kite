package manifest

type ConfigSchema struct {
	Sections []ConfigSectionSchema `json:"sections"`
	Fields   []ConfigFieldSchema   `json:"fields"`
}

func (c ConfigSchema) DefaultConfig() map[string]interface{} {
	config := make(map[string]interface{})
	for _, field := range c.Fields {
		config[field.Key] = field.DefaultValue
	}
	for _, section := range c.Sections {
		for _, field := range section.Fields {
			config[field.Key] = field.DefaultValue
		}
	}
	return config
}

type ConfigSectionSchema struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Fields      []ConfigFieldSchema `json:"fields"`
}

type ConfigFieldType string

const (
	ConfigFieldTypeString ConfigFieldType = "STRING"
	ConfigFieldTypeInt    ConfigFieldType = "INT"
	ConfigFieldTypeFloat  ConfigFieldType = "FLOAT"
	ConfigFieldTypeBool   ConfigFieldType = "BOOL"
)

type ConfigFieldSchema struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Key          string          `json:"key"`
	Type         ConfigFieldType `json:"type"`
	Required     bool            `json:"required"`
	DefaultValue interface{}     `json:"default_value"`
}
