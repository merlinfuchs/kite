package flow

import (
	"context"
	"errors"

	"github.com/diamondburned/arikawa/v3/utils/httputil"
)

type FlowNodeErrorCode string

const (
	FlowNodeErrorUnknown                 FlowNodeErrorCode = "unknown"
	FlowNodeErrorUnknownNodeType         FlowNodeErrorCode = "unknown_node_type"
	FlowNodeErrorMaxStackDepthReached    FlowNodeErrorCode = "max_stack_depth_reached"
	FlowNodeErrorMaxOperationsReached    FlowNodeErrorCode = "max_operations_reached"
	FlowNodeErrorMaxCreditsReached       FlowNodeErrorCode = "max_credits_reached"
	FlowNodeErrorMaxExecutionTimeReached FlowNodeErrorCode = "max_execution_time_reached"
	FlowNodeErrorTimeout                 FlowNodeErrorCode = "timeout"
	FlowNodeErrorDiscordAPI              FlowNodeErrorCode = "discord_api"
	FlowNodeErrorIntegrationAPI          FlowNodeErrorCode = "integration_api"
	FlowNodeErrorInvalidSetting          FlowNodeErrorCode = "invalid_setting"
)

type FlowError struct {
	Code    FlowNodeErrorCode
	Message string
}

func (e *FlowError) Error() string {
	return e.Message
}

// FlowErrorTrace is the block an error happened in.
type FlowErrorTrace struct {
	NodeID          string
	NodeType        FlowNodeType
	NodeCustomLabel string
	Next            error
}

func (e *FlowErrorTrace) Error() string {
	name := blockTitle(e.NodeType)
	if e.NodeCustomLabel != "" {
		name += ` "` + e.NodeCustomLabel + `"`
	}
	return name + " (block " + e.NodeID + "): " + e.Next.Error()
}

func (e *FlowErrorTrace) Unwrap() error {
	return e.Next
}

// traceError records that err happened in node. Errors that already have a
// trace keep it, so they point at the block they happened in and not at the
// blocks they passed through on the way up.
func traceError(node *CompiledFlowNode, err error) error {
	if err == nil {
		return nil
	}

	var trace *FlowErrorTrace
	if errors.As(err, &trace) {
		return err
	}

	return &FlowErrorTrace{
		NodeID:          node.ID,
		NodeType:        node.Type,
		NodeCustomLabel: node.Data.CustomLabel,
		Next:            err,
	}
}

// settingError is an invalid value of a block's setting.
type settingError struct {
	Field string
	Err   error
}

func (e *settingError) Error() string {
	return e.Err.Error()
}

func (e *settingError) Unwrap() error {
	return e.Err
}

// integrationError is an error status returned by an integration's API.
type integrationError struct {
	Integration string
	Status      int
	Err         error
}

func (e *integrationError) Error() string {
	return e.Err.Error()
}

func (e *integrationError) Unwrap() error {
	return e.Err
}

// ErrorInfo is what the logs store about an error of a flow, so the editor
// can point at the block it happened in.
type ErrorInfo struct {
	NodeID   string
	NodeType FlowNodeType
	Code     FlowNodeErrorCode
	// Nil when there's nothing more to tell.
	Details *ErrorDetails
}

type ErrorDetails struct {
	DiscordStatus int    `json:"discord_status,omitempty"`
	DiscordCode   int    `json:"discord_code,omitempty"`
	Integration   string `json:"integration,omitempty"`
	HTTPStatus    int    `json:"http_status,omitempty"`
	Field         string `json:"field,omitempty"`
}

func DescribeError(err error) ErrorInfo {
	info := ErrorInfo{Code: FlowNodeErrorUnknown}

	var trace *FlowErrorTrace
	if errors.As(err, &trace) {
		info.NodeID = trace.NodeID
		info.NodeType = trace.NodeType
	}

	var details ErrorDetails
	var flowErr *FlowError
	var discordErr *httputil.HTTPError
	var integrationErr *integrationError
	var settingErr *settingError
	switch {
	case errors.As(err, &flowErr):
		info.Code = flowErr.Code
	case errors.As(err, &discordErr):
		info.Code = FlowNodeErrorDiscordAPI
		details.DiscordStatus = discordErr.Status
		details.DiscordCode = int(discordErr.Code)
	case errors.As(err, &integrationErr):
		info.Code = FlowNodeErrorIntegrationAPI
		details.Integration = integrationErr.Integration
		details.HTTPStatus = integrationErr.Status
	case errors.Is(err, context.DeadlineExceeded):
		info.Code = FlowNodeErrorMaxExecutionTimeReached
	case errors.As(err, &settingErr):
		info.Code = FlowNodeErrorInvalidSetting
		details.Field = settingErr.Field
	}

	if details != (ErrorDetails{}) {
		info.Details = &details
	}
	return info
}

func blockTitle(nodeType FlowNodeType) string {
	if title := blockDefinitions[nodeType].Title; title != "" {
		return title
	}
	return string(nodeType)
}
