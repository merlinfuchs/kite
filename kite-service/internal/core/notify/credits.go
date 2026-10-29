package notify

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// AppOwnerNotifier tells the owners of apps about the credits of their app,
// with a DM from the app's own bot. If that fails, because the owner shares no
// server with their bot or doesn't allow its DMs, Kite's bot sends it.
//
// The terms allow these DMs as notices about the app, so they only say what
// happened and how to get the bot running, never advertise a plan.
type AppOwnerNotifier struct {
	userStore  store.UserStore
	tokenCrypt *util.SymmetricCrypt
	// fallback is Kite's bot, nil if there is none.
	fallback   *Notifier
	appBaseURL string
}

func NewAppOwnerNotifier(userStore store.UserStore, tokenCrypt *util.SymmetricCrypt, fallback *Notifier, appBaseURL string) *AppOwnerNotifier {
	return &AppOwnerNotifier{
		userStore:  userStore,
		tokenCrypt: tokenCrypt,
		fallback:   fallback,
		appBaseURL: appBaseURL,
	}
}

// CreditsLow tells the owner that the app used most of its credits.
func (n *AppOwnerNotifier) CreditsLow(app *model.App, used int, limit int, resetAt time.Time) {
	n.dm(app, fmt.Sprintf(
		"Your bot **%s** used %s of its %s credits this month. If it uses all of them, its commands and events pause until %s (UTC). See its usage: %s/apps/%s%s",
		escape(app.Name), formatCredits(used), formatCredits(limit), resetAt.Format("January 2"), n.appBaseURL, app.ID, n.footer(app),
	))
}

// CreditsUsedUp tells the owner that the app used all of its credits and is
// paused.
func (n *AppOwnerNotifier) CreditsUsedUp(app *model.App, limit int, resetAt time.Time) {
	n.dm(app, fmt.Sprintf(
		"Your bot **%s** used all of its %s credits this month, so its commands and events are paused. Anyone who uses it is told why. It starts again on %s (UTC), or a few minutes after you upgrade its plan: %s/apps/%s/premium%s",
		escape(app.Name), formatCredits(limit), resetAt.Format("January 2"), n.appBaseURL, app.ID, n.footer(app),
	))
}

func (n *AppOwnerNotifier) footer(app *model.App) string {
	return fmt.Sprintf("\n-# You get this because you own this app on Kite. Turn these messages off in its settings: <%s/apps/%s/settings>", n.appBaseURL, app.ID)
}

// dm sends the content to the owner of the app in the background.
func (n *AppOwnerNotifier) dm(app *model.App, content string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), sendTimeout)
		defer cancel()

		owner, err := n.userStore.User(ctx, app.OwnerUserID)
		if err != nil {
			slog.Error(
				"Failed to get owner of app for credits notice",
				slog.String("app_id", app.ID),
				slog.String("error", err.Error()),
			)
			return
		}
		ownerID, err := discord.ParseSnowflake(owner.DiscordID)
		if err != nil {
			return
		}

		token, err := n.tokenCrypt.DecryptString(app.DiscordToken)
		if err == nil {
			err = dm(api.NewClient("Bot "+token).WithContext(ctx), discord.UserID(ownerID), content)
		}
		if err == nil {
			return
		}

		slog.Info(
			"Failed to DM owner with their bot, using Kite's bot",
			slog.String("app_id", app.ID),
			slog.String("error", err.Error()),
		)
		n.fallback.DM(owner.DiscordID, content)
	}()
}

// formatCredits writes the number with commas, like 10,000.
func formatCredits(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
