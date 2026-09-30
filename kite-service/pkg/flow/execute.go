package flow

import (
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

func (n *CompiledFlowNode) Execute(ctx *FlowContext) error {
	if n == nil {
		// TODO: Figure out why nodes are some times nil, this is probably a bug in the compiler?
		return fmt.Errorf("node is nil")
	}

	if err := ctx.startOperation(n.CreditsCost()); err != nil {
		return traceError(n, err)
	}
	defer ctx.endOperation()

	if handler, ok := nodeHandlers[n.Type]; ok {
		if err := n.checkIntegrations(ctx); err != nil {
			return traceError(n, err)
		}
		return handler(n, ctx)
	}

	if block, ok := blockDefinitions[n.Type]; ok && block.Run.Kind == "request" {
		return n.executeBlockDefinition(ctx, block)
	}

	return &FlowError{
		Code:    FlowNodeErrorUnknownNodeType,
		Message: fmt.Sprintf("unknown node type: %s", n.Type),
	}
}

// nodeHandlers runs the blocks written in Go. Blocks defined as data run from
// blockDefinitions instead. The handlers_*.go files register their blocks in
// init, as the handlers refer back to the map through Execute.
var nodeHandlers = map[FlowNodeType]nodeHandler{}

type nodeHandler func(*CompiledFlowNode, *FlowContext) error

func registerHandlers(handlers map[FlowNodeType]nodeHandler) {
	for nodeType, handler := range handlers {
		if _, ok := nodeHandlers[nodeType]; ok {
			panic(fmt.Sprintf("handler for %s registered twice", nodeType))
		}
		nodeHandlers[nodeType] = handler
	}
}

func (n *CompiledFlowNode) CreditsCost() int {
	if block, ok := blockDefinitions[n.Type]; ok && block.Run.Kind == "request" {
		return *block.Credits
	}

	switch n.Type {
	case FlowNodeTypeActionAIChatCompletion, FlowNodeTypeActionAISearchWeb:
		data := n.Data.AIChatCompletionData
		if data == nil {
			return 0
		}

		return AICreditsCost(data.Model, n.Type == FlowNodeTypeActionAISearchWeb)
	case FlowNodeTypeActionHTTPRequest:
		return 3
	}

	if n.IsAction() {
		return 1
	}

	return 0
}

func (n *CompiledFlowNode) ExecuteChildrenByHandle(ctx *FlowContext, handle string) error {
	children, ok := n.Children.Handles[handle]
	if !ok {
		return nil
	}

	for _, child := range children {
		// We could spawn a goroutine here to execute children in parallel
		// but we'll just execute them sequentially for now
		if err := child.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}

	return nil
}

func (n *CompiledFlowNode) ExecuteChildren(ctx *FlowContext) error {
	for _, child := range n.Children.Default {
		// We could spawn a goroutine here to execute children in parallel
		// but we'll just execute them sequentially for now
		if err := child.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}
	return nil
}

func (n *CompiledFlowNode) autoDeferInteraction(ctx *FlowContext) error {
	return autoDeferInteraction(ctx, n.FirstChildMatching(isResponseNode))
}

// autoDeferInteraction defers the interaction if the flow doesn't respond in
// time, guessing the kind of defer from responseNode, the first response the
// flow can reach.
//
// This can't be right for every flow — branches may disagree, and only one of
// them runs. Users who need certainty should defer explicitly instead.
func autoDeferInteraction(ctx *FlowContext, responseNode *CompiledFlowNode) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	go ctx.Discord.AutoDeferInteraction(ctx, interaction.ID, interaction.Token, autoDeferResponse(interaction, responseNode))
	return nil
}

func autoDeferResponse(interaction *discord.InteractionEvent, responseNode *CompiledFlowNode) api.InteractionResponse {
	// A component interaction that doesn't create a new response either edits
	// the message the component is on or doesn't respond at all. Both only
	// need an acknowledgement, and a "thinking…" message would be edited
	// instead of the component's message.
	if _, ok := interaction.Data.(discord.ComponentInteraction); ok {
		if responseNode == nil || (responseNode.Type != FlowNodeTypeActionResponseCreate && responseNode.Type != FlowNodeTypeActionResponseDefer) {
			return api.InteractionResponse{Type: api.DeferredMessageUpdate}
		}
	}

	// The first response after this defer replaces the "thinking…" message and
	// keeps the defer's flags, so ephemeral-ness has to be decided now.
	resp := api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{},
	}
	if responseNode != nil && responseNode.Data.MessageEphemeral {
		resp.Data.Flags |= discord.EphemeralMessage
	}
	return resp
}

