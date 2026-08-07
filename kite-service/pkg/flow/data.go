package flow

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/openai/openai-go"
	"gopkg.in/guregu/null.v4"
)

// Allows between 1 and 3 words, each between 1 and 32 characters long.
var commandNameRe = regexp.MustCompile(`^[-_a-z0-9]{1,32}( [-_a-z0-9]{1,32}){0,2}$`)

// Allows only lowercase alphanumeric characters and underscores.
var commandOptionNameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

// Allows only lowercase alphanumeric characters and underscores.
var resultKeyRe = regexp.MustCompile(`^[a-z0-9_]+$`)

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

	FlowNodeTypeActionResponseCreate        FlowNodeType = "action_response_create"
	FlowNodeTypeActionResponseEdit          FlowNodeType = "action_response_edit"
	FlowNodeTypeActionResponseDelete        FlowNodeType = "action_response_delete"
	FlowNodeTypeActionResponseDefer         FlowNodeType = "action_response_defer"
	FlowNodeTypeActionMessageCreate         FlowNodeType = "action_message_create"
	FlowNodeTypeActionMessageEdit           FlowNodeType = "action_message_edit"
	FlowNodeTypeActionMessageDelete         FlowNodeType = "action_message_delete"
	FlowNodeTypeActionPrivateMessageCreate  FlowNodeType = "action_private_message_create"
	FlowNodeTypeActionMessageReactionCreate FlowNodeType = "action_message_reaction_create"
	FlowNodeTypeActionMessageReactionDelete FlowNodeType = "action_message_reaction_delete"
	FlowNodeTypeActionMemberBan             FlowNodeType = "action_member_ban"
	FlowNodeTypeActionMemberUnban           FlowNodeType = "action_member_unban"
	FlowNodeTypeActionMemberKick            FlowNodeType = "action_member_kick"
	FlowNodeTypeActionMemberTimeout         FlowNodeType = "action_member_timeout"
	FlowNodeTypeActionMemberEdit            FlowNodeType = "action_member_edit"
	FlowNodeTypeActionMemberRoleAdd         FlowNodeType = "action_member_role_add"
	FlowNodeTypeActionMemberRoleRemove      FlowNodeType = "action_member_role_remove"
	FlowNodeTypeActionMemberGet             FlowNodeType = "action_member_get"
	FlowNodeTypeActionUserGet               FlowNodeType = "action_user_get"
	FlowNodeTypeActionChannelGet            FlowNodeType = "action_channel_get"
	FlowNodeTypeActionChannelCreate         FlowNodeType = "action_channel_create"
	FlowNodeTypeActionChannelEdit           FlowNodeType = "action_channel_edit"
	FlowNodeTypeActionChannelDelete         FlowNodeType = "action_channel_delete"
	FlowNodeTypeActionThreadCreate          FlowNodeType = "action_thread_create"
	FlowNodeTypeActionThreadMemberAdd       FlowNodeType = "action_thread_member_add"
	FlowNodeTypeActionThreadMemberRemove    FlowNodeType = "action_thread_member_remove"
	FlowNodeTypeActionForumPostCreate       FlowNodeType = "action_forum_post_create"
	FlowNodeTypeActionRoleGet               FlowNodeType = "action_role_get"
	FlowNodeTypeActionGuildGet              FlowNodeType = "action_guild_get"
	FlowNodeTypeActionMessageGet            FlowNodeType = "action_message_get"
	FlowNodeTypeActionRobloxUserGet         FlowNodeType = "action_roblox_user_get"
	FlowNodeTypeActionHTTPRequest           FlowNodeType = "action_http_request"
	FlowNodeTypeActionAIChatCompletion      FlowNodeType = "action_ai_chat_completion"
	FlowNodeTypeActionAISearchWeb           FlowNodeType = "action_ai_web_search"
	FlowNodeTypeActionExpressionEvaluate    FlowNodeType = "action_expression_evaluate"
	FlowNodeTypeActionRandomGenerate        FlowNodeType = "action_random_generate"
	FlowNodeTypeActionLog                   FlowNodeType = "action_log"
	FlowNodeTypeActionVariableSet           FlowNodeType = "action_variable_set"
	FlowNodeTypeActionVariableDelete        FlowNodeType = "action_variable_delete"
	FlowNodeTypeActionVariableGet           FlowNodeType = "action_variable_get"

	FlowNodeTypeControlConditionCompare     FlowNodeType = "control_condition_compare"
	FlowNodeTypeControlConditionItemCompare FlowNodeType = "control_condition_item_compare"
	FlowNodeTypeControlConditionUser        FlowNodeType = "control_condition_user"
	FlowNodeTypeControlConditionItemUser    FlowNodeType = "control_condition_item_user"
	FlowNodeTypeControlConditionChannel     FlowNodeType = "control_condition_channel"
	FlowNodeTypeControlConditionItemChannel FlowNodeType = "control_condition_item_channel"
	FlowNodeTypeControlConditionRole        FlowNodeType = "control_condition_role"
	FlowNodeTypeControlConditionItemRole    FlowNodeType = "control_condition_item_role"
	FlowNodeTypeControlConditionItemElse    FlowNodeType = "control_condition_item_else"
	FlowNodeTypeControlErrorHandler         FlowNodeType = "control_error_handler"
	FlowNodeTypeControlLoop                 FlowNodeType = "control_loop"
	FlowNodeTypeControlLoopEach             FlowNodeType = "control_loop_each"
	FlowNodeTypeControlLoopEnd              FlowNodeType = "control_loop_end"
	FlowNodeTypeControlLoopExit             FlowNodeType = "control_loop_exit"
	FlowNodeTypeControlSleep                FlowNodeType = "control_sleep"

	FlowNodeTypeSuspendResponseModal FlowNodeType = "suspend_response_modal"
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
	Name           string `json:"name,omitempty"`
	Description    string `json:"description,omitempty"`
	CustomLabel    string `json:"custom_label,omitempty"`
	AuditLogReason string `json:"audit_log_reason,omitempty"`

	// Temporary Variables
	TemporaryName string `json:"temporary_name,omitempty"`

	// Command Argument
	CommandArgumentType      CommandArgumentType         `json:"command_argument_type,omitempty"`
	CommandArgumentRequired  bool                        `json:"command_argument_required,omitempty"`
	CommandArgumentChoices   []CommandArgumentChoiceData `json:"command_argument_choices,omitempty"`
	CommandArgumentMinValue  float64                     `json:"command_argument_min_value,omitempty"`
	CommandArgumentMaxValue  float64                     `json:"command_argument_max_value,omitempty"`
	CommandArgumentMaxLength int                         `json:"command_argument_max_length,omitempty"`

	// Command Permissions
	CommandPermissions string `json:"command_permissions,omitempty"`

	// Command Contexts
	CommandDisabledContexts []CommandContextType `json:"command_disabled_contexts,omitempty"`
	// Command Installations
	CommandDisabledIntegrations []CommandDisabledIntegrationType `json:"command_disabled_integrations,omitempty"`

	// Guild Get
	GuildTarget string `json:"guild_target,omitempty"`

	// Message & Response Create, Edit, Delete
	MessageTarget     string               `json:"message_target,omitempty"`
	MessageData       *message.MessageData `json:"message_data,omitempty"`
	MessageTemplateID string               `json:"message_template_id,omitempty"`
	MessageEphemeral  bool                 `json:"message_ephemeral,omitempty"`

	// Message Reaction Create, Delete
	EmojiData *EmojiData `json:"emoji_data,omitempty"`

	// Modal
	ModalData *ModalData `json:"modal_data,omitempty"`

	// Member Ban, Kick, Timeout, Edit, Get
	UserTarget                            string      `json:"user_target,omitempty"`
	MemberBanDeleteMessageDurationSeconds string      `json:"member_ban_delete_message_duration_seconds,omitempty"`
	MemberTimeoutDurationSeconds          string      `json:"member_timeout_duration_seconds,omitempty"`
	MemberData                            *MemberData `json:"member_data,omitempty"`

	// Channel Create, Edit, Delete, Get
	ChannelTarget string       `json:"channel_target,omitempty"`
	ChannelData   *ChannelData `json:"channel_data,omitempty"`

	// Role Create, Edit, Delete, Get
	RoleTarget string    `json:"role_target,omitempty"`
	RoleData   *RoleData `json:"role_data,omitempty"`

	// Roblox User Get
	RobloxUserTarget string           `json:"roblox_user_target,omitempty"`
	RobloxLookupMode RobloxLookupType `json:"roblox_lookup_mode,omitempty"`

	// Variable Set, Delete
	VariableID        string                     `json:"variable_id,omitempty"`
	VariableScope     string                     `json:"variable_scope,omitempty"`
	VariableValue     string                     `json:"variable_value,omitempty"`
	VariableOperation provider.VariableOperation `json:"variable_operation,omitempty"`

	// HTTP Request
	HTTPRequestData *HTTPRequestData `json:"http_request_data,omitempty"`

	// AI Chat Completion
	AIChatCompletionData *AIChatCompletionData `json:"ai_chat_completion_data,omitempty"`

	// Random Generate
	RandomMin string `json:"random_min,omitempty"`
	RandomMax string `json:"random_max,omitempty"`

	// Event Entry
	EventType string `json:"event_type,omitempty"`

	// Event Filter
	EventFilterTarget EventFilterTarget `json:"event_filter_target,omitempty"`
	EventFilterMode   ComparsionMode    `json:"event_filter_mode,omitempty"`
	EventFilterValue  string            `json:"event_filter_value,omitempty"`

	// Log
	LogLevel   provider.LogLevel `json:"log_level,omitempty"`
	LogMessage string            `json:"log_message,omitempty"`

	// Expression Evaluate
	Expression string `json:"expression,omitempty"`

	// Condition
	ConditionBaseValue     string         `json:"condition_base_value,omitempty"`
	ConditionAllowMultiple bool           `json:"condition_allow_multiple,omitempty"`
	ConditionItemMode      ComparsionMode `json:"condition_item_mode,omitempty"`
	ConditionItemValue     string         `json:"condition_item_value,omitempty"`
	// Loop
	LoopCount string `json:"loop_count,omitempty"`
	// Sleep
	SleepDurationSeconds string `json:"sleep_duration_seconds,omitempty"`
}

