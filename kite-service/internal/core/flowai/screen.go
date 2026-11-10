package flowai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// screenModel is the decision model that checks prompts.
const screenModel = "typesafe/jev-1.14"

// screenThreshold is the chance of harm from which a prompt is refused. In
// tests, prompts and flows to nuke or raid servers, phish or steal tokens
// scored 0.94 and up, and ones for moderation, anti-nuke, tickets and
// announcements 0.6 at most.
const screenThreshold = 0.9

const screenTimeout = 10 * time.Second

const screenQuestion = "Is this person using a Discord bot builder to make a bot that attacks or harms Discord servers or users? " +
	"Yes means things like nuking or raiding servers (mass creating or deleting channels, roles or bans, spam with @everyone), " +
	"phishing or scams, stealing accounts, tokens or personal data, or self-bots. " +
	"Bots that protect servers, like anti-nuke or anti-raid, and normal moderation, tickets, announcements and staff-only bulk actions are No."

// Refuses checks whether the chat asks for a bot that attacks Discord servers
// or their members, so it can be refused before it reaches OpenAI, which
// warns and restricts accounts whose users ask for that.
func (a *Assistant) Refuses(ctx context.Context, req Request) (bool, error) {
	if a.config.Decisions == nil {
		return false, nil
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
		return false, fmt.Errorf("failed to screen prompt: %w", err)
	}
	if decision.Noul < screenThreshold {
		return false, nil
	}

	slog.Warn(
		"Flow AI refused a prompt",
		slog.String("app_id", req.AppID),
		slog.String("user_id", req.UserID),
		slog.String("prompt_id", req.PromptID),
		slog.Float64("chance", decision.Noul),
	)
	return true, nil
}

// screenInput is the user's messages and the flow, as earlier messages and
// the flow can show what a harmless looking message is about.
func screenInput(req Request) string {
	var b strings.Builder
	for _, m := range req.Messages {
		if m.Role == "user" && m.Content != "" {
			b.WriteString("Request: ")
			b.WriteString(m.Content)
			b.WriteString("\n\n")
		}
	}
	if req.Flow != "" {
		b.WriteString("Current flow:\n")
		b.WriteString(req.Flow)
	}
	return b.String()
}
