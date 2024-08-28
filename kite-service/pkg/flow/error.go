package flow

import "fmt"

type FlowNodeErrorCode string

const (
	FlowNodeErrorUnknown                 FlowNodeErrorCode = "unknown"
	FlowNodeErrorUnknownNodeType         FlowNodeErrorCode = "unknown_node_type"
	FlowNodeErrorMaxStackDepthReached    FlowNodeErrorCode = "max_stack_depth_reached"
	FlowNodeErrorMaxOperationsReached    FlowNodeErrorCode = "max_operations_reached"
	FlowNodeErrorMaxActionsReached       FlowNodeErrorCode = "max_actions_reached"
	FlowNodeErrorMaxExecutionTimeReached FlowNodeErrorCode = "max_execution_time_reached"
	FlowNodeErrorTimeout                 FlowNodeErrorCode = "timeout"
)

type FlowError struct {
	Code    FlowNodeErrorCode
	Message string
}

func (e *FlowError) Error() string {
	return fmt.Sprintf("Flow error (%s): %s", e.Code, e.Message)
}

type FlowErrorTrace struct {
	NodeID          string
	NodeType        FlowNodeType
	NodeCustomLabel string
	Next            error
}

func (e *FlowErrorTrace) Error() string {
	return fmt.Sprintf("Flow error (%s): %s", e.NodeType, e.Next.Error())
}

func traceError(node *CompiledFlowNode, err error) error {
	if err == nil {
		return nil
	}

	return &FlowErrorTrace{
		NodeID:          node.ID,
		NodeType:        node.Type,
		NodeCustomLabel: node.Data.CustomLabel,
		Next:            err,
	}
}
