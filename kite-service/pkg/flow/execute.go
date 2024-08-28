package flow

import (
	"fmt"
	"time"

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

	nodeState := ctx.GetNodeState(n.ID)

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

		nodeState.Result = NewFlowValueMessage(*msg)
		return n.executeChildren(ctx)
	case FlowNodeTypeActionMessageEdit:
		channelTarget, err := n.Data.ChannelTarget.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := n.Data.MessageTarget.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		msg, err := ctx.Discord.EditMessage(
			ctx,
			discord.ChannelID(channelTarget.Number()),
			discord.MessageID(messageTarget.Number()),
			api.EditMessageData{
				Content: option.NewNullableString(n.Data.MessageData.Content),
				Embeds:  &n.Data.MessageData.Embeds,
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		nodeState.Result = NewFlowValueMessage(*msg)
		return n.executeChildren(ctx)
	case FlowNodeTypeActionMessageDelete:
		channelTarget, err := n.Data.ChannelTarget.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := n.Data.MessageTarget.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.DeleteMessage(
			ctx,
			discord.ChannelID(channelTarget.Number()),
			discord.MessageID(messageTarget.Number()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionMemberBan:
		memberTarget, err := n.Data.MemberTarget.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		messageDeleteSeconds, err := n.Data.MemberBanDeleteMessageDurationSeconds.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := n.Data.AuditLogReason.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.BanMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(memberTarget.Number()),
			api.BanData{
				DeleteDays:     option.NewUint(uint(messageDeleteSeconds.Number() / 86400)),
				AuditLogReason: api.AuditLogReason(auditLogReason),
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeActionMemberKick:
		memberID, err := n.Data.MemberTarget.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := n.Data.AuditLogReason.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.KickMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(memberID.Number()),
			auditLogReason.String(),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.executeChildren(ctx)

	// TODO: implement other action types

	case FlowNodeTypeActionLog:
		logMessage, err := n.Data.LogMessage.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, logMessage.String())
		return n.executeChildren(ctx)
	case FlowNodeTypeControlConditionCompare:
		fmt.Println("condition compare", n.Data.ConditionBaseValue)
		baseValue, err := n.Data.ConditionBaseValue.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}
		fmt.Println("condition compare 2", baseValue)

		nodeState.ConditionBaseValue = baseValue

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

		parentState := ctx.GetNodeState(parent.ID)

		if parentState.ConditionItemMet && parent.Data.ConditionAllowMultiple {
			// Another condition item has already been met
			return nil
		}

		itemValue, err := n.Data.ConditionItemValue.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		baseValue := parentState.ConditionBaseValue

		fmt.Println("base value: '" + baseValue + "'")
		fmt.Println("item value: '" + itemValue + "'")

		var conditionMet bool
		switch n.Data.ConditionItemMode {
		case ConditionItemModeEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ConditionItemModeNotEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ConditionItemModeGreaterThan:
			conditionMet = baseValue.GreaterThan(&itemValue)
		case ConditionItemModeGreaterThanOrEqual:
			conditionMet = baseValue.GreaterThanOrEqual(&itemValue)
		case ConditionItemModeLessThan:
			conditionMet = baseValue.LessThan(&itemValue)
		case ConditionItemModeLessThanOrEqual:
			conditionMet = baseValue.LessThanOrEqual(&itemValue)
		case ConditionItemModeContains:
			conditionMet = baseValue.Contains(&itemValue)
		}

		if conditionMet {
			parentState.ConditionItemMet = true
			return n.executeChildren(ctx)
		}
	case FlowNodeTypeControlConditionItemElse:
		parent := n.FindDirectParentWithType(FlowNodeTypeControlConditionCompare)
		if parent == nil {
			return nil
		}

		parentState := ctx.GetNodeState(parent.ID)

		if parentState.ConditionItemMet {
			// Another condition item has already been met
			return nil
		}

		return n.executeChildren(ctx)
	case FlowNodeTypeControlSleep:
		sleepSeconds, err := n.Data.SleepDurationSeconds.FillPlaceholders(ctx, ctx.Placeholders)
		if err != nil {
			return traceError(n, err)
		}

		duration := time.Duration(sleepSeconds.Number()) * time.Second

		deadline, ok := ctx.Deadline()
		if ok && time.Now().Add(duration).After(deadline) {
			return &FlowError{
				Code:    FlowNodeErrorTimeout,
				Message: "sleep would exceed deadline",
			}
		}

		time.Sleep(duration)
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
