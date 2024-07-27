package flow

import (
	"github.com/diamondburned/arikawa/v3/api"
)

type FlowData struct {
	Nodes []FlowNode `json:"nodes"`
	Edges []FlowEdge `json:"edges"`
}

type FlowNodeType string

const (
	FlowNodeTypeEntryCommand            FlowNodeType = "entry_command"
	FlowNodeTypeEntryEvent              FlowNodeType = "entry_event"
	FlowNodeTypeActionResponseCreate    FlowNodeType = "action_response_create"
	FlowNodeTypeActionMessageCreate     FlowNodeType = "action_message_create"
	FlowNodeTypeActionLog               FlowNodeType = "action_log"
	FlowNodeTypeConditionCompare        FlowNodeType = "condition_compare"
	FlowNodeTypeConditionItemCompare    FlowNodeType = "condition_item_compare"
	FlowNodeTypeConditionItemElse       FlowNodeType = "condition_item_else"
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
