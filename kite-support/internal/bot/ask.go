package bot

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-support/internal/llm"
)

func (b *Bot) onInteraction(e *gateway.InteractionCreateEvent) {
	switch d := e.Data.(type) {
	case *discord.CommandInteraction:
		b.handleCommand(e, d)
	case *discord.ButtonInteraction:
		b.handleButton(e, d)
	case *discord.ModalInteraction:
		b.handleModal(e, d)
	}
}

func (b *Bot) handleCommand(e *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) {
	switch cmd.Name {
	case "ask":
		question := ""
		for _, opt := range cmd.Options {
			if opt.Name == "question" {
				question = opt.String()
			}
		}
		b.runAsk(e, strings.TrimSpace(question), nil)
	case "setup-menu":
		b.handleSetupMenu(e)
	}
}

func (b *Bot) handleButton(e *gateway.InteractionCreateEvent, btn *discord.ButtonInteraction) {
	switch btn.CustomID {
	case customIDMenuAsk, customIDFollowupAsk:
		if err := b.state.RespondInteraction(e.ID, e.Token, askModalResponse()); err != nil {
			slog.With("err", err).Error("open modal failed")
		}
	case customIDFeedbackOpen:
		if err := b.state.RespondInteraction(e.ID, e.Token, feedbackModalResponse()); err != nil {
			slog.With("err", err).Error("open feedback modal failed")
		}
	case customIDHelpOpen:
		if err := b.state.RespondInteraction(e.ID, e.Token, helpModalResponse()); err != nil {
			slog.With("err", err).Error("open help modal failed")
		}
	}
}

func (b *Bot) handleModal(e *gateway.InteractionCreateEvent, modal *discord.ModalInteraction) {
	switch modal.CustomID {
	case customIDAskModal:
		// A follow-up is asked from the previous answer, so it continues that
		// conversation.
		var history []llm.Turn
		if e.Message != nil {
			if prev, ok := b.feedback.Get(e.Message.ID); ok {
				history = prev.History
			}
		}
		b.runAsk(e, strings.TrimSpace(extractText(modal, customIDQuestion)), history)
	case customIDFeedbackModal:
		b.handleFeedbackSubmit(e, strings.TrimSpace(extractText(modal, customIDFeedbackDetail)))
	case customIDHelpModal:
		b.handleHelpSubmit(e, strings.TrimSpace(extractText(modal, customIDHelpDetail)))
	}
}

func extractText(modal *discord.ModalInteraction, customID discord.ComponentID) string {
	for _, row := range modal.Components {
		row, ok := row.(*discord.ActionRowComponent)
		if !ok {
			continue
		}
		for _, comp := range *row {
			ti, ok := comp.(*discord.TextInputComponent)
			if !ok {
				continue
			}
			if ti.CustomID == customID {
				return ti.Value
			}
		}
	}
	return ""
}

func (b *Bot) runAsk(e *gateway.InteractionCreateEvent, question string, history []llm.Turn) {
	var userID discord.UserID
	if sender := e.Sender(); sender != nil {
		userID = sender.ID
	}

	if !b.limiter.Allow(userID.String()) {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("You're sending questions a bit too fast — give me a few seconds."),
			},
		})
		return
	}

	if question == "" {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("Please include a question."),
			},
		})
		return
	}

	if err := b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.DeferredMessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags: discord.EphemeralMessage | discord.SuppressEmbeds,
		},
	}); err != nil {
		slog.With("err", err).Warn("defer interaction failed")
		return
	}

	go b.replyAsk(e, question, history, userID)
}

func (b *Bot) replyAsk(e *gateway.InteractionCreateEvent, question string, history []llm.Turn, userID discord.UserID) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	var reply, intent string
	answer, err := b.llm.Answer(ctx, *b.instructions.Load(), history, question, userID.String())
	if err != nil {
		slog.With("err", err).Error("answer failed")
		reply = "Something went wrong while answering. Please try again in a moment."
	} else {
		reply = formatAnswer(answer)
		intent = answer.Intent
	}

	showFeedback := b.feedbackC != 0 && (intent == llm.IntentBug || intent == llm.IntentSuggestion)
	showHelp := b.helpC != 0
	msg, err := b.state.EditInteractionResponse(b.appID, e.Token, api.EditInteractionResponseData{
		Content:    option.NewNullableString(reply),
		Components: followupComponents(showFeedback, showHelp),
	})
	if err != nil {
		slog.With("err", err).Error("edit interaction response failed")
		return
	}
	if msg == nil {
		return
	}

	// After an error, a follow-up continues the conversation before it.
	turns := history
	if answer != nil {
		// Clipped so it doesn't write into the history of the previous answer.
		turns = append(slices.Clip(history), llm.Turn{Question: question, Answer: answer.Text})
		if len(turns) > maxHistory {
			turns = turns[len(turns)-maxHistory:]
		}
	} else if len(history) == 0 {
		return
	}
	b.feedback.Put(msg.ID, feedbackContext{
		Answer:    reply,
		Intent:    intent,
		History:   turns,
		UserID:    userID,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	})
}

