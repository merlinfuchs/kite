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

	// Make node data available to placeholders
	ctx.nodePlaceholderProvider.setNode(n)

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

		content, err := ctx.Placeholders.Fill(ctx, data.Content)
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

		content, err := ctx.Placeholders.Fill(ctx, data.Content)
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

		n.State.Result = NewFlowValueMessage(*msg)
		return n.executeChildren(ctx)
	case FlowNodeTypeActionMessageEdit:
		if err := n.Data.MessageTarget.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		messageID := n.Data.MessageTarget.Number()

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

		n.State.Result = NewFlowValueMessage(*msg)
		return n.executeChildren(ctx)
	case FlowNodeTypeActionMessageDelete:
		if err := n.Data.MessageTarget.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		messageID := n.Data.MessageTarget.Number()

		err := ctx.Discord.DeleteMessage(
			ctx,
			ctx.Data.ChannelID(),
			discord.MessageID(messageID),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionMemberBan:
		if err := n.Data.MemberTarget.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		memberID := n.Data.MemberTarget.Number()

		if err := n.Data.MemberBanDeleteMessageDuration.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		deleteSeconds := n.Data.MemberBanDeleteMessageDuration.Number()

		if err := n.Data.AuditLogReason.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		err := ctx.Discord.BanMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(memberID),
			api.BanData{
				DeleteDays:     option.NewUint(uint(deleteSeconds / 86400)),
				AuditLogReason: api.AuditLogReason(n.Data.AuditLogReason.String()),
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionMemberKick:
		if err := n.Data.AuditLogReason.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		memberID := n.Data.MemberTarget.Number()

		if err := n.Data.AuditLogReason.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		err := ctx.Discord.KickMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(memberID),
			n.Data.AuditLogReason.String(),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)

	// TODO: implement other action types

	case FlowNodeTypeActionLog:
		err := n.Data.LogMessage.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, n.Data.LogMessage.String())
		return n.executeChildren(ctx)
	case FlowNodeTypeControlConditionCompare:
		if err := n.Data.ConditionBaseValue.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

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
		parent := n.FindDirectParentWithType(FlowNodeTypeControlConditionCompare)
		if parent == nil {
			return nil
		}

		if parent.State.ConditionItemMet && parent.Data.ConditionAllowMultiple {
			// Another condition item has already been met
			return nil
		}

		if err := n.Data.ConditionItemValue.FillPlaceholders(ctx, ctx.Placeholders); err != nil {
			return traceError(n, err)
		}

		baseValue := parent.Data.ConditionBaseValue

		var conditionMet bool
		switch n.Data.ConditionItemMode {
		case ConditionItemModeEqual:
			conditionMet = baseValue.Equals(&n.Data.ConditionItemValue)
		case ConditionItemModeNotEqual:
			conditionMet = baseValue.Equals(&n.Data.ConditionItemValue)
		case ConditionItemModeGreaterThan:
			conditionMet = baseValue.GreaterThan(&n.Data.ConditionItemValue)
		case ConditionItemModeGreaterThanOrEqual:
			conditionMet = baseValue.GreaterThanOrEqual(&n.Data.ConditionItemValue)
		case ConditionItemModeLessThan:
			conditionMet = baseValue.LessThan(&n.Data.ConditionItemValue)
		case ConditionItemModeLessThanOrEqual:
			conditionMet = baseValue.LessThanOrEqual(&n.Data.ConditionItemValue)
		case ConditionItemModeContains:
			conditionMet = baseValue.Contains(&n.Data.ConditionItemValue)
		}

		if conditionMet {
			parent.State.ConditionItemMet = true
			return n.executeChildren(ctx)
		}
	case FlowNodeTypeControlConditionItemElse:
		parent := n.FindDirectParentWithType(FlowNodeTypeControlConditionCompare)
		if parent == nil {
			return nil
		}

		if parent.State.ConditionItemMet {
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
