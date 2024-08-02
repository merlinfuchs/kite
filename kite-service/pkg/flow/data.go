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
	FlowNodeTypeEntryCommand FlowNodeType = "entry_command"
	FlowNodeTypeEntryEvent   FlowNodeType = "entry_event"

	FlowNodeTypeOptionCommandArgument    FlowNodeType = "option_command_argument"
	FlowNodeTypeOptionCommandPermissions FlowNodeType = "option_command_permissions"
	FlowNodeTypeOptionCommandContexts    FlowNodeType = "option_command_contexts"
	FlowNodeTypeOptionEventFilter        FlowNodeType = "option_event_filter"

	FlowNodeTypeActionResponseCreate FlowNodeType = "action_response_create"
	FlowNodeTypeActionResponseEdit   FlowNodeType = "action_response_edit"
	FlowNodeTypeActionResponseDelete FlowNodeType = "action_response_delete"
	FlowNodeTypeActionMessageCreate  FlowNodeType = "action_message_create"
	FlowNodeTypeActionMessageEdit    FlowNodeType = "action_message_edit"
	FlowNodeTypeActionMessageDelete  FlowNodeType = "action_message_delete"
	FlowNodeTypeActionMemberBan      FlowNodeType = "action_member_ban"
	FlowNodeTypeActionMemberKick     FlowNodeType = "action_member_kick"
	FlowNodeTypeActionMemberTimeout  FlowNodeType = "action_member_timeout"
	FlowNodeTypeActionChannelCreate  FlowNodeType = "action_channel_create"
	FlowNodeTypeActionChannelEdit    FlowNodeType = "action_channel_edit"
	FlowNodeTypeActionChannelDelete  FlowNodeType = "action_channel_delete"
	FlowNodeTypeActionThreadCreate   FlowNodeType = "action_thread_create"
	FlowNodeTypeActionRoleCreate     FlowNodeType = "action_role_create"
	FlowNodeTypeActionRoleEdit       FlowNodeType = "action_role_edit"
	FlowNodeTypeActionRoleDelete     FlowNodeType = "action_role_delete"
	FlowNodeTypeActionVariableSet    FlowNodeType = "action_variable_set"
	FlowNodeTypeActionVariableDelete FlowNodeType = "action_variable_delete"
	FlowNodeTypeActionHTTPRequest    FlowNodeType = "action_http_request"
	FlowNodeTypeActionLog            FlowNodeType = "action_log"

	FlowNodeTypeControlConditionCompare     FlowNodeType = "control_condition_compare"
	FlowNodeTypeControlConditionItemCompare FlowNodeType = "control_condition_item_compare"
	FlowNodeTypeControlConditionUser        FlowNodeType = "control_condition_user"
	FlowNodeTypeControlConditionItemUser    FlowNodeType = "control_condition_item_user"
	FlowNodeTypeControlConditionChannel     FlowNodeType = "control_condition_channel"
	FlowNodeTypeControlConditionItemChannel FlowNodeType = "control_condition_item_channel"
	FlowNodeTypeControlConditionRole        FlowNodeType = "control_condition_role"
	FlowNodeTypeControlConditionItemRole    FlowNodeType = "control_condition_item_role"
	FlowNodeTypeControlConditionItemElse    FlowNodeType = "control_condition_item_else"
	FlowNodeTypeControlLoop                 FlowNodeType = "control_loop"
	FlowNodeTypeControlLoopEach             FlowNodeType = "control_loop_each"
	FlowNodeTypeControlLoopEnd              FlowNodeType = "control_loop_end"
	FlowNodeTypeControlLoopExit             FlowNodeType = "control_loop_exit"
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
	AuditLogReason     string `json:"audit_log_reason,omitempty"`

	// Command Argument
	CommandArgumentType     CommandArgumentType `json:"command_argument_type,omitempty"`
	CommandArgumentRequired bool                `json:"command_argument_required,omitempty"`

	// Command Permissions
	CommandPermissions string `json:"command_permissions,omitempty"`

	// Command Contexts
	CommandDisabledContexts []CommandContextType `json:"command_disabled_contexts,omitempty"`

	// Message & Response Create, edit, Delete
	MessageTarget     string              `json:"message_target,omitempty"`
	MessageData       api.SendMessageData `json:"message_data,omitempty"`
	MessageTemplateID string              `json:"message_template_id,omitempty"`
	MessageEphemeral  bool                `json:"message_ephemeral,omitempty"`

	// Member Ban, Kick, Timeout
	MemberTarget                   string `json:"member_target,omitempty"`
	MemberBanDeleteMessageDuration string `json:"member_ban_delete_message_duration,omitempty"`
	MemberTimeoutDuration          string `json:"member_timeout_duration,omitempty"`

	// Channel Create, Edit, Delete
	ChannelTarget string                `json:"channel_target,omitempty"`
	ChannelData   api.CreateChannelData `json:"channel_data,omitempty"`

	// Role Create, Edit, Delete
	RoleTarget string             `json:"role_target,omitempty"`
	RoleData   api.CreateRoleData `json:"role_data,omitempty"`

	// Variable Set, Delete
	VariableName  string    `json:"variable_name,omitempty"`
	VariableValue FlowValue `json:"variable_value,omitempty"`

	// HTTP Request
	HTTPRequestData HTTPRequestData `json:"http_request_data,omitempty"`

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
	// Loop
	LoopCount string `json:"loop_count,omitempty"`
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

type CommandContextType string

const (
	CommandContextTypeGuild          CommandContextType = "guild"
	CommandContextTypeBotDM          CommandContextType = "bot_dm"
	CommandContextTypePrivateChannel CommandContextType = "private_channel"
)

type EventFilterTarget string

const (
	EventFilterTypeMessageContent EventFilterTarget = "message_content"
)

type HTTPRequestData struct {
	URL    string `json:"url,omitempty"`
	Method string `json:"method,omitempty"`
}

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