func (d FlowNodeData) Validate(nodeType FlowNodeType) error {
	// We currently only validate data for entry nodes, as for the other nodes it's less critical that they are valid.

	return validation.ValidateStruct(&d,
		// Shared
		validation.Field(&d.TemporaryName,
			validation.Length(1, 32),
			validation.Match(resultKeyRe).Error("must be lowercase without special characters"),
		),

		// Command Entry
		validation.Field(&d.Name, validation.When(nodeType == FlowNodeTypeEntryCommand,
			validation.Required,
			validation.Length(1, 32),
			validation.Match(commandNameRe).
				Error("must be lowercase without special characters and up to two spaces"),
		)),
		validation.Field(&d.Description, validation.When(nodeType == FlowNodeTypeEntryCommand,
			validation.Required,
			validation.Length(1, 100),
		)),

		// Command Option
		validation.Field(&d.Name, validation.When(nodeType == FlowNodeTypeOptionCommandArgument,
			validation.Required,
			validation.Length(1, 32),
			validation.Match(commandOptionNameRe).
				Error("must be lowercase without special characters"),
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

		// AI Chat Completion
		// Both AI node types read AIChatCompletionData, so both have to require
		// it -- the web search node is the more expensive of the two.
		validation.Field(&d.AIChatCompletionData, validation.When(
			nodeType == FlowNodeTypeActionAIChatCompletion ||
				nodeType == FlowNodeTypeActionAISearchWeb,
			validation.Required,
		)),

		// Expression Evaluate
		// Bounded for every node type rather than just entry nodes: an oversized
		// expression is a resource-exhaustion risk at execution time, not just a
		// correctness problem. eval enforces the same limit as a backstop for
		// flows stored before this check existed.
		validation.Field(&d.Expression, validation.Length(0, eval.MaxExpressionLength)),
	)
}

type ComparsionMode string

const (
	ComparsionModeEqual              ComparsionMode = "equal"
	ComparsionModeNotEqual           ComparsionMode = "not_equal"
	ComparsionModeGreaterThan        ComparsionMode = "greater_than"
	ComparsionModeGreaterThanOrEqual ComparsionMode = "greater_than_or_equal"
	ComparsionModeLessThan           ComparsionMode = "less_than"
	ComparsionModeLessThanOrEqual    ComparsionMode = "less_than_or_equal"
	ComparsionModeContains           ComparsionMode = "contains"
	ComparsionModeStartsWith         ComparsionMode = "starts_with"
	ComparsionModeEndsWith           ComparsionMode = "ends_with"

	// User condition
	ComparsionModeHasRole          ComparsionMode = "has_role"
	ComparsionModeNotHasRole       ComparsionMode = "not_has_role"
	ComparsionModeHasPermission    ComparsionMode = "has_permission"
	ComparsionModeNotHasPermission ComparsionMode = "not_has_permission"
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
	EventFilterTypeUserID         EventFilterTarget = "user_id"
	EventFilterTypeGuildID        EventFilterTarget = "guild_id"
	EventFilterTypeChannelID      EventFilterTarget = "channel_id"
)

type RobloxLookupType string

const (
	RobloxLookupTypeID   RobloxLookupType = "id"
	RobloxLookupTypeName RobloxLookupType = "username"
)

type CommandArgumentChoiceData struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type ChannelData struct {
	Name                 string                    `json:"name,omitempty"`
	Type                 int                       `json:"type,omitempty"`
	Topic                string                    `json:"topic,omitempty"`
	NSFW                 bool                      `json:"nsfw,omitempty"`
	ParentID             string                    `json:"parent,omitempty"`
	Bitrate              string                    `json:"bitrate,omitempty"`
	UserLimit            string                    `json:"user_limit,omitempty"`
	Position             string                    `json:"position,omitempty"`
	PermissionOverwrites []PermissionOverwriteData `json:"permission_overwrites,omitempty"`

	// Thread specific
	Invitable bool `json:"invitable,omitempty"`
}

func (d *ChannelData) ToCreateChannelData(ctx context.Context, evalCtx eval.Context) (api.CreateChannelData, error) {
	res := api.CreateChannelData{
		Type:       discord.ChannelType(d.Type),
		NSFW:       d.NSFW,
		Overwrites: make([]discord.Overwrite, 0, len(d.PermissionOverwrites)),
	}

	name, err := eval.EvalTemplate(ctx, d.Name, evalCtx)
	if err != nil {
		return res, err
	}
	res.Name = name.String()

	// Fields the selected channel type doesn't support are dropped rather than
	// sent. The editor only hides their inputs when the type changes, so a
	// value entered beforehand survives in the stored data, and Discord
	// rejects the whole request for e.g. a topic on a voice channel.
	topic, err := eval.EvalTemplate(ctx, d.Topic, evalCtx)
	if err != nil {
		return res, err
	}
	if channelTypeSupportsTopic(res.Type) {
		res.Topic = topic.String()
	}

	parentID, err := eval.EvalTemplate(ctx, d.ParentID, evalCtx)
	if err != nil {
		return res, err
	}
	res.CategoryID = discord.ChannelID(parentID.Snowflake())

	bitrate, err := eval.EvalTemplate(ctx, d.Bitrate, evalCtx)
	if err != nil {
		return res, err
	}
	userLimit, err := eval.EvalTemplate(ctx, d.UserLimit, evalCtx)
	if err != nil {
		return res, err
	}
	if channelTypeSupportsVoice(res.Type) {
		res.VoiceBitrate = uint(bitrate.Int())
		res.VoiceUserLimit = uint(userLimit.Int())
	}

	position, err := eval.EvalTemplate(ctx, d.Position, evalCtx)
	if err != nil {
		return res, err
	}
	res.Position = option.NewInt(int(position.Int()))

	for _, overwrite := range d.PermissionOverwrites {
		id, err := eval.EvalTemplate(ctx, overwrite.ID, evalCtx)
		if err != nil {
			return res, err
		}

		allow, err := eval.EvalTemplate(ctx, overwrite.Allow, evalCtx)
		if err != nil {
			return res, err
		}

		deny, err := eval.EvalTemplate(ctx, overwrite.Deny, evalCtx)
		if err != nil {
			return res, err
		}

		res.Overwrites = append(res.Overwrites, discord.Overwrite{
			ID:    discord.Snowflake(id.Snowflake()),
			Type:  discord.OverwriteType(overwrite.Type),
			Allow: discord.Permissions(allow.Int()),
			Deny:  discord.Permissions(deny.Int()),
		})
	}

	return res, nil
}

type PermissionOverwriteData struct {
	ID    string `json:"id,omitempty"`
	Type  int    `json:"type,omitempty"`
	Allow string `json:"allow,omitempty"`
	Deny  string `json:"deny,omitempty"`
}

type RoleData struct {
	Name        string `json:"name,omitempty"`
	Color       int    `json:"color,omitempty"`
	Hoist       bool   `json:"hoist,omitempty"`
	Permissions int    `json:"permissions,omitempty"`
	Position    int    `json:"position,omitempty"`
}

type MemberData struct {
	Nick *string `json:"nick,omitempty"`
}

type EmojiData struct {
	ID string `json:"id,omitempty"`
	// Name is the name of a custom emoji or the unicode of a standard emoji.
	Name string `json:"name,omitempty"`
}

type ModalData struct {
	Title      string               `json:"title,omitempty"`
	Components []ModalComponentData `json:"components,omitempty"`
}

type ModalComponentData struct {
	CustomID    string               `json:"custom_id,omitempty"`
	Style       int                  `json:"style,omitempty"`
	Label       string               `json:"label,omitempty"`
	MinLength   int                  `json:"min_length,omitempty"`
	MaxLength   int                  `json:"max_length,omitempty"`
	Required    bool                 `json:"required,omitempty"`
	Value       string               `json:"value,omitempty"`
	Placeholder string               `json:"placeholder,omitempty"`
	Components  []ModalComponentData `json:"components,omitempty"`
}

type HTTPRequestData struct {
	URL      string                    `json:"url,omitempty"`
	Method   string                    `json:"method,omitempty"`
	Headers  []HTTPRequestDataKeyValue `json:"headers,omitempty"`
	Query    []HTTPRequestDataKeyValue `json:"query,omitempty"`
	BodyJSON json.RawMessage           `json:"body_json,omitempty"`
}

type HTTPRequestDataKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AIChatCompletionData struct {
	Model               string `json:"model,omitempty"`
	SystemPrompt        string `json:"system_prompt,omitempty"`
	Prompt              string `json:"prompt,omitempty"`
	MaxCompletionTokens string `json:"max_completion_tokens,omitempty"`
}

// aiModelCosts is the set of models a flow may run, and what each costs in
// credits for a plain completion versus one with web search.
//
// Single source of truth for both the save-time allowlist and metering. Pricing
// used to be a switch with a cheap default arm, which meant any model not named
// in it -- a newer, more expensive SKU, say -- billed the tenant at the floor
// while the operator paid the real rate on their own API key. Keeping the
// allowlist and the prices in one table means a model that has no price also
// cannot be selected.
//
// The empty string is the provider default (gpt-4o-mini), so it is priced.
var aiModelCosts = map[string]aiModelCost{
	"":                  {Chat: 5, Search: 25},
	openai.ChatModelGPT4_1:     {Chat: 100, Search: 500},
	openai.ChatModelGPT4_1Mini: {Chat: 20, Search: 100},
	openai.ChatModelGPT4_1Nano: {Chat: 5, Search: 25},
	openai.ChatModelGPT4oMini:  {Chat: 5, Search: 25},
}

type aiModelCost struct {
	Chat   int
	Search int
}

// maxAIModelCost is the ceiling of aiModelCosts, taken per field so it does not
// depend on one model being the most expensive for both.
var maxAIModelCost = func() aiModelCost {
	var ceiling aiModelCost
	for _, c := range aiModelCosts {
		ceiling.Chat = max(ceiling.Chat, c.Chat)
		ceiling.Search = max(ceiling.Search, c.Search)
	}
	return ceiling
}()

// aiModelsAllowed is aiModelCosts' keys as validation.In wants them. The empty
// string is left out because In treats an empty value as valid regardless.
var aiModelsAllowed = func() []any {
	models := make([]any, 0, len(aiModelCosts))
	for model := range aiModelCosts {
		if model != "" {
			models = append(models, model)
		}
	}
	return models
}()

// AIModelAllowed reports whether a flow may run the given model.
func AIModelAllowed(model string) bool {
	_, ok := aiModelCosts[model]
	return ok
}

// AICreditsCost prices one AI node.
//
// Unknown models are charged the most expensive entry rather than the least, so
// anything that reaches execution without passing the allowlist -- a flow saved
// before this existed, or one embedded in a message, which the API does not
// validate -- is over-charged rather than run for free.
func AICreditsCost(model string, webSearch bool) int {
	cost, ok := aiModelCosts[model]
	if !ok {
		cost = maxAIModelCost
	}

	if webSearch {
		return cost.Search
	}
	return cost.Chat
}

func (d AIChatCompletionData) Validate() error {
	return validation.ValidateStruct(&d,
		validation.Field(&d.Model, validation.In(aiModelsAllowed...)),
		validation.Field(&d.Prompt, validation.Required, validation.Length(1, 2000)),
	)
}

type FlowNodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type FlowEdge struct {
	ID           string      `json:"id"`
	Type         string      `json:"type,omitempty"`
	Source       string      `json:"source"`
	Target       string      `json:"target"`
	SourceHandle null.String `json:"sourceHandle,omitempty"`
	TargetHandle null.String `json:"targetHandle,omitempty"`
}

func (e FlowEdge) Validate() error {
	return validation.ValidateStruct(&e,
		validation.Field(&e.ID, validation.Required),
		validation.Field(&e.Source, validation.Required),
		validation.Field(&e.Target, validation.Required),
	)
}

// guildMediaChannel is Discord's media channel type, which arikawa doesn't
// define yet.
const guildMediaChannel discord.ChannelType = 16

// channelTypeSupportsTopic reports whether Discord accepts a topic for the
// given channel type. Voice, stage and category channels have no topic.
func channelTypeSupportsTopic(t discord.ChannelType) bool {
	switch t {
	case discord.GuildText, discord.GuildAnnouncement, discord.GuildForum, guildMediaChannel:
		return true
	default:
		return false
	}
}

// channelTypeSupportsVoice reports whether Discord accepts bitrate and user
// limit for the given channel type.
func channelTypeSupportsVoice(t discord.ChannelType) bool {
	return t == discord.GuildVoice || t == discord.GuildStageVoice
}
