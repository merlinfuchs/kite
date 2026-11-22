package gateway

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// paused reports whether the app is paused for credits. Its gateway stays
// connected so interactions get a reply saying why, but it runs no flows.
func (g *Gateway) paused() bool {
	return model.HasNoCredits(g.currentApp())
}

func (g *Gateway) respondPaused(session *state.State, e *gateway.InteractionCreateEvent) {
	RespondPaused(session, e, g.currentApp(), g.appBaseURL)
}

// RespondPaused tells the user why the bot of the paused app doesn't respond
// and when it starts again.
func RespondPaused(session *state.State, e *gateway.InteractionCreateEvent, app *model.App, appBaseURL string) {
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
				Content: option.NewNullableString(pausedMessage(app, time.Now().UTC(), appBaseURL)),
				Flags:   discord.EphemeralMessage,
			},
		}
	}

	if err := session.RespondInteraction(e.ID, e.Token, resp); err != nil {
		slog.Warn(
			"Failed to respond to interaction of paused app",
			slog.String("app_id", app.ID),
			slog.String("error", err.Error()),
		)
	}
}

func pausedMessage(app *model.App, now time.Time, appBaseURL string) string {
	footer := fmt.Sprintf("\n-# Made with [Kite](<%s>), the no-code Discord bot builder", appBaseURL)

	start, _ := util.StartAndEndOfMonth(now)
	if app.CreditsPausedAt.Time.Before(start) {
		// The month it was paused in is over, the credit sweep unpauses it next.
		return "This bot is paused because it used all of its credits for last month. It starts again in a few minutes." + footer
	}
	return fmt.Sprintf(
		"This bot used all of its credits for this month, so its commands and events are paused until %s (UTC). Its owner can start it again sooner by upgrading its plan on Kite.",
		start.AddDate(0, 1, 0).Format("January 2"),
	) + footer
}