// maxHistory is how many earlier questions and answers a follow-up is sent
// with.
const maxHistory = 3

var bareURL = regexp.MustCompile(`(?:^|[^<\w])(https?://[^\s<>)\]]+)`)

// formatAnswer adds links to the docs pages the answer is based on. Every link
// is wrapped in <> so Discord doesn't show a preview.
func formatAnswer(answer *llm.Answer) string {
	var links []string
	for _, l := range answer.Links[:min(len(answer.Links), 2)] {
		links = append(links, "<"+l+">")
	}

	text := bareURL.ReplaceAllStringFunc(strings.TrimSpace(answer.Text), func(m string) string {
		i := strings.Index(m, "http")
		return m[:i] + "<" + m[i:] + ">"
	})
	suffix := ""
	if len(links) > 0 {
		suffix = "\n\nMore: " + strings.Join(links, " ")
	}
	return truncate(text, 2000-len(suffix)) + suffix
}

func (b *Bot) handleFeedbackSubmit(e *gateway.InteractionCreateEvent, details string) {
	if b.feedbackC == 0 {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("Feedback isn't set up on this server."),
			},
		})
		return
	}

	var msgID discord.MessageID
	if e.Message != nil {
		msgID = e.Message.ID
	}
	fctx, ok := b.feedback.ClaimFeedback(msgID)
	if !ok {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("This answer was already sent as feedback or is too old. Ask the question again to send feedback."),
			},
		})
		return
	}

	userKey := "unknown"
	if sender := e.Sender(); sender != nil {
		userKey = sender.ID.String()
	}
	if !b.feedbackLimiter.Allow(userKey) {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("You've already sent feedback recently. Please wait a few minutes before sending more."),
			},
		})
		return
	}

	embed := buildFeedbackEmbed(fctx, details)
	threadName := buildThreadName(fctx)
	thread, err := b.createForumPost(b.feedbackC, forumThreadCreateData{
		Name:                threadName,
		AutoArchiveDuration: discord.OneDayArchive,
		Message: api.SendMessageData{
			Embeds: []discord.Embed{embed},
		},
	})
	if err != nil {
		slog.With("err", err).Error("post feedback failed")
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("Couldn't send your feedback. Please try again later."),
			},
		})
		return
	}

	if fctx.UserID != 0 {
		if err := b.state.AddThreadMember(thread.ID, fctx.UserID); err != nil {
			slog.With("err", err).Warn("add user to feedback thread failed")
		}
	}
	b.grantRole(e.GuildID, fctx.UserID, b.feedbackRoleID, "feedback")

	if err := b.state.ModifyChannel(thread.ID, api.ModifyChannelData{
		Locked: option.True,
	}); err != nil {
		slog.With("err", err).Warn("lock feedback thread failed")
	}

	if _, err := b.state.EditMessageComplex(e.ChannelID, msgID, api.EditMessageData{
		Components: followupComponents(false, false),
	}); err != nil {
		slog.With("err", err).Warn("strip feedback button failed")
	}

	_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags:   discord.EphemeralMessage,
			Content: option.NewNullableString(fmt.Sprintf("Thanks — your feedback was sent. We'll continue in <#%s>.", thread.ID)),
		},
	})
}

func buildThreadName(fctx feedbackContext) string {
	prefix := "Feedback"
	switch fctx.Intent {
	case llm.IntentBug:
		prefix = "Bug"
	case llm.IntentSuggestion:
		prefix = "Suggestion"
	}
	q := strings.ReplaceAll(fctx.question(), "\n", " ")
	q = strings.TrimSpace(q)
	name := fmt.Sprintf("[%s] %s", prefix, q)
	return truncate(name, 100)
}

