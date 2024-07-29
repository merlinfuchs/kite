package flow

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func (n *CompiledFlowNode) Execute(ctx *FlowContext) error {
	if err := ctx.startOperation(); err != nil {
		return traceError(n, err)
	}
	defer ctx.endOperation()

	if n.IsAction() {
		if err := ctx.startAction(); err != nil {
			return traceError(n, err)
		}
	}

	switch n.Type {
	case FlowNodeTypeEntryCommand:
		return n.executeChildren(ctx)
	case FlowNodeTypeEntryEvent:
		return n.executeChildren(ctx)
	case FlowNodeTypeActionResponseCreate:
		interaction := ctx.Data.Interaction()
		if interaction == nil {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "interaction is nil",
			}
		}

		data := n.Data.MessageData

		var flags discord.MessageFlags
		if n.Data.MessageEphemeral {
			flags |= discord.EphemeralMessage
		}

		content, err := ctx.Variables.ParseAndExecute(data.Content)
		if err != nil {
			return traceError(n, err)
		}

		resp := api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString(content),
				Embeds:  &data.Embeds,
				Flags:   flags,
				// TODO: other fields
			},
		}

		err = ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionMessageCreate:
		msg, err := ctx.Discord.CreateMessage(ctx, ctx.Data.ChannelID(), n.Data.MessageData)
		if err != nil {
			return traceError(n, err)
		}

		if n.Data.ResultVariableName != "" {
			ctx.Variables.SetVariable(n.Data.ResultVariableName, FlowValue{
				Type:  FlowValueTypeMessage,
				Value: *msg,
			})
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionLog:
		ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, n.Data.LogMessage)
		return n.executeChildren(ctx)
	case FlowNodeTypeConditionCompare:
		if err := n.Data.ConditionBaseValue.ResolveVariables(ctx.Variables); err != nil {
			return traceError(n, err)
		}

		ctx.Tempories.InitCondition(n.Data.ConditionBaseValue, n.Data.ConditionAllowMultiple)

		var elseNode *CompiledFlowNode

		for _, child := range n.Children {
			if child.Type == FlowNodeTypeConditionItemElse {
				elseNode = child
			} else {
				if err := child.Execute(ctx); err != nil {
					return traceError(n, err)
				}
			}
		}

		if elseNode != nil {
			// else node has to be executed last
			if err := elseNode.Execute(ctx); err != nil {
				return traceError(n, err)
			}
		}
	case FlowNodeTypeConditionItemCompare:
		if ctx.Tempories.ConditionItemMet && !ctx.Tempories.ConditionAllowMultiple {
			// Another condition item has already been met
			return nil
		}

		if err := n.Data.ConditionItemValue.ResolveVariables(ctx.Variables); err != nil {
			return traceError(n, err)
		}

		var conditionMet bool
		switch n.Data.ConditionItemMode {
		case ConditionItemModeEqual:
			conditionMet = ctx.Tempories.ConditionBaseValue.Equals(&n.Data.ConditionItemValue)
		case ConditionItemModeNotEqual:
			conditionMet = ctx.Tempories.ConditionBaseValue.Equals(&n.Data.ConditionItemValue)
		case ConditionItemModeGreaterThan:
			conditionMet = ctx.Tempories.ConditionBaseValue.GreaterThan(&n.Data.ConditionItemValue)
		case ConditionItemModeGreaterThanOrEqual:
			conditionMet = ctx.Tempories.ConditionBaseValue.GreaterThanOrEqual(&n.Data.ConditionItemValue)
		case ConditionItemModeLessThan:
			conditionMet = ctx.Tempories.ConditionBaseValue.LessThan(&n.Data.ConditionItemValue)
		case ConditionItemModeLessThanOrEqual:
			conditionMet = ctx.Tempories.ConditionBaseValue.LessThanOrEqual(&n.Data.ConditionItemValue)
		case ConditionItemModeContains:
			conditionMet = ctx.Tempories.ConditionBaseValue.Contains(&n.Data.ConditionItemValue)
		}

		if conditionMet {
			ctx.Tempories.ConditionItemMet = true
			return n.executeChildren(ctx)
		}
	case FlowNodeTypeConditionItemElse:
		if ctx.Tempories.ConditionItemMet {
			// Another condition item has already been met
			return nil
		}

		return n.executeChildren(ctx)
	default:
		return &FlowError{
			Code:    FlowNodeErrorUnknownNodeType,
			Message: fmt.Sprintf("unknown node type: %s", n.Type),
		}
	}

	return nil
}

func (n *CompiledFlowNode) executeChildren(ctx *FlowContext) error {
	for _, child := range n.Children {
		if err := child.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}
	return nil
}
