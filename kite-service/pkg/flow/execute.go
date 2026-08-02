package flow

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"slices"
	"time"

	"github.com/openai/openai-go"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
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

	switch n.Type {
	case FlowNodeTypeEntryCommand, FlowNodeTypeEntryComponentButton:
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

		return nil
	case FlowNodeTypeEntryEvent:
		if !ctx.IsEntry() {
			return fmt.Errorf("event entry isn't the entry node")
		}

		err := n.ExecuteChildren(ctx)
		if err != nil {
			return traceError(n, err)
		}

		return nil
	case FlowNodeTypeActionResponseCreate:
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

		responseData, resumePointID, err := n.prepareMessageResponseData(ctx)
		if err != nil {
			return traceError(n, err)
		}

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
				ChannelID:         ctx.Data.ChannelID(),
				GuildID:           ctx.Data.GuildID(),
				Ephemeral:         n.Data.MessageEphemeral,
			})
			if err != nil {
				ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, fmt.Sprintf("failed to link message template instance: %s", err.Error()))
			}
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionResponseEdit:
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

		responseData, resumePointID, err := n.prepareMessageResponseData(ctx)
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
				msg, err = ctx.Discord.EditInteractionResponse(ctx, interaction.AppID, interaction.Token, api.EditInteractionResponseData{
					Content:    responseData.Content,
					Embeds:     responseData.Embeds,
					Components: responseData.Components,
				})
				if err != nil {
					return traceError(n, err)
				}
			} else {
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
				api.EditInteractionResponseData{
					Content: responseData.Content,
					Embeds:  responseData.Embeds,
				},
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

		if n.Data.MessageTemplateID != "" {
			err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, provider.MessageTemplateInstance{
				MessageTemplateID: n.Data.MessageTemplateID,
				MessageID:         msg.ID,
				ChannelID:         ctx.Data.ChannelID(),
				GuildID:           ctx.Data.GuildID(),
				Ephemeral:         n.Data.MessageEphemeral,
			})
			if err != nil {
				ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, fmt.Sprintf("failed to link message template instance: %s", err.Error()))
			}
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionResponseDelete:
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
	case FlowNodeTypeActionResponseDefer:
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
	case FlowNodeTypeSuspendResponseModal:
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

		resumePoint, err := ctx.suspend(ResumePointTypeModal, util.UniqueID(), n.ID)
		if err != nil {
			return traceError(n, fmt.Errorf("failed to suspend: %w", err))
		}

		componentRows := make(discord.ContainerComponents, len(n.Data.ModalData.Components))
		for i, row := range n.Data.ModalData.Components {
			r := make(discord.ActionRowComponent, len(row.Components))
			for j, component := range row.Components {
				r[j] = &discord.TextInputComponent{
					CustomID:     discord.ComponentID(component.CustomID),
					Label:        component.Label,
					Style:        discord.TextInputStyle(component.Style),
					Required:     component.Required,
					LengthLimits: [2]int{component.MinLength, component.MaxLength},
					Value:        component.Value,
					Placeholder:  component.Placeholder,
				}
			}

			componentRows[i] = discord.ContainerComponent(&r)
		}

		resp := api.InteractionResponse{
			Type: api.ModalResponse,
			Data: &api.InteractionResponseData{
				CustomID:   option.NewNullableString(message.CustomIDModalResumePoint(resumePoint.ID)),
				Title:      option.NewNullableString(n.Data.ModalData.Title),
				Components: &componentRows,
			},
		}

		_, err = ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp)
		if err != nil {
			return traceError(n, err)
		}

		return traceError(n, nil)
	case FlowNodeTypeActionMessageCreate:
		if ctx.IsEntry() {
			return n.resumeFromComponent(ctx)
		}

		messageData, resumePointID, err := n.prepareMessageSendData(ctx)
		if err != nil {
			return traceError(n, err)
		}

		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		msg, err := ctx.Discord.CreateMessage(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			messageData,
		)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, thing.NewDiscordMessage(*msg))
		if resumePointID != "" {
			// We have to create the resume point after the message has been stored
			_, err = ctx.suspend(ResumePointTypeMessageComponents, resumePointID, n.ID)
			if err != nil {
				return traceError(n, err)
			}
		}

		if n.Data.MessageTemplateID != "" {
			err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, provider.MessageTemplateInstance{
				MessageTemplateID: n.Data.MessageTemplateID,
				MessageID:         msg.ID,
				ChannelID:         ctx.Data.ChannelID(),
				GuildID:           ctx.Data.GuildID(),
				Ephemeral:         n.Data.MessageEphemeral,
			})
			if err != nil {
				ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, fmt.Sprintf("failed to link message template instance: %s", err.Error()))
			}
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMessageEdit:
		if ctx.IsEntry() {
			return n.resumeFromComponent(ctx)
		}

		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageData, resumePointID, err := n.prepareMessageSendData(ctx)
		if err != nil {
			return traceError(n, err)
		}

		msg, err := ctx.Discord.EditMessage(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			discord.MessageID(messageTarget.Snowflake()),
			api.EditMessageData{
				Content: option.NewNullableString(messageData.Content),
				Embeds:  &messageData.Embeds,
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, thing.NewDiscordMessage(*msg))
		if resumePointID != "" {
			// We have to create the resume point after the message has been stored
			_, err = ctx.suspend(ResumePointTypeMessageComponents, resumePointID, n.ID)
			if err != nil {
				return traceError(n, err)
			}
		}

		if n.Data.MessageTemplateID != "" {
			err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, provider.MessageTemplateInstance{
				MessageTemplateID: n.Data.MessageTemplateID,
				MessageID:         msg.ID,
				ChannelID:         ctx.Data.ChannelID(),
				GuildID:           ctx.Data.GuildID(),
				Ephemeral:         n.Data.MessageEphemeral,
			})
			if err != nil {
				ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, fmt.Sprintf("failed to link message template instance: %s", err.Error()))
			}
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMessageDelete:
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.DeleteMessage(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			discord.MessageID(messageTarget.Snowflake()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionPrivateMessageCreate:
		if ctx.IsEntry() {
			return n.resumeFromComponent(ctx)
		}

		messageData, resumePointID, err := n.prepareMessageSendData(ctx)
		if err != nil {
			return traceError(n, err)
		}

		userTarget, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		channel, err := ctx.Discord.CreatePrivateChannel(ctx, discord.UserID(userTarget.Snowflake()))
		if err != nil {
			return traceError(n, err)
		}

		msg, err := ctx.Discord.CreateMessage(ctx, channel.ID, messageData)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, thing.NewDiscordMessage(*msg))
		if resumePointID != "" {
			// We have to create the resume point after the message has been stored
			_, err = ctx.suspend(ResumePointTypeMessageComponents, resumePointID, n.ID)
			if err != nil {
				return traceError(n, err)
			}
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMessageReactionCreate:
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		if n.Data.EmojiData == nil {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "emoji_data is nil",
			}
		}

		emoji := discord.APIEmoji(n.Data.EmojiData.Name)
		if n.Data.EmojiData.ID != "" {
			emoji = discord.APIEmoji(fmt.Sprintf("%s:%s", n.Data.EmojiData.Name, n.Data.EmojiData.ID))
		}

		err = ctx.Discord.CreateMessageReaction(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			discord.MessageID(messageTarget.Snowflake()),
			emoji,
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMessageReactionDelete:
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		emoji := discord.APIEmoji(n.Data.EmojiData.Name)
		if n.Data.EmojiData.ID != "" {
			emoji = discord.APIEmoji(fmt.Sprintf("%s:%s", n.Data.EmojiData.Name, n.Data.EmojiData.ID))
		}

		err = ctx.Discord.DeleteMessageReaction(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			discord.MessageID(messageTarget.Snowflake()),
			emoji,
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberBan:
		userID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageDeleteSeconds, err := ctx.EvalTemplate(n.Data.MemberBanDeleteMessageDurationSeconds)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.BanMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(userID.Snowflake()),
			api.BanData{
				DeleteDays:     option.NewUint(uint(messageDeleteSeconds.Float() / 86400)),
				AuditLogReason: api.AuditLogReason(auditLogReason.String()),
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberUnban:
		userID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.UnbanMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(userID.Snowflake()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberKick:
		userID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.KickMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(userID.Snowflake()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberTimeout:
		memberID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		timeoutSeconds, err := ctx.EvalTemplate(n.Data.MemberTimeoutDurationSeconds)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		communicationDisabledUntil := discord.Timestamp(time.Now().UTC().Add(
			time.Duration(timeoutSeconds.Float()) * time.Second,
		))

		err = ctx.Discord.EditMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(memberID.Snowflake()),
			api.ModifyMemberData{
				CommunicationDisabledUntil: &communicationDisabledUntil,
				AuditLogReason:             api.AuditLogReason(auditLogReason.String()),
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberEdit:
		userID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		data := api.ModifyMemberData{
			AuditLogReason: api.AuditLogReason(auditLogReason.String()),
		}

		if n.Data.MemberData != nil {
			if n.Data.MemberData.Nick != nil {
				nick, err := eval.EvalTemplateToString(ctx, *n.Data.MemberData.Nick, ctx.EvalCtx)
				if err != nil {
					return traceError(n, err)
				}

				data.Nick = &nick
			}
		}

		err = ctx.Discord.EditMember(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(userID.Snowflake()),
			data,
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberRoleAdd:
		userID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		roleID, err := ctx.EvalTemplate(n.Data.RoleTarget)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.AddMemberRole(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(userID.Snowflake()),
			discord.RoleID(roleID.Snowflake()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberRoleRemove:
		userID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		roleID, err := ctx.EvalTemplate(n.Data.RoleTarget)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.RemoveMemberRole(
			ctx,
			ctx.Data.GuildID(),
			discord.UserID(userID.Snowflake()),
			discord.RoleID(roleID.Snowflake()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMemberGet:
		guildID := ctx.Data.GuildID()

		if n.Data.GuildTarget != "" {
			guildTarget, err := ctx.EvalTemplate(n.Data.GuildTarget)
			if err != nil {
				return traceError(n, err)
			}

			guildID = discord.GuildID(guildTarget.Snowflake())
		}

		memberID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		member, err := ctx.Discord.Member(ctx, guildID, discord.UserID(memberID.Snowflake()))
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}

		if member != nil {
			ctx.StoreNodeResult(n, thing.NewDiscordMember(*member))
		} else {
			ctx.StoreNodeResult(n, thing.Null)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionUserGet:
		userID, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		user, err := ctx.Discord.User(ctx, discord.UserID(userID.Snowflake()))
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}

		if user != nil {
			ctx.StoreNodeResult(n, thing.NewDiscordUser(*user))
		} else {
			ctx.StoreNodeResult(n, thing.Null)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionChannelGet:
		channelID, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		channel, err := ctx.Discord.Channel(ctx, discord.ChannelID(channelID.Snowflake()))
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}

		if channel != nil {
			ctx.StoreNodeResult(n, thing.NewDiscordChannel(*channel))
		} else {
			ctx.StoreNodeResult(n, thing.Null)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionChannelCreate:
		if n.Data.ChannelData == nil {
			return traceError(n, fmt.Errorf("channel data is required"))
		}

		guildID := ctx.Data.GuildID()
		if n.Data.GuildTarget != "" {
			guildTarget, err := ctx.EvalTemplate(n.Data.GuildTarget)
			if err != nil {
				return traceError(n, err)
			}

			guildID = discord.GuildID(guildTarget.Snowflake())
		}

		channelData, err := n.Data.ChannelData.ToCreateChannelData(ctx, ctx.EvalCtx)
		if err != nil {
			return traceError(n, err)
		}

		channel, err := ctx.Discord.CreateChannel(ctx, guildID, channelData)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, thing.NewDiscordChannel(*channel))
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionChannelEdit:
		if n.Data.ChannelData == nil {
			return traceError(n, fmt.Errorf("channel data is required"))
		}

		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		createData, err := n.Data.ChannelData.ToCreateChannelData(ctx, ctx.EvalCtx)
		if err != nil {
			return traceError(n, err)
		}

		editData := api.ModifyChannelData{
			Name: createData.Name,
			NSFW: &option.NullableBoolData{
				Val:  createData.NSFW,
				Init: true,
			},
		}
		if createData.Topic != "" {
			editData.Topic = option.NewNullableString(createData.Topic)
		}
		if createData.CategoryID != discord.NullChannelID && createData.CategoryID != 0 {
			parentTarget, err := ctx.EvalTemplate(n.Data.ChannelData.ParentID)
			if err != nil {
				return traceError(n, err)
			}

			editData.CategoryID = discord.ChannelID(parentTarget.Snowflake())
		}
		if createData.VoiceBitrate != 0 {
			editData.VoiceBitrate = option.NewNullableUint(createData.VoiceBitrate)
		}
		if createData.VoiceUserLimit != 0 {
			editData.VoiceUserLimit = option.NewNullableUint(createData.VoiceUserLimit)
		}
		if createData.Position != nil {
			editData.Position = option.NewNullableInt(*createData.Position)
		}
		if len(createData.Overwrites) > 0 {
			editData.Overwrites = &createData.Overwrites
		}

		err = ctx.Discord.EditChannel(ctx, discord.ChannelID(channelTarget.Snowflake()), editData)
		if err != nil {
			return traceError(n, err)
		}

		// TODO: ctx.StoreNodeResult(n, thing.NewDiscordChannel(*channel))
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionChannelDelete:
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.DeleteChannel(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionThreadCreate:
		if n.Data.ChannelData == nil {
			return traceError(n, fmt.Errorf("channel data is required"))
		}

		parentTarget, err := ctx.EvalTemplate(n.Data.ChannelData.ParentID)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		channelName, err := ctx.EvalTemplate(n.Data.ChannelData.Name)
		if err != nil {
			return traceError(n, err)
		}

		threadData := api.StartThreadData{
			Name:      channelName.String(),
			Type:      discord.ChannelType(n.Data.ChannelData.Type),
			Invitable: n.Data.ChannelData.Invitable,
		}
		if threadData.Type == 0 {
			threadData.Type = discord.GuildPublicThread
		}

		var thread *discord.Channel
		if messageTarget.IsEmpty() || messageTarget.IsNil() {
			thread, err = ctx.Discord.StartThreadWithoutMessage(
				ctx,
				discord.ChannelID(parentTarget.Snowflake()),
				threadData,
			)
			if err != nil {
				return traceError(n, err)
			}
		} else {
			thread, err = ctx.Discord.StartThreadWithMessage(
				ctx,
				discord.ChannelID(parentTarget.Snowflake()),
				discord.MessageID(messageTarget.Snowflake()),
				threadData,
			)
			if err != nil {
				return traceError(n, err)
			}
		}

		ctx.StoreNodeResult(n, thing.NewDiscordChannel(*thread))
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionThreadMemberAdd:
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		userTarget, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.AddThreadMember(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			discord.UserID(userTarget.Snowflake()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionThreadMemberRemove:
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		userTarget, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Discord.RemoveThreadMember(
			ctx,
			discord.ChannelID(channelTarget.Snowflake()),
			discord.UserID(userTarget.Snowflake()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionForumPostCreate:
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionRoleGet:
		guildID := ctx.Data.GuildID()

		if n.Data.GuildTarget != "" {
			guildTarget, err := ctx.EvalTemplate(n.Data.GuildTarget)
			if err != nil {
				return traceError(n, err)
			}

			guildID = discord.GuildID(guildTarget.Snowflake())
		}

		roleID, err := ctx.EvalTemplate(n.Data.RoleTarget)
		if err != nil {
			return traceError(n, err)
		}

		role, err := ctx.Discord.Role(ctx, guildID, discord.RoleID(roleID.Snowflake()))
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}

		if role != nil {
			ctx.StoreNodeResult(n, thing.NewDiscordRole(*role))
		} else {
			ctx.StoreNodeResult(n, thing.Null)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionGuildGet:
		guildID, err := ctx.EvalTemplate(n.Data.GuildTarget)
		if err != nil {
			return traceError(n, err)
		}

		guild, err := ctx.Discord.Guild(ctx, discord.GuildID(guildID.Snowflake()))
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}

		if guild != nil {
			ctx.StoreNodeResult(n, thing.NewDiscordGuild(*guild))
		} else {
			ctx.StoreNodeResult(n, thing.Null)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMessageGet:
		channelID := ctx.Data.ChannelID()

		if n.Data.ChannelTarget != "" {
			channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
			if err != nil {
				return traceError(n, err)
			}

			channelID = discord.ChannelID(channelTarget.Snowflake())
		}

		messageID, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		message, err := ctx.Discord.Message(
			ctx,
			channelID,
			discord.MessageID(messageID.Snowflake()),
		)
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}

		if message != nil {
			ctx.StoreNodeResult(n, thing.NewDiscordMessage(*message))
		} else {
			ctx.StoreNodeResult(n, thing.Null)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionRobloxUserGet:
		userID, err := ctx.EvalTemplate(n.Data.RobloxUserTarget)
		if err != nil {
			return traceError(n, err)
		}

		res := thing.Null

		switch n.Data.RobloxLookupMode {
		case RobloxLookupTypeID:
			user, err := ctx.Roblox.UserByID(ctx, userID.Int())
			if err != nil && !errors.Is(err, provider.ErrNotFound) {
				return traceError(n, err)
			}
			if user != nil {
				res = thing.NewRobloxUser(*user)
			}
		case RobloxLookupTypeName:
			users, err := ctx.Roblox.UsersByUsername(ctx, userID.String())
			if err != nil && !errors.Is(err, provider.ErrNotFound) {
				return traceError(n, err)
			}
			if len(users) > 0 {
				res = thing.NewRobloxUser(users[0])
			}
		}

		ctx.StoreNodeResult(n, res)
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionVariableSet:
		scope, err := ctx.EvalTemplate(n.Data.VariableScope)
		if err != nil {
			return traceError(n, err)
		}

		value, err := ctx.EvalTemplate(n.Data.VariableValue)
		if err != nil {
			return traceError(n, err)
		}

		newValue, err := ctx.Variable.UpdateVariable(
			ctx,
			n.Data.VariableID,
			null.NewString(scope.String(), !scope.IsEmpty()),
			n.Data.VariableOperation,
			value,
		)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, newValue)
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionVariableDelete:
		scope, err := ctx.EvalTemplate(n.Data.VariableScope)
		if err != nil {
			return traceError(n, err)
		}

		err = ctx.Variable.DeleteVariable(
			ctx,
			n.Data.VariableID,
			null.NewString(scope.String(), !scope.IsEmpty()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionVariableGet:
		scope, err := ctx.EvalTemplate(n.Data.VariableScope)
		if err != nil {
			return traceError(n, err)
		}

		val, err := ctx.Variable.Variable(
			ctx,
			n.Data.VariableID,
			null.NewString(scope.String(), !scope.IsEmpty()),
		)
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, val)
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionHTTPRequest:
		if n.Data.HTTPRequestData == nil {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "http_request_data is nil",
			}
		}

		method := n.Data.HTTPRequestData.Method
		if method == "" {
			method = "GET"
		}

		url, err := ctx.EvalTemplate(n.Data.HTTPRequestData.URL)
		if err != nil {
			return traceError(n, err)
		}

		req, err := http.NewRequest(method, url.String(), nil)
		if err != nil {
			return traceError(n, err)
		}

		for _, header := range n.Data.HTTPRequestData.Headers {
			value, err := ctx.EvalTemplate(header.Value)
			if err != nil {
				return traceError(n, err)
			}

			req.Header.Add(header.Key, value.String())
		}

		query := req.URL.Query()
		for _, queryParam := range n.Data.HTTPRequestData.Query {
			value, err := ctx.EvalTemplate(queryParam.Value)
			if err != nil {
				return traceError(n, err)
			}

			query.Add(queryParam.Key, value.String())
		}
		req.URL.RawQuery = query.Encode()

		if n.Data.HTTPRequestData.BodyJSON != nil {
			// This can potentially break the JSON if an expression returns a string containing double quotes
			// We should probably escape the expression results or only run the eval engine on the actual JSON values
			body, err := ctx.EvalTemplate(string(n.Data.HTTPRequestData.BodyJSON))
			if err != nil {
				return traceError(n, err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Body = io.NopCloser(bytes.NewReader([]byte(body.String())))
		}

		resp, err := ctx.HTTP.HTTPRequest(ctx, req)
		if err != nil {
			return traceError(n, err)
		}

		result, err := thing.NewFromHTTPResponse(resp)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, result)
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionAIChatCompletion:
		data := n.Data.AIChatCompletionData
		if data == nil || data.Prompt == "" {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "ai_chat_completion_data is nil",
			}
		}

		systemPrompt, err := ctx.EvalTemplate(data.SystemPrompt)
		if err != nil {
			return traceError(n, err)
		}

		prompt, err := ctx.EvalTemplate(data.Prompt)
		if err != nil {
			return traceError(n, err)
		}

		maxCompletionTokens, err := ctx.EvalTemplate(data.MaxCompletionTokens)
		if err != nil {
			return traceError(n, err)
		}

		response, err := ctx.AI.CreateResponse(ctx, provider.CreateResponseOpts{
			Model:           data.Model,
			Prompt:          prompt.String(),
			SystemPrompt:    systemPrompt.String(),
			MaxOutputTokens: int(maxCompletionTokens.Int()),
		})
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, thing.NewString(response))
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionAISearchWeb:
		data := n.Data.AIChatCompletionData
		if data == nil || data.Prompt == "" {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "ai_search_web_data is nil",
			}
		}

		systemPrompt, err := ctx.EvalTemplate(data.SystemPrompt)
		if err != nil {
			return traceError(n, err)
		}

		prompt, err := ctx.EvalTemplate(data.Prompt)
		if err != nil {
			return traceError(n, err)
		}

		maxCompletionTokens, err := ctx.EvalTemplate(data.MaxCompletionTokens)
		if err != nil {
			return traceError(n, err)
		}

		response, err := ctx.AI.CreateResponse(ctx, provider.CreateResponseOpts{
			Model:           data.Model,
			Prompt:          prompt.String(),
			SystemPrompt:    systemPrompt.String(),
			MaxOutputTokens: int(maxCompletionTokens.Int()),
			Tools:           []provider.AIToolType{provider.AIToolTypeWebSearchPreview},
		})
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, thing.NewString(response))
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionRandomGenerate:
		min, err := ctx.EvalTemplate(n.Data.RandomMin)
		if err != nil {
			return traceError(n, err)
		}

		max, err := ctx.EvalTemplate(n.Data.RandomMax)
		if err != nil {
			return traceError(n, err)
		}

		minInt := int(min.Int())
		maxInt := int(max.Int())
		if maxInt <= 0 || minInt >= maxInt {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "random_generate_max must be greater than random_generate_min and greater than 0",
			}
		}

		ctx.StoreNodeResult(n, thing.NewInt(rand.Intn(maxInt-minInt)+minInt))
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionExpressionEvaluate:
		expression, err := ctx.EvalTemplate(n.Data.Expression)
		if err != nil {
			return traceError(n, err)
		}

		res, err := eval.Eval(ctx, expression.String(), ctx.EvalCtx)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeResult(n, res)
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionLog:
		logMessage, err := ctx.EvalTemplate(n.Data.LogMessage)
		if err != nil {
			return traceError(n, err)
		}

		ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, logMessage.String())
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeControlConditionCompare,
		FlowNodeTypeControlConditionUser,
		FlowNodeTypeControlConditionChannel,
		FlowNodeTypeControlConditionRole:

		baseValue, err := ctx.EvalTemplate(n.Data.ConditionBaseValue)
		if err != nil {
			return traceError(n, err)
		}

		ctx.StoreNodeBaseValue(n, baseValue)

		var elseNode *CompiledFlowNode

		for _, child := range n.Children.Default {
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

		if parentState.ConditionItemMet && !parent.Data.ConditionAllowMultiple {
			// Another condition item has already been met
			return nil
		}

		itemValue, err := ctx.EvalTemplate(n.Data.ConditionItemValue)
		if err != nil {
			return traceError(n, err)
		}

		baseValue := parentState.ConditionBaseValue

		var conditionMet bool
		switch n.Data.ConditionItemMode {
		case ComparsionModeEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ComparsionModeNotEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ComparsionModeGreaterThan:
			conditionMet = baseValue.GreaterThan(&itemValue)
		case ComparsionModeGreaterThanOrEqual:
			conditionMet = baseValue.GreaterThanOrEqual(&itemValue)
		case ComparsionModeLessThan:
			conditionMet = baseValue.LessThan(&itemValue)
		case ComparsionModeLessThanOrEqual:
			conditionMet = baseValue.LessThanOrEqual(&itemValue)
		case ComparsionModeContains:
			conditionMet = baseValue.Contains(&itemValue)
		case ComparsionModeStartsWith:
			conditionMet = baseValue.StartsWith(&itemValue)
		case ComparsionModeEndsWith:
			conditionMet = baseValue.EndsWith(&itemValue)
		}

		if conditionMet {
			parentState.ConditionItemMet = true
			return n.ExecuteChildren(ctx)
		}
	case FlowNodeTypeControlConditionItemUser,
		FlowNodeTypeControlConditionItemChannel,
		FlowNodeTypeControlConditionItemRole:
		parent := n.FindDirectParentWithType(
			FlowNodeTypeControlConditionUser,
			FlowNodeTypeControlConditionChannel,
			FlowNodeTypeControlConditionRole,
		)
		if parent == nil {
			return nil
		}

		parentState := ctx.GetNodeState(parent.ID)

		if parentState.ConditionItemMet && !parent.Data.ConditionAllowMultiple {
			// Another condition item has already been met
			return nil
		}

		itemValue, err := ctx.EvalTemplate(n.Data.ConditionItemValue)
		if err != nil {
			return traceError(n, err)
		}

		baseValue := parentState.ConditionBaseValue

		var conditionMet bool
		switch n.Data.ConditionItemMode {
		case ComparsionModeEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ComparsionModeNotEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ComparsionModeHasRole:
			member := baseValue.DiscordMember()
			if !member.User.ID.IsValid() {
				// TODO?: fetch member by id from discord?
				return nil
			}
			conditionMet = slices.Contains(member.RoleIDs, discord.RoleID(itemValue.Int()))
		case ComparsionModeNotHasRole:
			member := baseValue.DiscordMember()
			if !member.User.ID.IsValid() {
				// TODO?: fetch member by id from discord?
				return nil
			}
			conditionMet = !slices.Contains(member.RoleIDs, discord.RoleID(itemValue.Int()))
		case ComparsionModeHasPermission:
			member := baseValue.DiscordMember()
			if !member.User.ID.IsValid() {
				// TODO?: fetch member by id from discord?
				return nil
			}

			roles, err := ctx.Discord.GuildRoles(ctx, ctx.Data.GuildID())
			if err != nil {
				return traceError(n, err)
			}

			var permission discord.Permissions
			for _, role := range roles {
				if slices.Contains(member.RoleIDs, role.ID) {
					permission |= role.Permissions
				}
			}

			itemPermissions := discord.Permissions(itemValue.Int())
			conditionMet = permission&itemPermissions == itemPermissions
		case ComparsionModeNotHasPermission:
			member := baseValue.DiscordMember()
			if !member.User.ID.IsValid() {
				// TODO?: fetch member by id from discord?
				return nil
			}

			roles, err := ctx.Discord.GuildRoles(ctx, ctx.Data.GuildID())
			if err != nil {
				return traceError(n, err)
			}

			var permission discord.Permissions
			for _, role := range roles {
				if slices.Contains(member.RoleIDs, role.ID) {
					permission |= role.Permissions
				}
			}

			itemPermissions := discord.Permissions(itemValue.Int())
			conditionMet = permission&itemPermissions != itemPermissions
		}

		if conditionMet {
			parentState.ConditionItemMet = true
			return n.ExecuteChildren(ctx)
		}
	case FlowNodeTypeControlConditionItemElse:
		parent := n.FindDirectParentWithType(
			FlowNodeTypeControlConditionCompare,
			FlowNodeTypeControlConditionUser,
			FlowNodeTypeControlConditionChannel,
			FlowNodeTypeControlConditionRole,
		)
		if parent == nil {
			return nil
		}

		parentState := ctx.GetNodeState(parent.ID)

		if parentState.ConditionItemMet {
			// Another condition item has already been met
			return nil
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeControlErrorHandler:
		err := n.ExecuteChildren(ctx)
		if err != nil {
			ctx.StoreNodeResult(n, thing.NewString(err.Error()))
			return n.ExecuteChildrenByHandle(ctx, "error")
		}

		return nil
	case FlowNodeTypeControlLoop:
		loopCount, err := ctx.EvalTemplate(n.Data.LoopCount)
		if err != nil {
			return traceError(n, err)
		}

		eachNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEach)
		endNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEnd)

		nodeState := ctx.GetNodeState(n.ID)

		for i := 0; i < int(loopCount.Int()); i++ {
			if nodeState.LoopExited {
				break
			}

			if err := eachNode.Execute(ctx); err != nil {
				return traceError(n, err)
			}
		}

		if err := endNode.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	case FlowNodeTypeControlLoopEach:
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeControlLoopEnd:
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeControlLoopExit:
		// Mark all parent loops as exited
		parentLoops := n.FindAllParentsWithType(FlowNodeTypeControlLoop)
		for _, loop := range parentLoops {
			ctx.GetNodeState(loop.ID).LoopExited = true
		}
	case FlowNodeTypeControlSleep:
		sleepSeconds, err := ctx.EvalTemplate(n.Data.SleepDurationSeconds)
		if err != nil {
			return traceError(n, err)
		}

		duration := time.Duration(sleepSeconds.Float()) * time.Second

		deadline, ok := ctx.Deadline()
		if ok && time.Now().Add(duration).After(deadline) {
			return &FlowError{
				Code:    FlowNodeErrorTimeout,
				Message: "sleep would exceed deadline",
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(duration):
			return n.ExecuteChildren(ctx)
		}
	default:
		return &FlowError{
			Code:    FlowNodeErrorUnknownNodeType,
			Message: fmt.Sprintf("unknown node type: %s", n.Type),
		}
	}

	return nil
}

func (n *CompiledFlowNode) CreditsCost() int {
	switch n.Type {
	case FlowNodeTypeActionAIChatCompletion:
		data := n.Data.AIChatCompletionData
		if data == nil {
			return 0
		}

		switch data.Model {
		case openai.ChatModelGPT4_1:
			return 100
		case openai.ChatModelGPT4_1Mini:
			return 20
		default:
			return 5
		}
	case FlowNodeTypeActionAISearchWeb:
		data := n.Data.AIChatCompletionData
		if data == nil {
			return 0
		}

		switch data.Model {
		case openai.ChatModelGPT4_1:
			return 500
		case openai.ChatModelGPT4_1Mini:
			return 100
		default:
			return 25
		}
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
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	// We check if the next response will be ephemeral and adjust the defer flags accordingly
	var responseFlags discord.MessageFlags
	respondeNode := n.FindChildWithType(FlowNodeTypeActionResponseCreate, FlowNodeTypeActionResponseEdit, FlowNodeTypeActionResponseDefer)
	if respondeNode != nil && respondeNode.Data.MessageEphemeral {
		responseFlags |= discord.EphemeralMessage
	}

	go ctx.Discord.AutoDeferInteraction(ctx, interaction.ID, interaction.Token, responseFlags)
	return nil
}

func (n *CompiledFlowNode) resumeFromComponent(ctx *FlowContext) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	err := n.autoDeferInteraction(ctx)
	if err != nil {
		return traceError(n, err)
	}

	data := interaction.Data.(*discord.ButtonInteraction)
	_, compID, ok := message.DecodeCustomIDMessageComponentResumePoint(string(data.CustomID))
	if !ok {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "invalid custom ID",
		}
	}

	err = n.ExecuteChildrenByHandle(ctx, fmt.Sprintf("component_%d", compID))
	if err != nil {
		createDefaultErrorResponse(ctx, err)
		return traceError(n, err)
	}

	return nil
}

func (n *CompiledFlowNode) prepareMessageData(ctx *FlowContext) (message.MessageData, error) {
	var data message.MessageData
	if n.Data.MessageTemplateID != "" {
		template, err := ctx.MessageTemplate.MessageTemplate(ctx, n.Data.MessageTemplateID)
		if err != nil {
			return message.MessageData{}, err
		}
		data = *template
	} else {
		data = n.Data.MessageData.Copy()
	}

	if n.Data.MessageEphemeral {
		data.Flags |= int(discord.EphemeralMessage)
	}

	err := data.EachString(func(s *string) error {
		if s == nil {
			return nil
		}

		res, err := eval.EvalTemplateToString(ctx, *s, ctx.EvalCtx)
		if err != nil {
			return err
		}

		*s = res
		return nil
	})
	if err != nil {
		return message.MessageData{}, err
	}

	return data, nil
}

func (n *CompiledFlowNode) prepareMessageResponseData(ctx *FlowContext) (api.InteractionResponseData, string, error) {
	data, err := n.prepareMessageData(ctx)
	if err != nil {
		return api.InteractionResponseData{}, "", err
	}

	var resumePointID string
	if n.Data.MessageTemplateID == "" && len(data.Components) > 0 {
		resumePointID = util.UniqueID()
	}

	responseData := data.ToInteractionResponseData(message.ConvertOptions{
		ComponentIDFactory: func(component *message.ComponentData) discord.ComponentID {
			if resumePointID != "" {
				return discord.ComponentID(message.CustomIDMessageComponentResumePoint(resumePointID, component.ID))
			}
			return discord.ComponentID(component.FlowSourceID)
		},
	})

	return responseData, resumePointID, nil
}

func (n *CompiledFlowNode) prepareMessageSendData(ctx *FlowContext) (api.SendMessageData, string, error) {
	data, err := n.prepareMessageData(ctx)
	if err != nil {
		return api.SendMessageData{}, "", err
	}

	var resumePointID string
	if n.Data.MessageTemplateID == "" && len(data.Components) > 0 {
		// The resume point will be created after the message has been sent, we just need the ID here already
		resumePointID = util.UniqueID()
	}

	sendData := data.ToSendMessageData(message.ConvertOptions{
		ComponentIDFactory: func(component *message.ComponentData) discord.ComponentID {
			if resumePointID != "" {
				return discord.ComponentID(message.CustomIDMessageComponentResumePoint(resumePointID, component.ID))
			}
			return discord.ComponentID(component.FlowSourceID)
		},
	})

	return sendData, resumePointID, nil
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
