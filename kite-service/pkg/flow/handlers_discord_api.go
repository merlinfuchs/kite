package flow

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionDiscordAPIRequest: executeActionDiscordAPIRequest,
	})
}

func executeActionDiscordAPIRequest(n *CompiledFlowNode, ctx *FlowContext) error {
	data := n.Data.DiscordAPIRequestData
	if data == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "discord_api_request_data is nil",
		}
	}

	op, ok := discordAPIOperations[data.Operation]
	if !ok {
		return traceError(n, fmt.Errorf("unknown Discord API endpoint: %s", data.Operation))
	}

	pathParams, err := ctx.evalKeyValues(data.PathParams)
	if err != nil {
		return traceError(n, err)
	}
	query, err := ctx.evalKeyValues(data.Query)
	if err != nil {
		return traceError(n, err)
	}

	path, err := discordAPIPath(op, pathParams, query)
	if err != nil {
		return traceError(n, err)
	}

	var body []byte
	if len(data.BodyJSON) > 0 && string(data.BodyJSON) != "null" {
		if !op.HasBody {
			return traceError(n, fmt.Errorf("the %s endpoint doesn't take a body", op.ID))
		}

		body, err = ctx.EvalJSONTemplate(data.BodyJSON)
		if err != nil {
			return traceError(n, err)
		}
	}

	auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
	if err != nil {
		return traceError(n, err)
	}

	resBody, err := ctx.Discord.APIRequest(ctx, provider.DiscordAPIRequest{
		Method: op.Method,
		Path:   path,
		Body:   body,
		Reason: api.AuditLogReason(auditLogReason.String()),
	})
	if err != nil {
		return traceError(n, err)
	}

	result, err := discordAPIResult(resBody)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, result)
	return n.ExecuteChildren(ctx)
}
