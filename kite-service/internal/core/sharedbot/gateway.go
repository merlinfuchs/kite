package sharedbot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	kitegateway "github.com/kitecloud/kite/kite-service/internal/core/gateway"
	"github.com/kitecloud/kite/kite-service/internal/metrics"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

// Intents are what the Kite bot identifies with. It has no privileged intents
// yet, so Lite modules can't need them.
const Intents = gateway.IntentGuilds |
	gateway.IntentGuildMessages |
	gateway.IntentGuildMessageReactions

// presence is the Kite bot's status, the same in every server. It has to fit
// the member list, which cuts it off after about 25 characters.
var presence = gateway.UpdatePresenceCommand{
	Status: discord.OnlineStatus,
	Activities: []discord.Activity{{
		Type:  discord.CustomActivity,
		Name:  "kite.onl",
		State: "🪁 Make a bot at kite.onl",
	}},
}

type GatewayConfig struct {
	PopulateInterval time.Duration
	// AppBaseURL is where the Kite app runs, for the link in the reply of
	// paused apps.
	AppBaseURL string
}

// Gateway is the Kite bot's connection to Discord. It sends each event to the
// app of the server it happened in, and drops the events of servers without
// one.
type Gateway struct {
	config       GatewayConfig
	bot          *Bot
	appStore     store.AppStore
	eventHandler kitegateway.EventHandler
	session      *state.State
	botAppID     discord.AppID
	channels     *channelGuilds
	// failures outlive the views, which are made again after reconnects.
	failures *appFailures
	// populateMu keeps populates in order, so an older read can't replace a
	// newer one.
	populateMu sync.Mutex

	mu sync.Mutex
	// apps are the enabled shared apps by their server, and guilds their
	// servers by app ID.
	apps   map[discord.GuildID]*model.App
	guilds map[string]discord.GuildID
	// states are the views of the session of the apps, by app ID. They're
	// made again after the bot reconnects, as they copy its Ready event.
	states map[string]*state.State
	// intents are what the templates of the apps read, by app ID.
	intents map[string]*appIntents
}

func NewGateway(bot *Bot, appStore store.AppStore, eventHandler kitegateway.EventHandler, config GatewayConfig) *Gateway {
	identifier := gateway.DefaultIdentifier("Bot " + bot.token)
	identifier.IdentifyCommand.Presence = &presence
	session := state.NewWithIdentifier(identifier)
	session.AddIntents(Intents)
	// Before the views of the apps copy the cache.
	kitegateway.TrackMemberCounts(session)

	return &Gateway{
		config:       config,
		bot:          bot,
		appStore:     appStore,
		eventHandler: eventHandler,
		session:      session,
		channels:     newChannelGuilds(session.Cabinet.ChannelStore, session.Client),
		failures:     newAppFailures(maxGatewayFailures),
		apps:         make(map[discord.GuildID]*model.App),
		guilds:       make(map[string]discord.GuildID),
		states:       make(map[string]*state.State),
		intents:      make(map[string]*appIntents),
	}
}

