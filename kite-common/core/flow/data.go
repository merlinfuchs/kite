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
	FlowNodeTypeEntryCommand             FlowNodeType = "entry_command"
	FlowNodeTypeEntryEvent               FlowNodeType = "entry_event"
	FlowNodeTypeActionResponseCreate     FlowNodeType = "action_response_create"
	FlowNodeTypeActionMessageCreate      FlowNodeType = "action_message_create"
	FlowNodeTypeActionLog                FlowNodeType = "action_log"
	FlowNodeTypeConditionCompare         FlowNodeType = "condition_compare"
	FlowNodeTypeConditionItemCompare     FlowNodeType = "condition_item_compare"
	FlowNodeTypeConditionItemElse        FlowNodeType = "condition_item_else"
	FlowNodeTypeOptionCommandArgument    FlowNodeType = "option_command_argument"
	FlowNodeTypeOptionCommandPermissions FlowNodeType = "option_command_permissions"
	FlowNodeTypeOptionEventFilter        FlowNodeType = "option_event_filter"
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
	// Shared
	Name               string `json:"name,omitempty"`
	Description        string `json:"description,omitempty"`
	CustomLabel        string `json:"custom_label,omitempty"`
	ResultVariableName string `json:"result_variable_name,omitempty"`

	// Command Argument
	CommandArgumentType     CommandArgumentType `json:"command_argument_type,omitempty"`
	CommandArgumentRequired bool                `json:"command_argument_required,omitempty"`

	// Command Permissions
	CommandPermissions string `json:"command_permissions,omitempty"`

	// Message Create & Command Response
	MessageData      api.SendMessageData `json:"message_data,omitempty"`
	MessageEphemeral bool                `json:"message_ephemeral,omitempty"`

	// Event Entry
	EventType string `json:"event_type,omitempty"`

	// Event Filter
	EventFilterTarget     EventFilterTarget `json:"event_filter_target,omitempty"`
	EventFilterExpression string            `json:"event_filter_expression,omitempty"`

	// Log
	LogLevel   LogLevel `json:"log_level,omitempty"`
	LogMessage string   `json:"log_message,omitempty"`

	// Condition
	ConditionBaseValue     FlowValue         `json:"condition_base_value,omitempty"`
	ConditionAllowMultiple bool              `json:"condition_allow_multiple,omitempty"`
	ConditionItemMode      ConditionItemType `json:"condition_item_mode,omitempty"`
	ConditionItemValue     FlowValue         `json:"condition_item_value,omitempty"`
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
		validation.Field(&d.Name, validation.When(nodeType == FlowNodeTypeOptionCommandArgument,
			validation.Required,
			validation.Length(1, 32),
		)),
		validation.Field(&d.Description, validation.When(nodeType == FlowNodeTypeOptionCommandArgument,
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

type CommandArgumentType string

const (
	CommandArgumentTypeString      CommandArgumentType = "string"
	CommandArgumentTypeInteger     CommandArgumentType = "integer"
	CommandArgumentTypeBoolean     CommandArgumentType = "boolean"
	CommandArgumentTypeUser        CommandArgumentType = "user"
	CommandArgumentTypeRole        CommandArgumentType = "role"
	CommandArgumentTypeChannel     CommandArgumentType = "channel"
	CommandArgumentTypeMentionable CommandArgumentType = "mentionable"
	CommandArgumentTypeNumber      CommandArgumentType = "number"
	CommandArgumentTypeAttachment  CommandArgumentType = "attachment"
)

type EventFilterTarget string

const (
	EventFilterTypeMessageContent EventFilterTarget = "message_content"
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