// deferUnanswered defers the interaction the flow runs with, unless something
// already responded to it.
func (n *CompiledFlowNode) deferUnanswered(ctx *FlowContext) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return nil
	}

	if responded, err := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID); err != nil || responded {
		return err
	}

	resp := autoDeferResponse(interaction, FirstMatching(n.Children.Default, isResponseNode))
	_, err := ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp)
	if err != nil {
		// The entry's auto defer can respond at the same time.
		if responded, _ := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID); responded {
			return nil
		}
	}
	return err
}

// targetGuildID returns the guild a block acts on: the guild target if one is
// set, otherwise the guild of the interaction or event that triggered the flow.
func (n *CompiledFlowNode) targetGuildID(ctx *FlowContext) (discord.GuildID, error) {
	if n.Data.GuildTarget == "" {
		return ctx.Data.GuildID(), nil
	}

	guildTarget, err := ctx.EvalTemplate(n.Data.GuildTarget)
	if err != nil {
		return 0, err
	}

	return discord.GuildID(guildTarget.Snowflake()), nil
}

func (n *CompiledFlowNode) resumeFromComponent(ctx *FlowContext) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	data, ok := interaction.Data.(discord.ComponentInteraction)
	if !ok {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is not a component interaction",
		}
	}

	_, compID, ok := message.DecodeCustomIDMessageComponentResumePoint(string(data.ID()))
	if !ok {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "invalid custom ID",
		}
	}

	// Only the clicked component's branch runs, so guess the defer from it
	// rather than from the node's other children.
	handle := fmt.Sprintf("component_%d", compID)
	err := autoDeferInteraction(ctx, FirstMatching(n.Children.Handles[handle], isResponseNode))
	if err != nil {
		return traceError(n, err)
	}

	err = n.ExecuteChildrenByHandle(ctx, handle)
	if err != nil {
		createDefaultErrorResponse(ctx, err)
		return traceError(n, err)
	}

	acknowledgeUnansweredComponent(ctx)
	return nil
}

// acknowledgeUnansweredComponent acknowledges a component interaction the flow
// finished without responding to, e.g. a button that only sends a channel
// message. Discord shows "This interaction failed" otherwise, and the
// auto-defer doesn't fire for flows that finish quickly.
func acknowledgeUnansweredComponent(ctx *FlowContext) {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return
	}
	if _, ok := interaction.Data.(discord.ComponentInteraction); !ok {
		return
	}

	hasCreatedResponse, err := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
	if err != nil || hasCreatedResponse {
		return
	}

	_, _ = ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, api.InteractionResponse{
		Type: api.DeferredMessageUpdate,
	})
}

func createDefaultErrorResponse(fCtx *FlowContext, err error) {
	interaction := fCtx.Data.Interaction()
	if interaction == nil {
		return
	}

	ctx, cancel := context.WithTimeout(fCtx, time.Second*10)
	defer cancel()

	errStr := err.Error()
	if len(errStr) > 1800 {
		errStr = errStr[:1800]
	}

	respData := api.InteractionResponseData{
		Content: option.NewNullableString("An error occurred while executing the flow event: ```" + errStr + "```"),
		Flags:   discord.EphemeralMessage,
	}

	hasCreatedResponse, _ := fCtx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
	if hasCreatedResponse {
		_, _ = fCtx.Discord.CreateInteractionFollowup(ctx, interaction.AppID, interaction.Token, respData)
	} else {
		_, _ = fCtx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &respData,
		})
	}
}

// isResponseNode reports whether a node responds to the interaction, and so
// determines whether the deferred response is ephemeral.
func isResponseNode(n *CompiledFlowNode) bool {
	switch n.Type {
	case FlowNodeTypeActionResponseCreate,
		FlowNodeTypeActionResponseEdit,
		FlowNodeTypeActionResponseDefer:
		return true
	default:
		return false
	}
}
