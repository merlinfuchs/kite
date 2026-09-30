package flow

import "fmt"

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeEntryCommand:         executeEntryCommand,
		FlowNodeTypeEntryComponentButton: executeEntryCommand,
		FlowNodeTypeEntryEvent:           executeEntryEvent,
	})
}

func executeEntryCommand(n *CompiledFlowNode, ctx *FlowContext) error {
	if !ctx.IsEntry() {
		return fmt.Errorf("command entry isn't the entry node")
	}

	err := n.autoDeferInteraction(ctx)
	if err != nil {
		return traceError(n, err)
	}

	err = n.ExecuteChildren(ctx)
	if err != nil {
		createDefaultErrorResponse(ctx, err)
		return traceError(n, err)
	}

	acknowledgeUnansweredComponent(ctx)
	return nil
}

func executeEntryEvent(n *CompiledFlowNode, ctx *FlowContext) error {
	if !ctx.IsEntry() {
		return fmt.Errorf("event entry isn't the entry node")
	}

	err := n.ExecuteChildren(ctx)
	if err != nil {
		return traceError(n, err)
	}

	return nil
}