func (b *Bot) handleHelpSubmit(e *gateway.InteractionCreateEvent, details string) {
	if b.helpC == 0 {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("Human help isn't set up on this server."),
			},
		})
		return
	}
	if details == "" {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("Please describe what you need help with."),
			},
		})
		return
	}

	var msgID discord.MessageID
	if e.Message != nil {
		msgID = e.Message.ID
	}
	fctx, _ := b.feedback.Get(msgID)

	userKey := "unknown"
	var userID discord.UserID
	if sender := e.Sender(); sender != nil {
		userKey = sender.ID.String()
		userID = sender.ID
	}
	if !b.helpLimiter.Allow(userKey) {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("You've already opened a help thread recently. Please wait a few minutes before opening another."),
			},
		})
		return
	}

	if fctx.UserID != 0 {
		userID = fctx.UserID
	}

	threadName := buildHelpThreadName(fctx, details)
	embed := buildHelpEmbed(fctx, details, userID)
	thread, err := b.createForumPost(b.helpC, forumThreadCreateData{
		Name:                threadName,
		AutoArchiveDuration: discord.OneDayArchive,
		Message: api.SendMessageData{
			Embeds: []discord.Embed{embed},
		},
	})
	if err != nil {
		slog.With("err", err).Error("create help thread failed")
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("Couldn't open a help thread. Please try again later."),
			},
		})
		return
	}

	if userID != 0 {
		if err := b.state.AddThreadMember(thread.ID, userID); err != nil {
			slog.With("err", err).Warn("add user to help thread failed")
		}
	}
	b.grantRole(e.GuildID, userID, b.helpRoleID, "help")

	if _, err := b.state.EditMessageComplex(e.ChannelID, msgID, api.EditMessageData{
		Components: followupComponents(false, false),
	}); err != nil {
		slog.With("err", err).Warn("strip help button failed")
	}

	_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags:   discord.EphemeralMessage,
			Content: option.NewNullableString(fmt.Sprintf("Opened a help thread for you in <#%s>.", thread.ID)),
		},
	})
}

func buildHelpThreadName(fctx feedbackContext, details string) string {
	source := fctx.question()
	if source == "" {
		source = details
	}
	source = strings.ReplaceAll(source, "\n", " ")
	source = strings.TrimSpace(source)
	return truncate(fmt.Sprintf("[Help] %s", source), 100)
}

func buildHelpEmbed(fctx feedbackContext, details string, userID discord.UserID) discord.Embed {
	reporter := "unknown"
	if userID != 0 {
		reporter = userID.Mention()
	}
	fields := []discord.EmbedField{
		{Name: "From", Value: reporter, Inline: true},
		{Name: "Need help with", Value: truncate(details, 1024)},
	}
	if fctx.question() != "" {
		fields = append(fields, discord.EmbedField{Name: "Original question to bot", Value: truncate(fctx.question(), 1024)})
	}
	if fctx.Answer != "" {
		fields = append(fields, discord.EmbedField{Name: "Bot's answer", Value: truncate(fctx.Answer, 1024)})
	}
	return discord.Embed{
		Title:     "Help requested",
		Color:     discord.Color(0x2ECC71),
		Fields:    fields,
		Timestamp: discord.NowTimestamp(),
	}
}

func (b *Bot) grantRole(guildID discord.GuildID, userID discord.UserID, roleID discord.RoleID, label string) {
	if guildID == 0 || userID == 0 || roleID == 0 {
		return
	}
	if err := b.state.AddRole(guildID, userID, roleID, api.AddRoleData{
		AuditLogReason: api.AuditLogReason("kite-support: granted on " + label + " thread creation"),
	}); err != nil {
		slog.With("err", err).Warn("grant role failed", "label", label)
	}
}

func buildFeedbackEmbed(fctx feedbackContext, details string) discord.Embed {
	title := "Feedback"
	color := 0x95A5A6
	switch fctx.Intent {
	case llm.IntentBug:
		title = "Bug report"
		color = 0xE74C3C
	case llm.IntentSuggestion:
		title = "Suggestion"
		color = 0x3498DB
	}

	reporter := "unknown"
	if fctx.UserID != 0 {
		reporter = fctx.UserID.Mention()
	}

	fields := []discord.EmbedField{
		{Name: "From", Value: reporter, Inline: true},
		{Name: "Question", Value: truncate(fctx.question(), 1024)},
		{Name: "Bot answer", Value: truncate(fctx.Answer, 1024)},
	}
	if details != "" {
		fields = append(fields, discord.EmbedField{Name: "Details", Value: truncate(details, 1024)})
	}

	return discord.Embed{
		Title:     title,
		Color:     discord.Color(color),
		Fields:    fields,
		Timestamp: discord.NowTimestamp(),
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	const ellipsis = "…"
	cut := n - len(ellipsis)
	if cut < 0 {
		cut = 0
	}
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + ellipsis
}

func (b *Bot) handleSetupMenu(e *gateway.InteractionCreateEvent) {
	if e.ChannelID == 0 {
		_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Flags:   discord.EphemeralMessage,
				Content: option.NewNullableString("This command must be run in a channel."),
			},
		})
		return
	}

	_, err := b.state.SendMessageComplex(e.ChannelID, api.SendMessageData{
		Embeds:     []discord.Embed{menuEmbed()},
		Components: *menuComponents(),
	})

	content := "Menu posted."
	if err != nil {
		slog.With("err", err).Error("post menu failed")
		content = fmt.Sprintf("Failed to post menu: %v", err)
	}
	_ = b.state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Flags:   discord.EphemeralMessage,
			Content: option.NewNullableString(content),
		},
	})
}
