package flow

import (
	"fmt"
	"strconv"

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
	case FlowNodeTypeActionResponseEdit:
		interaction := ctx.Data.Interaction()

		// TODO: this should figure if it's a follow-up or not

		data := n.Data.MessageData

		content, err := ctx.Variables.ParseAndExecute(data.Content)
		if err != nil {
			return traceError(n, err)
		}

		resp := api.EditInteractionResponseData{
			Content: option.NewNullableString(content),
			Embeds:  &data.Embeds,
			// TODO: other fields
		}

		err = ctx.Discord.EditInteractionResponse(ctx, interaction.AppID, interaction.Token, resp)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionResponseDelete:
		interaction := ctx.Data.Interaction()

		err := ctx.Discord.DeleteInteractionResponse(ctx, interaction.AppID, interaction.Token)
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
	case FlowNodeTypeActionMessageEdit:
		messageTarget, err := ctx.Variables.ParseAndExecute(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageID, _ := strconv.ParseInt(messageTarget, 10, 64)

		msg, err := ctx.Discord.EditMessage(
			ctx,
			ctx.Data.ChannelID(),
			discord.MessageID(messageID),
			api.EditMessageData{
				Content: option.NewNullableString(n.Data.MessageData.Content),
				Embeds:  &n.Data.MessageData.Embeds,
			},
		)
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
	case FlowNodeTypeActionMessageDelete:
		messageTarget, err := ctx.Variables.ParseAndExecute(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageID, _ := strconv.ParseInt(messageTarget, 10, 64)

		err = ctx.Discord.DeleteMessage(
			ctx,
			ctx.Data.ChannelID(),
			discord.MessageID(messageID),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionMemberBan:
		memberTarget, err := ctx.Variables.ParseAndExecute(n.Data.MemberTarget)
		if err != nil {
			return traceError(n, err)
		}

		memberID, _ := strconv.ParseInt(memberTarget, 10, 64)

		rawDeleteSeconds, err := ctx.Variables.ParseAndExecute(n.Data.MemberBanDeleteMessageDuration)
		if err != nil {
			return traceError(n, err)
		}

		deleteSeconds, _ := strconv.Atoi(rawDeleteSeconds)

		err = ctx.Discord.BanMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(memberID),
			api.BanData{
				DeleteDays:     option.NewUint(uint(deleteSeconds / 86400)),
				AuditLogReason: api.AuditLogReason(n.Data.AuditLogReason),
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionMemberKick:
		memberTarget, err := ctx.Variables.ParseAndExecute(n.Data.MemberTarget)
		if err != nil {
			return traceError(n, err)
		}

		memberID, _ := strconv.ParseInt(memberTarget, 10, 64)

		err = ctx.Discord.KickMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(memberID),
			n.Data.AuditLogReason,
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)

	// TODO: implement other action types

	case FlowNodeTypeActionLog:
		ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, n.Data.LogMessage)
		return n.executeChildren(ctx)
	case FlowNodeTypeControlConditionCompare:
		if err := n.Data.ConditionBaseValue.ResolveVariables(ctx.Variables); err != nil {
			return traceError(n, err)
		}

		ctx.Tempories.InitCondition(n.Data.ConditionBaseValue, n.Data.ConditionAllowMultiple)

		var elseNode *CompiledFlowNode

		for _, child := range n.Children {
			if child.Type == FlowNodeTypeControlConditionItemElse {
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
	case FlowNodeTypeControlConditionItemCompare:
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
	case FlowNodeTypeControlConditionItemElse:
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
