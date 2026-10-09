package flow

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionHTTPRequest: executeActionHTTPRequest,
	})
}

func executeActionHTTPRequest(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.HTTPRequestData == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "http_request_data is nil",
		}
	}

	result, err := ctx.executeHTTPRequest(n.Data.HTTPRequestData)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, result)
	return n.ExecuteChildren(ctx)
}
