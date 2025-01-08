package flow

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
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
		return n.executeChildren(ctx)
	case FlowNodeTypeEntryEvent:
		return n.executeChildren(ctx)
	case FlowNodeTypeEntryComponentButton:
		return n.executeChildren(ctx)
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

			res, err := eval.EvalTemplateToString(ctx, *s, ctx.EvalEnv)
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

		return n.executeChildren(ctx)
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

			res, err := eval.EvalTemplateToString(ctx, *s, ctx.EvalEnv)
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
		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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

			res, err := eval.EvalTemplateToString(ctx, *s, ctx.EvalEnv)
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
		return n.executeChildren(ctx)
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

			res, err := eval.EvalTemplateToString(ctx, *s, ctx.EvalEnv)
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
		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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

			res, err := eval.EvalTemplateToString(ctx, *s, ctx.EvalEnv)
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
		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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
				nick, err := eval.EvalTemplateToString(ctx, *n.Data.MemberData.Nick, ctx.EvalEnv)
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

		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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
		return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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
		return n.executeChildren(ctx)
	case FlowNodeTypeActionHTTPRequest:
		if n.Data.HTTPRequestData == nil {
			return &FlowError{
				Code:    FlowNodeErrorUnknown,
				Message: "http_request_data is nil",
			}
		}

		url, err := ctx.EvalTemplate(n.Data.HTTPRequestData.URL)
		if err != nil {
			return traceError(n, err)
		}

		req, err := http.NewRequest(n.Data.HTTPRequestData.Method, url.String(), nil)
		if err != nil {
			return traceError(n, err)
		}

		resp, err := ctx.HTTP.HTTPRequest(ctx, req)
		if err != nil {
			return traceError(n, err)
		}

		nodeState.Result = thing.New(resp)
		return n.executeChildren(ctx)
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
		return n.executeChildren(ctx)
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
		return n.executeChildren(ctx)
	case FlowNodeTypeActionExpressionEvaluate:
		expression, err := ctx.EvalTemplate(n.Data.Expression)
		if err != nil {
			return traceError(n, err)
		}

		res, err := eval.Eval(ctx, expression.String(), ctx.EvalEnv)
		if err != nil {
			return traceError(n, err)
		}

		nodeState.Result = res
		return n.executeChildren(ctx)
	case FlowNodeTypeActionLog:
		logMessage, err := ctx.EvalTemplate(n.Data.LogMessage)
		if err != nil {
			return traceError(n, err)
		}

		ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, logMessage.String())
		return n.executeChildren(ctx)
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
			return n.executeChildren(ctx)
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
		}

		if conditionMet {
			parentState.ConditionItemMet = true
			return n.executeChildren(ctx)
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

		return n.executeChildren(ctx)
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
		return n.executeChildren(ctx)
	case FlowNodeTypeControlLoopEnd:
		return n.executeChildren(ctx)
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
		return n.executeChildren(ctx)
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

func (n *CompiledFlowNode) executeChildren(ctx *FlowContext) error {
	for _, child := range n.Children {
		// We could spawn a goroutine here to execute children in parallel
		// but we'll just execute them sequentially for now
		if err := child.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}
	return nil
}