// Run loads the apps before it connects, so the first events have an app to
// go to.
func (g *Gateway) Run(ctx context.Context) error {
	appID, err := g.bot.AppID(ctx)
	if err != nil {
		return err
	}
	g.botAppID = appID

	if err := g.populate(ctx); err != nil {
		return err
	}

	g.session.AddHandler(func(e gateway.Event) {
		// Protocol frames have no type, see Gateway.startGateway.
		if e.EventType() == "" {
			return
		}
		metrics.GatewayEvents.Add(string(e.EventType()), 1)
		g.dispatch(e)
	})
	kitegateway.TrackDerivedEvents(g.session, g.dispatch)

	g.session.AddHandler(func(e *state.GuildJoinEvent) {
		go g.guildJoined(ctx, e.ID)
	})
	g.session.AddHandler(func(e *state.GuildLeaveEvent) {
		go g.guildLeft(ctx, e.ID)
	})
	// The name of an app is its server's, also after the server is renamed.
	// Servers come with a create event after every connect, which catches
	// renames while the bot was offline.
	g.session.AddHandler(func(e *gateway.GuildCreateEvent) {
		go g.guildRenamed(ctx, e.ID, e.Name)
	})
	g.session.AddHandler(func(e *gateway.GuildUpdateEvent) {
		go g.guildRenamed(ctx, e.ID, e.Name)
	})

	g.session.AddHandler(func(e *gateway.ReadyEvent) {
		// Servers that removed the bot while it was offline.
		inGuild := make(map[discord.GuildID]bool, len(e.Guilds))
		for _, guild := range e.Guilds {
			inGuild[guild.ID] = true
		}
		g.mu.Lock()
		clear(g.states)
		var removed []discord.GuildID
		for guildID := range g.apps {
			if !inGuild[guildID] {
				removed = append(removed, guildID)
			}
		}
		g.mu.Unlock()
		for _, guildID := range removed {
			go g.guildLeft(ctx, guildID)
		}

		slog.Info(
			"Kite bot received ready event",
			slog.String("user_id", e.User.ID.String()),
			slog.Int("guilds", len(e.Guilds)),
		)
	})

	go func() {
		ticker := time.NewTicker(util.IntervalOrDefault(g.config.PopulateInterval, 5*time.Second))
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := g.populate(ctx); err != nil {
					slog.Error("Failed to populate shared apps", slog.String("error", err.Error()))
				}
			}
		}
	}()

	go func() {
		if err := g.session.Connect(ctx); err != nil && ctx.Err() == nil {
			slog.Error("Kite bot failed to connect to the gateway", slog.String("error", err.Error()))
		}
	}()

	return nil
}

// populate reloads every enabled shared app. There are few, so this is
// cheaper than tracking changes like the gateway manager.
func (g *Gateway) populate(ctx context.Context) error {
	g.populateMu.Lock()
	defer g.populateMu.Unlock()

	apps, err := g.appStore.EnabledSharedApps(ctx)
	if err != nil {
		return fmt.Errorf("failed to get shared apps: %w", err)
	}
	bundleIDs, err := g.appStore.SharedAppBundleIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bundles of shared apps: %w", err)
	}

	byGuild := make(map[discord.GuildID]*model.App, len(apps))
	guilds := make(map[string]discord.GuildID, len(apps))
	for _, app := range apps {
		guildID, err := discord.ParseSnowflake(app.DiscordGuildID.String)
		if err != nil {
			slog.Error("Shared app has an invalid server", slog.String("app_id", app.ID))
			continue
		}
		byGuild[discord.GuildID(guildID)] = app
		guilds[app.ID] = discord.GuildID(guildID)
	}

	g.mu.Lock()
	g.apps = byGuild
	g.guilds = guilds
	for appID := range g.states {
		if _, ok := guilds[appID]; !ok {
			delete(g.states, appID)
		}
	}
	for appID := range g.intents {
		if _, ok := guilds[appID]; !ok {
			delete(g.intents, appID)
		}
	}
	for appID := range guilds {
		g.intentsOf(appID).set(bundleIDs[appID])
	}
	g.mu.Unlock()
	return nil
}

// guildJoined starts the app of a server that added the bot again. New
// servers get their app when they're set up, see the install of the API.
func (g *Gateway) guildJoined(ctx context.Context, guildID discord.GuildID) {
	started, err := g.appStore.EnableSharedAppDisabledWithReason(ctx, guildID.String(), model.AppDisabledReasonSharedBotRemoved, time.Now())
	if err != nil {
		slog.Error("Failed to start the app of a server that added the Kite bot", slog.String("guild_id", guildID.String()), slog.String("error", err.Error()))
		return
	}
	if started {
		if err := g.populate(ctx); err != nil {
			slog.Error("Failed to populate shared apps", slog.String("error", err.Error()))
		}
	}
}

