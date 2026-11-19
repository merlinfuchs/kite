package sharedbot

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// AppStateManager is the gateway manager's, with the shared apps added. The
// API runs in the clusters, which have no connection of the Kite bot, so a
// shared app's state comes from Discord's REST API.
type AppStateManager struct {
	next     store.AppStateManager
	appStore store.AppStore
	bot      *Bot
	channels *channelGuilds
}

func NewAppStateManager(next store.AppStateManager, appStore store.AppStore, bot *Bot) *AppStateManager {
	m := &AppStateManager{next: next, appStore: appStore, bot: bot}
	if bot != nil {
		m.channels = newChannelGuilds(nil, bot.base)
	}
	return m
}

// client is the app's view of the bot, see scopedClient.
func (m *AppStateManager) client(ctx context.Context, app *model.App) (*api.Client, discord.GuildID, error) {
	guildID, err := discord.ParseSnowflake(app.DiscordGuildID.String)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid server of shared app: %w", err)
	}
	botAppID, err := m.bot.AppID(ctx)
	if err != nil {
		return nil, 0, err
	}

	s := &scope{guildID: discord.GuildID(guildID), botAppID: botAppID, channels: m.channels}
	return scopedClient(m.bot.base, s).WithContext(ctx), discord.GuildID(guildID), nil
}

// sharedApp returns the app if it's a shared one. Own apps on this cluster
// have a gateway, so they never get here.
func (m *AppStateManager) sharedApp(ctx context.Context, appID string) (*model.App, error) {
	app, err := m.appStore.App(ctx, appID)
	if err != nil {
		return nil, err
	}
	if !app.IsShared() || m.bot == nil {
		return nil, store.ErrNotFound
	}
	return app, nil
}

func (m *AppStateManager) AppState(ctx context.Context, appID string) (store.AppStateStore, error) {
	state, err := m.next.AppState(ctx, appID)
	if !errors.Is(err, store.ErrNotFound) {
		return state, err
	}

	app, err := m.sharedApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	client, guildID, err := m.client(ctx, app)
	if err != nil {
		return nil, err
	}
	return &appState{client: client, guildID: guildID}, nil
}

func (m *AppStateManager) AppClient(ctx context.Context, appID string) (*api.Client, error) {
	client, err := m.next.AppClient(ctx, appID)
	if !errors.Is(err, store.ErrNotFound) {
		return client, err
	}

	app, err := m.sharedApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	client, _, err = m.client(ctx, app)
	return client, err
}

// appState is the Kite bot as one shared app sees it. The app only sees its
// own server, even though the bot is in many.
type appState struct {
	client  *api.Client
	guildID discord.GuildID
}

// ownGuild parses the guild ID and reports whether it's the app's server.
func (s *appState) ownGuild(guildID string) bool {
	id, err := discord.ParseSnowflake(guildID)
	return err == nil && discord.GuildID(id) == s.guildID
}

func (s *appState) AppStatus(ctx context.Context) (store.AppStateStatus, error) {
	// Whether the Kite bot is connected is only known in its own process.
	return store.AppStateStatus{Online: true}, nil
}

func (s *appState) AppUser(ctx context.Context) (*discord.User, error) {
	return s.client.WithContext(ctx).Me()
}

func (s *appState) AppGuilds(ctx context.Context) ([]discord.Guild, error) {
	guild, err := s.client.WithContext(ctx).Guild(s.guildID)
	if isNotInGuild(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return []discord.Guild{*guild}, nil
}

func (s *appState) AppGuildChannels(ctx context.Context, guildID string) ([]discord.Channel, error) {
	if !s.ownGuild(guildID) {
		return nil, nil
	}
	channels, err := s.client.WithContext(ctx).Channels(s.guildID)
	if isNotInGuild(err) {
		return nil, nil
	}
	return channels, err
}

func (s *appState) AppGuildRoles(ctx context.Context, guildID string) ([]discord.Role, error) {
	if !s.ownGuild(guildID) {
		return nil, nil
	}
	roles, err := s.client.WithContext(ctx).Roles(s.guildID)
	if isNotInGuild(err) {
		return nil, nil
	}
	return roles, err
}

func (s *appState) AppGuildEmojis(ctx context.Context, guildID string) ([]discord.Emoji, error) {
	if !s.ownGuild(guildID) {
		return nil, nil
	}
	emojis, err := s.client.WithContext(ctx).Emojis(s.guildID)
	if isNotInGuild(err) {
		return nil, nil
	}
	return emojis, err
}

func (s *appState) AppDiscordAPIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error) {
	return util.DiscordAPIRequest(ctx, s.client.WithContext(ctx), req)
}

func (s *appState) AppGuildBotPermissions(ctx context.Context, guildID string) (discord.Permissions, error) {
	if !s.ownGuild(guildID) {
		return 0, store.ErrNotFound
	}

	client := s.client.WithContext(ctx)
	me, err := client.Me()
	if err != nil {
		return 0, err
	}
	guild, err := client.Guild(s.guildID)
	if err != nil {
		return 0, err
	}
	if guild.OwnerID == me.ID {
		return ^discord.Permissions(0), nil
	}
	member, err := client.Member(s.guildID, me.ID)
	if err != nil {
		return 0, err
	}

	// The @everyone role has the ID of the guild.
	var perms discord.Permissions
	for _, role := range guild.Roles {
		if discord.Snowflake(role.ID) == discord.Snowflake(s.guildID) || slices.Contains(member.RoleIDs, role.ID) {
			perms |= role.Permissions
		}
	}
	if perms.Has(discord.PermissionAdministrator) {
		return ^discord.Permissions(0), nil
	}
	return perms, nil
}

func (s *appState) AppGuildBotMember(ctx context.Context, guildID string) (*provider.StyledMember, error) {
	if !s.ownGuild(guildID) {
		return nil, store.ErrNotFound
	}

	client := s.client.WithContext(ctx)
	me, err := client.Me()
	if err != nil {
		return nil, err
	}

	var member provider.StyledMember
	err = client.RequestJSON(&member, "GET", api.EndpointGuilds+s.guildID.String()+"/members/"+me.ID.String())
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func isNotInGuild(err error) bool {
	return util.IsDiscordRestStatusCode(err, http.StatusForbidden) ||
		util.IsDiscordRestStatusCode(err, http.StatusNotFound)
}
