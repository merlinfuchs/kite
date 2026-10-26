package flow

import (
	"encoding/json"
	"errors"
	"sync"
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
	return name + ": " + e.Next.Error()
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

var blockTitles = sync.OnceValue(func() map[FlowNodeType]string {
	var catalog struct {
		Nodes map[FlowNodeType]struct {
			Title string `json:"title"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(CatalogJSON, &catalog); err != nil {
		panic(err)
	}

	titles := make(map[FlowNodeType]string, len(catalog.Nodes))
	for nodeType, node := range catalog.Nodes {
		titles[nodeType] = node.Title
	}
	return titles
})

func blockTitle(nodeType FlowNodeType) string {
	if title := blockTitles()[nodeType]; title != "" {
		return title
	}
	return string(nodeType)
}
