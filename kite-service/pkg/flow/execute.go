package flow

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
)

func (n *CompiledFlowNode) Execute(ctx *FlowContext) error {
	if err := ctx.startOperation(n.CreditsCost()); err != nil {
		return traceError(n, err)
	}
	defer ctx.endOperation()

	nodeState := ctx.GetNodeState(n.ID)

	switch n.Type {
	case FlowNodeTypeEntryCommand:
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeEntryEvent:
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeEntryComponentButton:
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionResponseCreate:
		interaction := ctx.Data.Interaction()
		if interaction == nil {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "interaction is nil",
			}
		}

		var data message.MessageData
		if n.Data.MessageTemplateID != "" {
			template, err := ctx.MessageTemplate.MessageTemplate(ctx, n.Data.MessageTemplateID)
			if err != nil {
				return traceError(n, err)
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
			return traceError(n, err)
		}

		hasCreatedResponse, err := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
		if err != nil {
			return traceError(n, err)
		}

		var msg *discord.Message
		if hasCreatedResponse {
			responseData := data.ToInteractionResponseData()

			msg, err = ctx.Discord.CreateInteractionFollowup(ctx, interaction.AppID, interaction.Token, responseData)
			if err != nil {
				return traceError(n, err)
			}

			nodeState.Result = thing.New(msg)
		} else {
			responseData := data.ToInteractionResponseData()

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

		if n.Data.MessageTemplateID != "" && msg != nil {
			err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, FlowMessageTemplateInstance{
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
		interaction := ctx.Data.Interaction()
		if interaction == nil {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "interaction is nil",
			}
		}

		var data message.MessageData
		if n.Data.MessageTemplateID != "" {
			template, err := ctx.MessageTemplate.MessageTemplate(ctx, n.Data.MessageTemplateID)
			if err != nil {
				return traceError(n, err)
			}
			data = *template
		} else {
			data = n.Data.MessageData.Copy()
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
			return traceError(n, err)
		}

		responseData := data.ToInteractionResponseData()

		var msg *discord.Message
		if n.Data.MessageTarget == "" || n.Data.MessageTarget == "@original" {
			msg, err = ctx.Discord.EditInteractionResponse(ctx, interaction.AppID, interaction.Token, api.EditInteractionResponseData{
				Content:    responseData.Content,
				Embeds:     responseData.Embeds,
				Components: responseData.Components,
			})
			if err != nil {
				return traceError(n, err)
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
				discord.MessageID(messageTarget.Int()),
				api.EditInteractionResponseData{
					Content: responseData.Content,
					Embeds:  responseData.Embeds,
				},
			)
			if err != nil {
				return traceError(n, err)
			}
		}

		if n.Data.MessageTemplateID != "" {
			err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, FlowMessageTemplateInstance{
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

		nodeState.Result = thing.New(msg)
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
				discord.MessageID(messageTarget.Int()),
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

		resumePoint, err := ctx.suspend(FlowResumePointTypeModal, n.ID)
		if err != nil {
			return traceError(n, err)
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
				CustomID:   option.NewNullableString("resume:" + resumePoint.ID),
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
		var data message.MessageData
		if n.Data.MessageTemplateID != "" {
			template, err := ctx.MessageTemplate.MessageTemplate(ctx, n.Data.MessageTemplateID)
			if err != nil {
				return traceError(n, err)
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
			return traceError(n, err)
		}

		messageData := data.ToSendMessageData()

		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		msg, err := ctx.Discord.CreateMessage(ctx, discord.ChannelID(channelTarget.Int()), messageData)
		if err != nil {
			return traceError(n, err)
		}

		if n.Data.MessageTemplateID != "" {
			err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, FlowMessageTemplateInstance{
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

		nodeState.Result = thing.New(msg)
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionMessageEdit:
		channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
		if err != nil {
			return traceError(n, err)
		}

		messageTarget, err := ctx.EvalTemplate(n.Data.MessageTarget)
		if err != nil {
			return traceError(n, err)
		}

		var data message.MessageData
		if n.Data.MessageTemplateID != "" {
			template, err := ctx.MessageTemplate.MessageTemplate(ctx, n.Data.MessageTemplateID)
			if err != nil {
				return traceError(n, err)
			}
			data = *template
		} else {
			data = n.Data.MessageData.Copy()
		}

		if n.Data.MessageEphemeral {
			data.Flags |= int(discord.EphemeralMessage)
		}

		err = data.EachString(func(s *string) error {
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
			return traceError(n, err)
		}

		messageData := data.ToSendMessageData()

		msg, err := ctx.Discord.EditMessage(
			ctx,
			discord.ChannelID(channelTarget.Int()),
			discord.MessageID(messageTarget.Int()),
			api.EditMessageData{
				Content: option.NewNullableString(messageData.Content),
				Embeds:  &messageData.Embeds,
			},
		)
		if err != nil {
			return traceError(n, err)
		}

		if n.Data.MessageTemplateID != "" {
			err := ctx.MessageTemplate.LinkMessageTemplateInstance(ctx, FlowMessageTemplateInstance{
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

		nodeState.Result = thing.New(msg)
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
			discord.ChannelID(channelTarget.Int()),
			discord.MessageID(messageTarget.Int()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionPrivateMessageCreate:
		var data message.MessageData
		if n.Data.MessageTemplateID != "" {
			template, err := ctx.MessageTemplate.MessageTemplate(ctx, n.Data.MessageTemplateID)
			if err != nil {
				return traceError(n, err)
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
			return traceError(n, err)
		}

		messageData := data.ToSendMessageData()

		userTarget, err := ctx.EvalTemplate(n.Data.UserTarget)
		if err != nil {
			return traceError(n, err)
		}

		channel, err := ctx.Discord.CreatePrivateChannel(ctx, discord.UserID(userTarget.Int()))
		if err != nil {
			return traceError(n, err)
		}

		msg, err := ctx.Discord.CreateMessage(ctx, channel.ID, messageData)
		if err != nil {
			return traceError(n, err)
		}

		nodeState.Result = thing.New(msg)
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
			discord.UserID(userID.Int()),
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
			discord.UserID(userID.Int()),
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
			discord.UserID(userID.Int()),
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
			discord.UserID(memberID.Int()),
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
			discord.UserID(userID.Int()),
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
			discord.UserID(userID.Int()),
			discord.RoleID(roleID.Int()),
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
			discord.UserID(userID.Int()),
			discord.RoleID(roleID.Int()),
			api.AuditLogReason(auditLogReason.String()),
		)
		if err != nil {
			return traceError(n, err)
		}

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

		nodeState.Result = newValue
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
		if err != nil && !errors.Is(err, ErrNotFound) {
			return traceError(n, err)
		}

		nodeState.Result = val
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

		nodeState.Result = thing.New(resp)
		return n.ExecuteChildren(ctx)
	case FlowNodeTypeActionAIChatCompletion:
		data := n.Data.AIChatCompletionData
		if data == nil || data.Prompt == "" {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "ai_prompt is nil",
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

		response, err := ctx.AI.CreateChatCompletion(ctx, CreateChatCompletionOpts{
			SystemPrompt:        systemPrompt.String(),
			Prompt:              prompt.String(),
			MaxCompletionTokens: int(maxCompletionTokens.Int()),
		})
		if err != nil {
			return traceError(n, err)
		}

		nodeState.Result = thing.New(response)
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

		nodeState.Result = thing.New(rand.Intn(maxInt + minInt))
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

		nodeState.Result = res
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
		case ConditionItemModeEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ConditionItemModeNotEqual:
			conditionMet = baseValue.Equals(&itemValue)
		case ConditionItemModeHasRole:
			member, ok := thing.Cast[RolesCastable](baseValue)
			if !ok {
				// TODO?: fetch member by id from discord?
				return nil
			}

			conditionMet = member.HasRole(itemValue.String())
		case ConditionItemModeNotHasRole:
			member, ok := thing.Cast[RolesCastable](baseValue)
			if !ok {
				return nil
			}

			conditionMet = !member.HasRole(itemValue.String())
		case ConditionItemModeHasPermission:
			member, ok := thing.Cast[RolesCastable](baseValue)
			if !ok {
				return nil
			}

			roles, err := ctx.Discord.GuildRoles(ctx, ctx.Data.GuildID())
			if err != nil {
				return traceError(n, err)
			}

			var permission discord.Permissions
			for _, role := range roles {
				if slices.Contains(member.Roles(), role.ID.String()) {
					permission |= role.Permissions
				}
			}

			itemPermissions := discord.Permissions(itemValue.Int())
			conditionMet = permission&itemPermissions == itemPermissions
		case ConditionItemModeNotHasPermission:
			member, ok := thing.Cast[RolesCastable](baseValue)
			if !ok {
				return nil
			}

			roles, err := ctx.Discord.GuildRoles(ctx, ctx.Data.GuildID())
			if err != nil {
				return traceError(n, err)
			}

			var permission discord.Permissions
			for _, role := range roles {
				if slices.Contains(member.Roles(), role.ID.String()) {
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
	case FlowNodeTypeControlLoop:
		loopCount, err := ctx.EvalTemplate(n.Data.LoopCount)
		if err != nil {
			return traceError(n, err)
		}

		eachNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEach)
		endNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEnd)

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

		time.Sleep(duration)
		return n.ExecuteChildren(ctx)
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
		return 5
	case FlowNodeTypeActionHTTPRequest:
		return 3
	}

	if n.IsAction() {
		return 1
	}

	return 0
}

func (n *CompiledFlowNode) ExecuteChildren(ctx *FlowContext) error {
	for _, child := range n.Children {
		// We could spawn a goroutine here to execute children in parallel
		// but we'll just execute them sequentially for now
		if err := child.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}
	return nil
}

type RolesCastable interface {
	HasRole(roleID string) bool
	Roles() []string
}
