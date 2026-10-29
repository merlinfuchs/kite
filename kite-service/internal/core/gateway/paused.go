package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// paused reports whether the app is stopped for credits. Its gateway stays
// connected so interactions get a reply saying why, but it runs no flows.
func (g *Gateway) paused() bool {
	return model.HasNoCredits(g.currentApp())
}

// respondPaused tells the user why the bot doesn't respond. The owner and the
// collaborators of the app also get when it starts again and where to upgrade.
func (g *Gateway) respondPaused(session *state.State, e *gateway.InteractionCreateEvent) {
	var resp api.InteractionResponse
	switch e.Data.(type) {
	case *discord.PingInteraction:
		return
	case *discord.AutocompleteInteraction:
		resp = api.InteractionResponse{
			Type: api.AutocompleteResult,
			Data: &api.InteractionResponseData{Choices: api.AutocompleteStringChoices{}},
		}
	default:
		resp = api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content: option.NewNullableString(g.pausedMessage(e.SenderID(), time.Now().UTC())),
				Flags:   discord.EphemeralMessage,
			},
		}
	}

	if err := session.RespondInteraction(e.ID, e.Token, resp); err != nil {
		slog.Warn(
			"Failed to respond to interaction of paused app",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
	}
}

// pausedRestartWindow is how long after a month begins paused apps may still
// wait for the credit sweep to start them again.
const pausedRestartWindow = 10 * time.Minute

func (g *Gateway) pausedMessage(userID discord.UserID, now time.Time) string {
	start, _ := util.StartAndEndOfMonth(now)
	if now.Sub(start) < pausedRestartWindow {
		return "This bot is paused because it used all of its credits for last month. It starts again in a few minutes."
	}
	resetDate := start.AddDate(0, 1, 0).Format("January 2")

	if g.isAppMember(userID) {
		return fmt.Sprintf(
			"This bot used all of its credits for this month, so its commands and events are paused. It starts again on %s (UTC), or a few minutes after you upgrade its plan: %s/apps/%s/premium",
			resetDate, g.appBaseURL, g.appID,
		)
	}
	return fmt.Sprintf("This bot is paused until %s because it used all of its credits for this month.", resetDate)
}

// isAppMember reports whether the user owns the app or collaborates on it. A
// failed lookup counts as no, which only leaves out the details.
func (g *Gateway) isAppMember(userID discord.UserID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ids, err := g.appStore.AppMemberDiscordIDs(ctx, g.appID)
	if err != nil {
		slog.Error(
			"Failed to get app members for paused reply",
			slog.String("app_id", g.appID),
			slog.String("error", err.Error()),
		)
		return false
	}
	return slices.Contains(ids, userID.String())
}
