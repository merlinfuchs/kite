package flow

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionResponseCreate: executeActionResponseCreate,
		FlowNodeTypeActionResponseEdit:   executeActionResponseEdit,
		FlowNodeTypeActionResponseDelete: executeActionResponseDelete,
		FlowNodeTypeActionResponseDefer:  executeActionResponseDefer,
		FlowNodeTypeSuspendResponseModal: executeSuspendResponseModal,
	})
}

func executeActionResponseCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.IsEntry() {
		return n.resumeFromComponent(ctx)
	}

	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	data, opts, resumePointID, err := n.prepareMessage(ctx)
	if err != nil {
		return traceError(n, err)
	}
	responseData := data.ToInteractionResponseData(opts)

	hasCreatedResponse, err := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
	if err != nil {
		return traceError(n, err)
	}

	var msg *discord.Message
	if hasCreatedResponse {
		msg, err = ctx.Discord.CreateInteractionFollowup(ctx, interaction.AppID, interaction.Token, responseData)
		if err != nil {
			return traceError(n, err)
		}
	} else {
		resp := api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &responseData,
		}

		res, err := ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp)
		if err != nil {
			return traceError(n, err)
		}

		if res != nil {
			msg = res.Message
		}
	}

	if msg != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordMessage(*msg))
	}
	if resumePointID != "" {
		// We have to create the resume point after the message has been stored
		_, err = ctx.suspend(ResumePointTypeMessageComponents, resumePointID, n.ID)
		if err != nil {
			return traceError(n, err)
		}
	}

	if n.Data.MessageTemplateID != "" && msg != nil {
		err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, provider.MessageTemplateInstance{
			MessageTemplateID: n.Data.MessageTemplateID,
			MessageID:         msg.ID,
			ChannelID:         msg.ChannelID,
			GuildID:           ctx.Data.GuildID(),
			Ephemeral:         n.Data.MessageEphemeral,
		})
		if err != nil {
			ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, fmt.Sprintf("failed to link message template instance: %s", err.Error()))
		}
	}

	return n.ExecuteChildren(ctx)
}

func executeActionResponseEdit(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.IsEntry() {
		return n.resumeFromComponent(ctx)
	}

	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	data, opts, resumePointID, err := n.prepareMessage(ctx)
	if err != nil {
		return traceError(n, err)
	}

	var msg *discord.Message
	if n.Data.MessageTarget == "" || n.Data.MessageTarget == "@original" {
		hasCreatedResponse, err := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
		if err != nil {
			return traceError(n, err)
		}

		if hasCreatedResponse {
			msg, err = ctx.Discord.EditInteractionResponse(ctx, interaction.AppID, interaction.Token, data.ToEditInteractionResponseData(opts))
			if err != nil {
				return traceError(n, err)
			}
		} else {
			responseData := data.ToInteractionResponseData(opts)
			resp := api.InteractionResponse{
				Type: api.UpdateMessage,
				Data: &responseData,
			}

			res, err := ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp)
			if err != nil {
				return traceError(n, err)
			}

			if res != nil {
				msg = res.Message
			}
		}
	} else {
		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		msg, err = ctx.Discord.EditInteractionFollowup(
			ctx,
			interaction.AppID,
			interaction.Token,
			discord.MessageID(messageTarget.Snowflake()),
			data.ToEditInteractionResponseData(opts),
		)
		if err != nil {
			return traceError(n, err)
		}
	}

	if msg != nil {
		ctx.StoreNodeResult(n, thing.NewDiscordMessage(*msg))
	}
	if resumePointID != "" {
		// We have to create the resume point after the message has been stored
		_, err = ctx.suspend(ResumePointTypeMessageComponents, resumePointID, n.ID)
		if err != nil {
			return traceError(n, err)
		}
	}

	if n.Data.MessageTemplateID != "" && msg != nil {
		err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, provider.MessageTemplateInstance{
			MessageTemplateID: n.Data.MessageTemplateID,
			MessageID:         msg.ID,
			ChannelID:         msg.ChannelID,
			GuildID:           ctx.Data.GuildID(),
			Ephemeral:         n.Data.MessageEphemeral,
		})
		if err != nil {
			ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, fmt.Sprintf("failed to link message template instance: %s", err.Error()))
		}
	}

	return n.ExecuteChildren(ctx)
}

func executeActionResponseDelete(n *CompiledFlowNode, ctx *FlowContext) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	if n.Data.MessageTarget == "" || n.Data.MessageTarget == "@original" {
		err := ctx.Discord.DeleteInteractionResponse(ctx, interaction.AppID, interaction.Token)
		if err != nil {
			return traceError(n, err)
		}
	} else {
		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.DeleteInteractionFollowup(
			ctx,
			interaction.AppID,
			interaction.Token,
			discord.MessageID(messageTarget.Snowflake()),
		)
		if err != nil {
			return traceError(n, err)
		}
	}

	return n.ExecuteChildren(ctx)
}

func executeActionResponseDefer(n *CompiledFlowNode, ctx *FlowContext) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	resp := api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{},
	}

	if n.Data.MessageEphemeral {
		resp.Data.Flags |= discord.EphemeralMessage
	}

	_, err := ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp)
	if err != nil {
		return traceError(n, err)
	}

	return n.ExecuteChildren(ctx)
}

func executeSuspendResponseModal(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.IsEntry() {
		err := n.autoDeferInteraction(ctx)
		if err != nil {
			return traceError(n, err)
		}

		err = n.ExecuteChildren(ctx)
		if err != nil {
			createDefaultErrorResponse(ctx, err)
			return traceError(n, err)
		}

		return nil
	}

	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	if n.Data.ModalData == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "modal data is nil",
		}
	}

	// The custom IDs aren't evaluated, as input() looks values up by them.
	title, err := ctx.EvalTemplate(n.Data.ModalData.Title)
	if err != nil {
		return traceError(n, err)
	}

	componentRows := make(discord.TopLevelComponents, len(n.Data.ModalData.Components))
	for i, row := range n.Data.ModalData.Components {
		r := make(discord.ActionRowComponent, len(row.Components))
		for j, component := range row.Components {
			label, err := ctx.EvalTemplate(component.Label)
			if err != nil {
				return traceError(n, err)
			}

			placeholder, err := ctx.EvalTemplate(component.Placeholder)
			if err != nil {
				return traceError(n, err)
			}

			value, err := ctx.EvalTemplateKeepSpace(component.Value)
			if err != nil {
				return traceError(n, err)
			}

			r[j] = &discord.TextInputComponent{
				CustomID:     discord.ComponentID(component.CustomID),
				Label:        label.String(),
				Style:        discord.TextInputStyle(component.Style),
				Required:     component.Required,
				LengthLimits: [2]int{component.MinLength, component.MaxLength},
				Value:        value.String(),
				Placeholder:  placeholder.String(),
			}
		}

		componentRows[i] = discord.TopLevelComponent(&r)
	}

	// Suspend only once the modal is ready, so a failed template doesn't
	// leave a resume point behind.
	resumePoint, err := ctx.suspend(ResumePointTypeModal, util.UniqueID(), n.ID)
	if err != nil {
		return traceError(n, fmt.Errorf("failed to suspend: %w", err))
	}

	resp := api.InteractionResponse{
		Type: api.ModalResponse,
		Data: &api.InteractionResponseData{
			CustomID:   option.NewNullableString(message.CustomIDModalResumePoint(resumePoint.ID)),
			Title:      option.NewNullableString(title.String()),
			Components: &componentRows,
		},
	}

	_, err = ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp)
	if err != nil {
		return traceError(n, err)
	}

	return traceError(n, nil)
}
