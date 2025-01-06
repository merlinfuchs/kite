package flow

import (
	"regexp"

	"github.com/diamondburned/arikawa/v3/api"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

// Allows between 1 and 3 words, each between 1 and 32 characters long.
var commandNameRe = regexp.MustCompile(`^[-_\p{L}\p{N}]{1,32}( [-_\p{L}\p{N}]{1,32}){0,2}$`)

// Allows only alphanumeric characters and underscores.
var commandOptionNameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

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
	FlowNodeTypeEntryComponentButton FlowNodeType = "entry_component_button"

	FlowNodeTypeOptionCommandArgument    FlowNodeType = "option_command_argument"
	FlowNodeTypeOptionCommandPermissions FlowNodeType = "option_command_permissions"
	FlowNodeTypeOptionCommandContexts    FlowNodeType = "option_command_contexts"
	FlowNodeTypeOptionEventFilter        FlowNodeType = "option_event_filter"

	FlowNodeTypeActionResponseCreate       FlowNodeType = "action_response_create"
	FlowNodeTypeActionResponseEdit         FlowNodeType = "action_response_edit"
	FlowNodeTypeActionResponseDelete       FlowNodeType = "action_response_delete"
	FlowNodeTypeActionResponseDefer        FlowNodeType = "action_response_defer"
	FlowNodeTypeActionMessageCreate        FlowNodeType = "action_message_create"
	FlowNodeTypeActionMessageEdit          FlowNodeType = "action_message_edit"
	FlowNodeTypeActionMessageDelete        FlowNodeType = "action_message_delete"
	FlowNodeTypeActionPrivateMessageCreate FlowNodeType = "action_private_message_create"
	FlowNodeTypeActionMemberBan            FlowNodeType = "action_member_ban"
	FlowNodeTypeActionMemberUnban          FlowNodeType = "action_member_unban"
	FlowNodeTypeActionMemberKick           FlowNodeType = "action_member_kick"
	FlowNodeTypeActionMemberTimeout        FlowNodeType = "action_member_timeout"
	FlowNodeTypeActionMemberEdit           FlowNodeType = "action_member_edit"
	FlowNodeTypeActionMemberRoleAdd        FlowNodeType = "action_member_role_add"
	FlowNodeTypeActionMemberRoleRemove     FlowNodeType = "action_member_role_remove"
	FlowNodeTypeActionHTTPRequest          FlowNodeType = "action_http_request"
	FlowNodeTypeActionAIChatCompletion     FlowNodeType = "action_ai_chat_completion"
	FlowNodeTypeActionExpressionEvaluate   FlowNodeType = "action_expression_evaluate"
	FlowNodeTypeActionLog                  FlowNodeType = "action_log"
	FlowNodeTypeActionVariableSet          FlowNodeType = "action_variable_set"
	FlowNodeTypeActionVariableDelete       FlowNodeType = "action_variable_delete"
	FlowNodeTypeActionVariableGet          FlowNodeType = "action_variable_get"

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
	FlowNodeTypeControlSleep                FlowNodeType = "control_sleep"
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
	Name           string     `json:"name,omitempty"`
	Description    string     `json:"description,omitempty"`
	CustomLabel    string     `json:"custom_label,omitempty"`
	AuditLogReason FlowString `json:"audit_log_reason,omitempty"`

	// Command Argument
	CommandArgumentType     CommandArgumentType `json:"command_argument_type,omitempty"`
	CommandArgumentRequired bool                `json:"command_argument_required,omitempty"`

	// Command Permissions
	CommandPermissions string `json:"command_permissions,omitempty"`

	// Command Contexts
	CommandDisabledContexts []CommandContextType `json:"command_disabled_contexts,omitempty"`
	// Command Installations
	CommandDisabledIntegrations []CommandDisabledIntegrationType `json:"command_disabled_integrations,omitempty"`

	// Message & Response Create, edit, Delete
	MessageTarget     FlowString           `json:"message_target,omitempty"`
	MessageData       *message.MessageData `json:"message_data,omitempty"`
	MessageTemplateID string               `json:"message_template_id,omitempty"`
	MessageEphemeral  bool                 `json:"message_ephemeral,omitempty"`

	// Member Ban, Kick, Timeout, Edit
	UserTarget                            FlowString            `json:"user_target,omitempty"`
	MemberBanDeleteMessageDurationSeconds FlowString            `json:"member_ban_delete_message_duration_seconds,omitempty"`
	MemberTimeoutDurationSeconds          FlowString            `json:"member_timeout_duration_seconds,omitempty"`
	MemberData                            *api.ModifyMemberData `json:"member_data,omitempty"`

	// Channel Create, Edit, Delete
	ChannelTarget FlowString             `json:"channel_target,omitempty"`
	ChannelData   *api.CreateChannelData `json:"channel_data,omitempty"`

	// Role Create, Edit, Delete
	RoleTarget FlowString          `json:"role_target,omitempty"`
	RoleData   *api.CreateRoleData `json:"role_data,omitempty"`

	// Variable Set, Delete
	VariableID        string            `json:"variable_id,omitempty"`
	VariableScope     FlowString        `json:"variable_scope,omitempty"`
	VariableValue     FlowString        `json:"variable_value,omitempty"`
	VariableOperation VariableOperation `json:"variable_operation,omitempty"`

	// HTTP Request
	HTTPRequestData *HTTPRequestData `json:"http_request_data,omitempty"`

	// AI Chat Completion
	AIChatCompletionData *AIChatCompletionData `json:"ai_chat_completion_data,omitempty"`

	// Event Entry
	EventType string `json:"event_type,omitempty"`

	// Event Filter
	EventFilterTarget     EventFilterTarget `json:"event_filter_target,omitempty"`
	EventFilterExpression string            `json:"event_filter_expression,omitempty"`

	// Log
	LogLevel   LogLevel   `json:"log_level,omitempty"`
	LogMessage FlowString `json:"log_message,omitempty"`

	// Expression Evaluate
	Expression FlowString `json:"expression,omitempty"`

	// Condition
	ConditionBaseValue     FlowString        `json:"condition_base_value,omitempty"`
	ConditionAllowMultiple bool              `json:"condition_allow_multiple,omitempty"`
	ConditionItemMode      ConditionItemType `json:"condition_item_mode,omitempty"`
	ConditionItemValue     FlowString        `json:"condition_item_value,omitempty"`
	// Loop
	LoopCount FlowString `json:"loop_count,omitempty"`
	// Sleep
	SleepDurationSeconds FlowString `json:"sleep_duration_seconds,omitempty"`
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
			validation.Match(commandOptionNameRe),
		)),
		validation.Field(&d.Description, validation.When(nodeType == FlowNodeTypeOptionCommandArgument,
			validation.Required,
			validation.Length(1, 100),
		)),

		// Event Entry
		validation.Field(&d.EventType, validation.When(nodeType == FlowNodeTypeEntryEvent,
			validation.Required,
		)),
		validation.Field(&d.Description, validation.When(nodeType == FlowNodeTypeEntryEvent,
			validation.Required,
			validation.Length(1, 100),
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

type VariableOperation string

const (
	VariableOperationOverwrite VariableOperation = "overwrite"
	VariableOperationAppend    VariableOperation = "append"
	VariableOperationPrepend   VariableOperation = "prepend"
	VariableOperationIncrement VariableOperation = "increment"
	VariableOperationDecrement VariableOperation = "decrement"
)

func (o VariableOperation) IsOverwrite() bool {
	return o == VariableOperationOverwrite || o == ""
}

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

type CommandDisabledIntegrationType string

const (
	CommandDisabledIntegrationTypeGuildInstall CommandDisabledIntegrationType = "guild_install"
	CommandDisabledIntegrationTypeUserInstall  CommandDisabledIntegrationType = "user_install"
)

type EventFilterTarget string

const (
	EventFilterTypeMessageContent EventFilterTarget = "message_content"
)

type HTTPRequestData struct {
	URL    FlowString `json:"url,omitempty"`
	Method string     `json:"method,omitempty"`
}

type AIChatCompletionData struct {
	SystemPrompt        FlowString `json:"system_prompt,omitempty"`
	Prompt              FlowString `json:"prompt,omitempty"`
	MaxCompletionTokens FlowString `json:"max_completion_tokens,omitempty"`
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
