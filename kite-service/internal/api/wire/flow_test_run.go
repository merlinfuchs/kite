package wire

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

// FlowTestRequest runs a flow without changing anything. Reads from Discord
// are real, changes are returned instead of made, and requests to services
// other than Discord aren't sent unless the block is mocked.
type FlowTestRequest struct {
	// A draft of the flow to test instead of the saved one, for commands and
	// event listeners.
	FlowSource *flow.FlowData `json:"flow_source,omitempty"`
	// A draft of the message template to test instead of the saved one.
	Message *FlowTestMessage `json:"message,omitempty"`
	// The button or select menu of a message to test, by its flow.
	FlowSourceID string `json:"flow_source_id,omitempty"`

	// Who triggers the flow, by default the user of the request, and where.
	UserID    string `json:"user_id,omitempty"`
	GuildID   string `json:"guild_id,omitempty"`
	ChannelID string `json:"channel_id,omitempty"`
	// Runs a command in a DM with the bot.
	DM bool `json:"dm,omitempty"`

	// Arguments of a command by name.
	Options map[string]json.RawMessage `json:"options,omitempty" tstype:"{ [key: string]: any }"`
	// Values picked in the select menu of a message.
	Values []string `json:"values,omitempty"`
	// Merged into the Discord event or interaction, in Discord's format.
	Event   json.RawMessage  `json:"event,omitempty" tstype:"any"`
	Webhook *FlowTestWebhook `json:"webhook,omitempty"`

	// What blocks do instead of running, by block ID.
	Mocks map[string]FlowTestMock `json:"mocks,omitempty"`
	// Makes blocks like Random pick the same each time a test with the same
	// seed runs, so it can run again with a click on a button the flow sent.
	// 0 picks a seed.
	Seed uint64 `json:"seed,omitempty"`
	// Values of stored variables when the test starts, in place of the stored
	// ones.
	Variables []FlowTestVariable `json:"variables,omitempty"`
	// What the user does after the flow ran, in order.
	Then []FlowTestAction `json:"then,omitempty"`
}

func (req FlowTestRequest) Validate() error {
	return validation.ValidateStruct(&req,
		// Drafts are checked like the flows that are saved.
		validation.Field(&req.FlowSource),
		validation.Field(&req.Message),
		validation.Field(&req.Mocks, validation.Length(0, 100)),
		validation.Field(&req.Variables, validation.Length(0, 100)),
		validation.Field(&req.Values, validation.Length(0, 25)),
		validation.Field(&req.Options, validation.Length(0, 25)),
		validation.Field(&req.Then),
	)
}

type FlowTestMessage struct {
	Data        message.MessageData      `json:"data"`
	FlowSources map[string]flow.FlowData `json:"flow_sources"`
}

func (m FlowTestMessage) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.FlowSources),
	)
}

type FlowTestWebhook struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers,omitempty"`
	Query   map[string]string `json:"query,omitempty"`
}

type FlowTestMock struct {
	// The block's result, in the format of its result type.
	Result json.RawMessage `json:"result,omitempty" tstype:"any"`
	// Makes the block fail with this error.
	Error string `json:"error,omitempty"`
}

type FlowTestVariable struct {
	VariableID string `json:"variable_id"`
	// The scope, for variables with a value per scope, as the block
	// evaluates it, like "<server ID>:<user ID>".
	Scope *string `json:"scope,omitempty"`
	// The value, or null for no value.
	Value json.RawMessage `json:"value,omitempty" tstype:"any"`
	// Type makes value text that is read like on the page of the variable,
	// like "number" for "5". An invalid value is an error.
	Type VariableValueType `json:"type,omitempty"`
}

func (v FlowTestVariable) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.VariableID, validation.Required),
	)
}

type FlowTestAction struct {
	// Label or custom ID of a button to click.
	Click string `json:"click,omitempty"`
	// Placeholder or custom ID of a select menu, with the values to pick.
	Select string   `json:"select,omitempty"`
	Values []string `json:"values,omitempty"`
	// Values of the inputs of the open modal, by custom ID or label.
	Submit map[string]json.RawMessage `json:"submit,omitempty" tstype:"{ [key: string]: any }"`
	UserID string                     `json:"user_id,omitempty"`
}

func (a FlowTestAction) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Values, validation.Length(0, 25)),
		validation.Field(&a.Submit, validation.Length(0, 10)),
	)
}

type FlowTestResponse struct {
	Executions []FlowTestExecution `json:"executions"`
	// What the test would have cost.
	Credits int `json:"credits"`
	// Mocked blocks that never ran.
	UnusedMocks []string `json:"unused_mocks,omitempty"`
	// The index of the action in then that couldn't be done. Its error is
	// the last execution.
	FailedAction *int `json:"failed_action,omitempty"`
}

// FlowTestExecution is one run of the flow, like the command or a click on a
// button it sent.
type FlowTestExecution struct {
	Trigger   string         `json:"trigger"`
	Steps     []FlowTestStep `json:"steps"`
	Truncated bool           `json:"truncated,omitempty"`
	Notes     []string       `json:"notes,omitempty"`
	// The error that ended the run, if no Error Handler caught it.
	Error string `json:"error,omitempty"`
	// Set for runs of a flow other than the tested one, like the flow of a
	// button on a message template it sent. Their blocks aren't in the
	// tested flow, even if they have the same IDs.
	OtherFlow bool `json:"other_flow,omitempty"`
}

// FlowTestStep is a block that ran.
type FlowTestStep struct {
	NodeID    string           `json:"node_id"`
	NodeType  string           `json:"node_type"`
	Result    any              `json:"result,omitempty" tstype:"any"`
	Error     string           `json:"error,omitempty"`
	ErrorCode string           `json:"error_code,omitempty"`
	Notes     []flow.TestNote  `json:"notes,omitempty"`
	Effects   []FlowTestEffect `json:"effects,omitempty"`
}

// FlowTestEffect is a change the block would have made.
type FlowTestEffect struct {
	Action   string   `json:"action"`
	Data     any      `json:"data,omitempty" tstype:"any"`
	Warnings []string `json:"warnings,omitempty"`
}
