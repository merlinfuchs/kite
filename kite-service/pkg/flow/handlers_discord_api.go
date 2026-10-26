package flow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
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

	var interaction *discord.InteractionEvent
	if op.Interaction {
		interaction = ctx.Data.Interaction()
		if interaction == nil {
			return traceError(n, fmt.Errorf("the %s endpoint only works in flows started by an interaction, like a command or button click", op.ID))
		}
	}

	pathParams, err := ctx.evalKeyValues(data.PathParams)
	if err != nil {
		return traceError(n, err)
	}
	query, err := ctx.evalKeyValues(data.Query)
	if err != nil {
		return traceError(n, err)
	}

	path, err := discordAPIPath(op, pathParams, query, interactionPathParams(interaction))
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

	if interaction != nil {
		if err := deferForInteractionRequest(ctx, interaction, body); err != nil {
			return traceError(n, err)
		}
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

// interactionPathParams are the path parameters Kite fills in for endpoints of
// the interaction.
func interactionPathParams(interaction *discord.InteractionEvent) map[string]string {
	if interaction == nil {
		return nil
	}
	return map[string]string{
		"application_id":    interaction.AppID.String(),
		"interaction_token": interaction.Token,
	}
}

// deferForInteractionRequest responds to the interaction if nothing has yet, as
// Discord only accepts followups and edits once it has a response. The first
// one replaces the "thinking…" message and keeps its flags, so the defer is
// ephemeral if the request is.
func deferForInteractionRequest(ctx *FlowContext, interaction *discord.InteractionEvent, body []byte) error {
	resp := autoDeferResponse(interaction, nil)
	if resp.Data != nil && ephemeralBody(body) {
		resp.Data.Flags |= discord.EphemeralMessage
	}
	return deferInteraction(ctx, interaction, resp)
}

// ephemeralBody reports whether a message body has the ephemeral flag, given
// as a number or a string like a placeholder's result.
func ephemeralBody(body []byte) bool {
	var msg struct {
		Flags json.RawMessage `json:"flags"`
	}
	if json.Unmarshal(body, &msg) != nil {
		return false
	}
	flags, err := strconv.ParseUint(strings.Trim(string(msg.Flags), `"`), 10, 64)
	return err == nil && discord.MessageFlags(flags)&discord.EphemeralMessage != 0
}
