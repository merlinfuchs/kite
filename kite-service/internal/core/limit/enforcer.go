package limit

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

// GracePeriod is how long an app can have more enabled commands or event
// listeners than its plan allows, like after its plan ended, before the
// newest ones are disabled. Disabled ones aren't enabled again on an upgrade.
const GracePeriod = 3 * 24 * time.Hour

const interval = time.Hour

// maxWarningsPerRun spreads the DMs to owners over runs, like on the first
// run when many apps are over at once. The others are warned in later runs,
// and their grace period starts then.
const maxWarningsPerRun = 50

// Notifier tells the owner of an app about its limits. details continues a
// sentence that starts with the app, like "Your app ".
type Notifier interface {
	OverLimit(app *model.App, details string)
	LimitEnforced(app *model.App, details string)
}

type Features interface {
	DefaultFeatures() model.Features
	AppFeaturesForApps(ctx context.Context, appIDs []string) (map[string]model.Features, error)
}

// Enforcer disables the newest commands and event listeners of apps that have
// more enabled than their plan allows for longer than the GracePeriod. Only
// apps that had a plan are enforced, so free apps that were over before the
// free limits were lowered keep what they have.
type Enforcer struct {
	appStore           store.AppStore
	commandStore       store.CommandStore
	eventListenerStore store.EventListenerStore
	logStore           store.LogStore
	features           Features
	notifier           Notifier
}

func NewEnforcer(
	appStore store.AppStore,
	commandStore store.CommandStore,
	eventListenerStore store.EventListenerStore,
	logStore store.LogStore,
	features Features,
	notifier Notifier,
) *Enforcer {
	return &Enforcer{
		appStore:           appStore,
		commandStore:       commandStore,
		eventListenerStore: eventListenerStore,
		logStore:           logStore,
		features:           features,
		notifier:           notifier,
	}
}

