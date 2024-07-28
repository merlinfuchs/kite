package flow

import (
	"regexp"

	"github.com/diamondburned/arikawa/v3/api"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var commandNameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

type FlowData struct {
	Nodes []FlowNode `json:"nodes"`
	Edges []FlowEdge `json:"edges"`
}

func (d FlowData) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Nodes),
		validation.Field(&d.Edges),
	)
}

type FlowNodeType string

const (
	FlowNodeTypeEntryCommand         FlowNodeType = "entry_command"
	FlowNodeTypeEntryEvent           FlowNodeType = "entry_event"
	FlowNodeTypeActionResponseCreate FlowNodeType = "action_response_create"
	FlowNodeTypeActionMessageCreate  FlowNodeType = "action_message_create"
	FlowNodeTypeActionLog            FlowNodeType = "action_log"
	FlowNodeTypeConditionCompare     FlowNodeType = "condition_compare"
	FlowNodeTypeConditionItemCompare FlowNodeType = "condition_item_compare"
	FlowNodeTypeConditionItemElse    FlowNodeType = "condition_item_else"
	// TODO: FlowNodeTypeOptionCommandArgument?
	// TODO: FlowNodeTypeOptionCommandPermissions?
	FlowNodeTypeOptionCommandText       FlowNodeType = "option_command_text"
	FlowNodeTypeOptionCommandNumber     FlowNodeType = "option_command_number"
	FlowNodeTypeOptionCommandUser       FlowNodeType = "option_command_user"
	FlowNodeTypeOptionCommandChannel    FlowNodeType = "option_command_channel"
	FlowNodeTypeOptionCommandRole       FlowNodeType = "option_command_role"
	FlowNodeTypeOptionCommandAttachment FlowNodeType = "option_command_attachment"
)

type FlowNode struct {
	ID       string           `json:"id"`
	Type     FlowNodeType     `json:"type,omitempty"`
	Data     FlowNodeData     `json:"data" tstype:"FlowNodeData & StringIndexable"`
	Position FlowNodePosition `json:"position"`
}

func (n FlowNode) Validate() error {
	err := validation.ValidateStruct(&n,
		validation.Field(&n.ID, validation.Required),
		validation.Field(&n.Type, validation.Required),
		validation.Field(&n.Data, validation.Required),
	)
	if err != nil {
		return err
	}

	return n.Data.Validate(n.Type)
}

type FlowNodeData struct {
	CustomLabel      string              `json:"custom_label,omitempty"`
	Name             string              `json:"name,omitempty"`
	Description      string              `json:"description,omitempty"`
	EventType        string              `json:"event_type,omitempty"`
	MessageData      api.SendMessageData `json:"message_data,omitempty"`
	MessageEphemeral bool                `json:"message_ephemeral,omitempty"`
	LogLevel         LogLevel            `json:"log_level,omitempty"`
	LogMessage       string              `json:"log_message,omitempty"`

	ConditionBaseValue     FlowValue         `json:"condition_base_value,omitempty"`
	ConditionAllowMultiple bool              `json:"condition_allow_multiple,omitempty"`
	ConditionItemMode      ConditionItemType `json:"condition_item_mode,omitempty"`
	ConditionItemValue     FlowValue         `json:"condition_item_value,omitempty"`

	ResultVariableName string `json:"result_variable_name,omitempty"`
}

func (d FlowNodeData) Validate(nodeType FlowNodeType) error {
	// We currently only validate data for entry nodes, as for the other nodes it's less critical that they are valid.

	return validation.ValidateStruct(&d,
		// Command Entry
		validation.Field(&d.Name, validation.When(nodeType == FlowNodeTypeEntryCommand,
			validation.Required,
			validation.Length(1, 32),
			validation.Match(commandNameRe),
		)),
		validation.Field(&d.Description, validation.When(nodeType == FlowNodeTypeEntryCommand,
			validation.Required,
			validation.Length(1, 100),
		)),

		// Command Option
		// TODO: make all command options one node type?
		validation.Field(&d.Name, validation.When(nodeType == FlowNodeTypeOptionCommandText,
			validation.Required,
			validation.Length(1, 32),
		)),
		validation.Field(&d.Description, validation.When(nodeType == FlowNodeTypeOptionCommandText,
			validation.Required,
			validation.Length(1, 100),
		)),

		// Event Entry
		validation.Field(&d.EventType, validation.When(nodeType == FlowNodeTypeEntryEvent,
			validation.Required,
		)),
	)
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type ConditionItemType string

const (
	ConditionItemModeEqual              ConditionItemType = "equal"
	ConditionItemModeNotEqual           ConditionItemType = "not_equal"
	ConditionItemModeGreaterThan        ConditionItemType = "greater_than"
	ConditionItemModeGreaterThanOrEqual ConditionItemType = "greater_than_or_equal"
	ConditionItemModeLessThan           ConditionItemType = "less_than"
	ConditionItemModeLessThanOrEqual    ConditionItemType = "less_than_or_equal"
	ConditionItemModeContains           ConditionItemType = "contains"
)

type FlowNodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type FlowEdge struct {
	ID     string `json:"id"`
	Type   string `json:"type,omitempty"`
	Source string `json:"source"`
	Target string `json:"target"`
}

func (e FlowEdge) Validate() error {
	return validation.ValidateStruct(&e,
		validation.Field(&e.ID, validation.Required),
		validation.Field(&e.Source, validation.Required),
		validation.Field(&e.Target, validation.Required),
	)
}

func (d *FlowData) CommandName() string {
	for _, node := range d.Nodes {
		if node.Type == "entry_command" {
			return node.Data.Name
		}
	}
	return ""
}

func (d *FlowData) CommandDescription() string {
	for _, node := range d.Nodes {
		if node.Type == "entry_command" {
			return node.Data.Description
		}
	}
	return ""
}
