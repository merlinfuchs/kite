package flow

import (
	"errors"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionMessageCreate:        executeActionMessageCreate,
		FlowNodeTypeActionMessageEdit:          executeActionMessageEdit,
		FlowNodeTypeActionPrivateMessageCreate: executeActionPrivateMessageCreate,
		FlowNodeTypeActionPollCreate:           executeActionPollCreate,
		FlowNodeTypeActionMessageGet:           executeActionMessageGet,
	})
}

func executeActionMessageCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.IsEntry() {
		return n.resumeFromComponent(ctx)
	}

	data, opts, resumePointID, err := n.prepareMessage(ctx)
	if err != nil {
		return traceError(n, err)
	}
	messageData := data.ToSendMessageData(opts)

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

func executeActionMessageEdit(n *CompiledFlowNode, ctx *FlowContext) error {
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

	data, opts, resumePointID, err := n.prepareMessage(ctx)
	if err != nil {
		return traceError(n, err)
	}
	editData := data.ToEditMessageData(opts)

	msg, err := ctx.Discord.EditMessage(
		ctx,
		discord.ChannelID(channelTarget.Snowflake()),
		discord.MessageID(messageTarget.Snowflake()),
		editData,
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

func executeActionPrivateMessageCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	if ctx.IsEntry() {
		return n.resumeFromComponent(ctx)
	}

	data, opts, resumePointID, err := n.prepareMessage(ctx)
	if err != nil {
		return traceError(n, err)
	}
	messageData := data.ToSendMessageData(opts)

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
}

func executeActionPollCreate(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.PollData == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "poll_data is nil",
		}
	}

	channelTarget, err := ctx.EvalTemplate(n.Data.ChannelTarget)
	if err != nil {
		return traceError(n, err)
	}

	pollData, err := n.Data.PollData.ToCreatePollData(ctx, ctx.EvalCtx)
	if err != nil {
		return traceError(n, err)
	}

	msg, err := ctx.Discord.CreatePoll(
		ctx,
		discord.ChannelID(channelTarget.Snowflake()),
		pollData,
	)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, thing.NewDiscordMessage(*msg))
	return n.ExecuteChildren(ctx)
}

func executeActionMessageGet(n *CompiledFlowNode, ctx *FlowContext) error {
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

// prepareMessage evaluates the node's message and returns it with the options to
// convert it, pointing interactive components at a new resume point if needed.
func (n *CompiledFlowNode) prepareMessage(ctx *FlowContext) (message.MessageData, message.ConvertOptions, string, error) {
	data, err := n.prepareMessageData(ctx)
	if err != nil {
		return message.MessageData{}, message.ConvertOptions{}, "", err
	}

	var resumePointID string
	if n.Data.MessageTemplateID == "" && data.HasInteractiveComponents() {
		// The resume point will be created after the message has been sent, we just need the ID here already
		resumePointID = util.UniqueID()
	}

	opts := message.ConvertOptions{
		ComponentIDFactory: func(component *message.ComponentData) discord.ComponentID {
			if resumePointID != "" {
				return discord.ComponentID(message.CustomIDMessageComponentResumePoint(resumePointID, component.ID))
			}
			return discord.ComponentID(component.FlowSourceID)
		},
	}

	return data, opts, resumePointID, nil
}