// Run enforces the limits once right away, so restarts don't delay it, and
// then every interval.
func (e *Enforcer) Run(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			if err := e.Enforce(ctx, time.Now().UTC()); err != nil {
				slog.Error(
					"Failed to enforce limits",
					slog.String("error", err.Error()),
				)
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// limitOf returns the limit of the commands, or of the event listeners of
// source, 0 when there is none, and their name in the dashboard.
func limitOf(f model.Features, source model.EventSource) (int, string) {
	if source == "" {
		return f.MaxCommands, "commands"
	}
	return EventListenerLimit(f, source)
}

// exceeded is a limit of an app. count is how many enabled items it has, or
// how many were disabled.
type exceeded struct {
	// source is empty for commands.
	source model.EventSource
	count  int
	limit  int
}

func (x exceeded) noun() string {
	if x.source == "" {
		return "commands"
	}
	_, noun := EventListenerLimit(model.Features{}, x.source)
	return noun
}

// sources are the kinds of items with their own limit, commands first.
var sources = []model.EventSource{"", model.EventSourceDiscord, model.EventSourceSchedule, model.EventSourceWebhook}

type appSource struct {
	appID  string
	source model.EventSource
}

// Enforce warns the owners of apps that went over their limits, and disables
// the newest commands and event listeners of apps that have been over them
// for the GracePeriod.
func (e *Enforcer) Enforce(ctx context.Context, now time.Time) error {
	counts, err := e.countsOverDefaults(ctx)
	if err != nil {
		return err
	}

	appIDs := make([]string, 0, len(counts))
	seen := map[string]bool{}
	for k := range counts {
		if !seen[k.appID] {
			seen[k.appID] = true
			appIDs = append(appIDs, k.appID)
		}
	}

	features := map[string]model.Features{}
	if len(appIDs) > 0 {
		features, err = e.features.AppFeaturesForApps(ctx, appIDs)
		if err != nil {
			return fmt.Errorf("failed to get app features: %w", err)
		}
	}

	over := map[string][]exceeded{}
	for _, appID := range appIDs {
		for _, source := range sources {
			count, ok := counts[appSource{appID, source}]
			if !ok {
				continue
			}
			if limit, _ := limitOf(features[appID], source); limit > 0 && count > limit {
				over[appID] = append(over[appID], exceeded{source: source, count: count, limit: limit})
			}
		}
	}
	overIDs := make([]string, 0, len(over))
	for appID := range over {
		overIDs = append(overIDs, appID)
	}

	// Apps that are under their limits again, by an upgrade or by disabling
	// some themselves, get a new grace period if they go over again.
	if err := e.appStore.ClearAppsOverLimit(ctx, overIDs); err != nil {
		return fmt.Errorf("failed to clear apps over limit: %w", err)
	}
	if len(overIDs) == 0 {
		return nil
	}

	since, err := e.appStore.AppsOverLimitSince(ctx, overIDs)
	if err != nil {
		return fmt.Errorf("failed to get apps over limit: %w", err)
	}

	warnings := 0
	for _, appID := range overIDs {
		s, marked := since[appID]
		if !marked {
			if warnings < maxWarningsPerRun {
				warnings++
				e.warn(ctx, appID, over[appID], now)
			}
		} else if now.Sub(s) >= GracePeriod {
			e.disable(ctx, appID, over[appID], now)
		}
	}

	return nil
}

// countsOverDefaults returns the enabled counts that are over the limits of
// the default plan. No plan allows less, so apps under them can't be over
// theirs.
func (e *Enforcer) countsOverDefaults(ctx context.Context) (map[appSource]int, error) {
	defaults := e.features.DefaultFeatures()
	counts := map[appSource]int{}

	if limit := defaults.MaxCommands; limit > 0 {
		commandCounts, err := e.commandStore.AppsWithEnabledCommandsOver(ctx, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to count enabled commands: %w", err)
		}
		for appID, count := range commandCounts {
			counts[appSource{appID, ""}] = count
		}
	}

	minLimit := 0
	for _, source := range sources[1:] {
		if l, _ := limitOf(defaults, source); l > 0 && (minLimit == 0 || l < minLimit) {
			minLimit = l
		}
	}
	if minLimit == 0 {
		return counts, nil
	}

	listenerCounts, err := e.eventListenerStore.AppsWithEnabledEventListenersOver(ctx, minLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to count enabled event listeners: %w", err)
	}
	for _, c := range listenerCounts {
		if limit, _ := limitOf(defaults, c.Source); limit > 0 && c.Count > limit {
			counts[appSource{c.AppID, c.Source}] = c.Count
		}
	}

	return counts, nil
}

func (e *Enforcer) warn(ctx context.Context, appID string, over []exceeded, now time.Time) {
	marked, err := e.appStore.MarkAppOverLimit(ctx, appID, now)
	if err != nil {
		slog.Error(
			"Failed to mark app over limit",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return
	}
	if !marked {
		return
	}

	details := overLimitDetails(over, now.Add(GracePeriod))
	e.createLogEntry(ctx, appID, model.LogLevelWarn, "Your app "+details)
	if app := e.app(ctx, appID); app != nil {
		e.notifier.OverLimit(app, details)
	}
}

func (e *Enforcer) disable(ctx context.Context, appID string, over []exceeded, now time.Time) {
	var disabled []exceeded
	for _, o := range over {
		var n int
		var err error
		if o.source == "" {
			n, err = e.commandStore.DisableNewestCommands(ctx, appID, o.limit, now)
		} else {
			n, err = e.eventListenerStore.DisableNewestEventListeners(ctx, appID, o.source, o.limit, now)
		}
		if err != nil {
			slog.Error(
				"Failed to disable items over limit",
				slog.String("app_id", appID),
				slog.String("kind", o.noun()),
				slog.String("error", err.Error()),
			)
			continue
		}
		if n > 0 {
			disabled = append(disabled, exceeded{source: o.source, count: n, limit: o.limit})
		}
	}
	if len(disabled) == 0 {
		return
	}

	details := enforcedDetails(disabled)
	e.createLogEntry(ctx, appID, model.LogLevelWarn, "Your app "+details)
	if app := e.app(ctx, appID); app != nil {
		e.notifier.LimitEnforced(app, details)
	}
}

// overLimitDetails is like "has 18 enabled commands, but its plan allows 15.
// ...", for a sentence that starts with the app.
func overLimitDetails(over []exceeded, disableAt time.Time) string {
	have := make([]string, len(over))
	allowed := make([]string, len(over))
	for i, o := range over {
		have[i] = fmt.Sprintf("%d enabled %s", o.count, o.noun())
		allowed[i] = fmt.Sprintf("%d %s", o.limit, o.noun())
	}
	if len(over) == 1 {
		allowed[0] = fmt.Sprint(over[0].limit)
	}

	return fmt.Sprintf(
		"has %s, but its plan allows %s. Disable or delete some, or upgrade its plan. Otherwise Kite disables the newest ones on %s (UTC).",
		join(have), join(allowed), disableAt.Format("January 2"),
	)
}

// enforcedDetails is like "had more enabled commands than its plan allows,
// so Kite disabled the 3 newest ones. ...", for a sentence that starts with
// the app.
func enforcedDetails(disabled []exceeded) string {
	nouns := make([]string, len(disabled))
	counts := make([]string, len(disabled))
	for i, d := range disabled {
		nouns[i] = d.noun()
		counts[i] = fmt.Sprintf("the %d newest %s", d.count, d.noun())
	}
	if len(disabled) == 1 {
		counts[0] = fmt.Sprintf("the %d newest ones", disabled[0].count)
		if disabled[0].count == 1 {
			counts[0] = "the newest one"
		}
	}

	them := "them"
	if len(disabled) == 1 && disabled[0].count == 1 {
		them = "it"
	}

	return fmt.Sprintf(
		"had more enabled %s than its plan allows, so Kite disabled %s. You can enable %s again after you upgrade its plan or disable others.",
		join(nouns), join(counts), them,
	)
}

// join lists the items like "a, b, and c".
func join(items []string) string {
	switch len(items) {
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

func (e *Enforcer) app(ctx context.Context, appID string) *model.App {
	app, err := e.appStore.App(ctx, appID)
	if err != nil {
		slog.Error(
			"Failed to get app for limit notice",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
		return nil
	}
	return app
}

func (e *Enforcer) createLogEntry(ctx context.Context, appID string, level model.LogLevel, message string) {
	err := e.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     appID,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to create limit log entry",
			slog.String("app_id", appID),
			slog.String("error", err.Error()),
		)
	}
}