// guildLeft stops the app of a server that removed the bot, and keeps it, so
// adding the bot again starts it with everything it had.
func (g *Gateway) guildLeft(ctx context.Context, guildID discord.GuildID) {
	app := g.appOfGuild(guildID)
	if app == nil {
		return
	}

	_, err := g.appStore.DisableApp(ctx, store.AppDisableOpts{
		ID:             app.ID,
		DisabledReason: null.StringFrom(model.AppDisabledReasonSharedBotRemoved),
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		slog.Error("Failed to stop the app of a server that removed the Kite bot", slog.String("app_id", app.ID), slog.String("error", err.Error()))
		return
	}
	if err := g.populate(ctx); err != nil {
		slog.Error("Failed to populate shared apps", slog.String("error", err.Error()))
	}
}

// guildRenamed gives the app of the server the server's name, if it has
// another one.
func (g *Gateway) guildRenamed(ctx context.Context, guildID discord.GuildID, name string) {
	app := g.appOfGuild(guildID)
	if app == nil || name == "" || app.Name == name {
		return
	}
	if err := g.appStore.RenameSharedApp(ctx, guildID.String(), name, time.Now()); err != nil {
		slog.Error("Failed to rename the app of a server", slog.String("app_id", app.ID), slog.String("error", err.Error()))
	}
}

func (g *Gateway) appOfGuild(guildID discord.GuildID) *model.App {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.apps[guildID]
}

// intentsOf returns the intents of the app, which its scopes share. g.mu must
// be held.
func (g *Gateway) intentsOf(appID string) *appIntents {
	intents, ok := g.intents[appID]
	if !ok {
		intents = &appIntents{}
		g.intents[appID] = intents
	}
	return intents
}

// stateFor returns the app's view of the session, see scopedState. g.mu
// must be held.
func (g *Gateway) stateFor(app *model.App, guildID discord.GuildID) *state.State {
	if s, ok := g.states[app.ID]; ok {
		return s
	}
	s := scopedState(g.session, &scope{guildID: guildID, botAppID: g.botAppID, channels: g.channels, failures: g.failures.of(app.ID), all: g.failures.all, intents: g.intentsOf(app.ID)})
	g.states[app.ID] = s
	return s
}

func (g *Gateway) dispatch(e gateway.Event) {
	guildID := engine.EventGuildID(e)
	if !guildID.IsValid() {
		metrics.GatewayEventsDropped.Add("shared_no_guild", 1)
		return
	}

	g.mu.Lock()
	app := g.apps[guildID]
	var session *state.State
	var intents *appIntents
	if app != nil {
		intents = g.intentsOf(app.ID)
		if !model.HasNoCredits(app) {
			session = g.stateFor(app, guildID)
		}
	}
	g.mu.Unlock()

	if app == nil {
		metrics.GatewayEventsDropped.Add("shared_no_app", 1)
		return
	}
	e, ok := filterEvent(e, intents, discord.UserID(g.botAppID))
	if !ok {
		metrics.GatewayEventsDropped.Add("shared_no_intent", 1)
		return
	}
	if session == nil {
		if e, ok := e.(*gateway.InteractionCreateEvent); ok {
			go kitegateway.RespondPaused(g.session, e, app, g.config.AppBaseURL)
		}
		return
	}

	g.eventHandler.HandleEvent(app.ID, session, e)
}

// AppIDs returns the shared apps that run flows, for the scheduler.
func (g *Gateway) AppIDs() []string {
	g.mu.Lock()
	defer g.mu.Unlock()

	ids := make([]string, 0, len(g.apps))
	for _, app := range g.apps {
		if !model.HasNoCredits(app) {
			ids = append(ids, app.ID)
		}
	}
	return ids
}

// AppSession is the app's view of the Kite bot's session, for the shared
// apps that run flows.
func (g *Gateway) AppSession(ctx context.Context, appID string) (*state.State, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	guildID, ok := g.guilds[appID]
	if app := g.apps[guildID]; ok && !model.HasNoCredits(app) {
		return g.stateFor(app, guildID), nil
	}
	return nil, store.ErrNotFound
}
