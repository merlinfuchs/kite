package flowai

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// ErrRefused is returned by Respond for chats that ask for a bot that attacks
// Discord servers or their members. They are refused before they reach
// OpenAI, which warns and restricts accounts whose users ask for that.
var ErrRefused = errors.New("flow AI refused the prompt")

// screenModel is the decision model that checks chats. It's pinned apart from
// the AI blocks' model, as screenThreshold is tuned for it.
const screenModel = "typesafe/jev-1.14"

// screenThreshold is the chance of harm from which a chat is refused. In
// tests, prompts and flows to nuke or raid servers, phish or steal tokens
// scored 0.94 and up, and ones for moderation, anti-nuke, tickets and
// announcements 0.6 at most.
const screenThreshold = 0.9

// screenTimeout is short, as a check that fails lets the chat through anyway.
const screenTimeout = 5 * time.Second

const screenQuestion = "Is this person using a Discord bot builder to make a bot that attacks or harms Discord servers or users? " +
	"Yes means things like nuking or raiding servers (mass creating or deleting channels, roles or bans, spam with @everyone), " +
	"phishing or scams, stealing accounts, tokens or personal data, or self-bots. " +
	"Bots that protect servers, like anti-nuke or anti-raid, and normal moderation, tickets, announcements and staff-only bulk actions are No."

// refuses checks whether the chat asks for a bot that attacks Discord servers
// or their members. If the check fails, the chat isn't refused, so the flow AI
// works without the decision model.
func (a *Assistant) refuses(ctx context.Context, req Request) bool {
	if a.config.Decisions == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, screenTimeout)
	defer cancel()

	decision, err := a.config.Decisions.Decide(ctx, provider.DecideOpts{
		Model:    screenModel,
		Type:     provider.DecisionTypeNoul,
		Question: screenQuestion,
		Options:  map[string]string{"true": "Yes", "false": "No"},
		Input:    screenInput(req),
	})
	if err != nil {
		// Not when the user left, as the answer won't be made either.
		if !errors.Is(ctx.Err(), context.Canceled) {
			slog.Error("Failed to screen flow AI prompt", slog.String("app_id", req.AppID), slog.Any("error", err))
		}
		return false
	}
	if decision.Noul < screenThreshold {
		return false
	}

	slog.Warn(
		"Flow AI refused a prompt",
		slog.String("app_id", req.AppID),
		slog.String("user_id", req.UserID),
		slog.String("prompt_id", req.PromptID),
		slog.Float64("chance", decision.Noul),
	)
	return true
}

// screenInput is everything of the chat the model would see that the user
// wrote, as earlier messages and the flow can show what a harmless looking
// message is about.
func screenInput(req Request) string {
	var b strings.Builder
	for _, m := range req.Messages {
		if m.Role == "user" && m.Content != "" {
			b.WriteString("Request: ")
			b.WriteString(m.Content)
			b.WriteString("\n\n")
		}
	}
	if len(req.Issues) > 0 {
		b.WriteString("Problems to fix:\n- ")
		b.WriteString(strings.Join(req.Issues, "\n- "))
		b.WriteString("\n\n")
	}
	if req.Flow != "" {
		b.WriteString("Current flow:\n")
		b.WriteString(req.Flow)
	}
	return b.String()
}
