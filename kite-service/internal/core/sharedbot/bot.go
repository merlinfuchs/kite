// Package sharedbot runs the shared apps on the Kite bot, one app per server.
// See design/kite-lite.md.
package sharedbot

import (
	"context"
	"fmt"
	"sync"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

// Permissions are what the Kite bot asks for when it's added to a server,
// what the templates that can run on it need.
const Permissions = discord.PermissionViewChannel |
	discord.PermissionSendMessages |
	discord.PermissionEmbedLinks |
	discord.PermissionAttachFiles |
	discord.PermissionReadMessageHistory |
	discord.PermissionAddReactions |
	discord.PermissionUseExternalEmojis |
	discord.PermissionManageMessages |
	discord.PermissionManageRoles |
	discord.PermissionKickMembers |
	discord.PermissionBanMembers |
	discord.PermissionModerateMembers

// Bot is the Kite bot.
type Bot struct {
	token string
	// base is the one client of the bot, so every request shares its rate
	// limits.
	base *api.Client

	mu    sync.Mutex
	appID discord.AppID
}

// New returns the bot with the token, or nil without one. Shared apps don't
// run then.
func New(token string) *Bot {
	if token == "" {
		return nil
	}
	return &Bot{token: token, base: api.NewClient("Bot " + token)}
}

// Client reaches every server of the bot, so it's never given to an app.
func (b *Bot) Client(ctx context.Context) *api.Client {
	return b.base.WithContext(ctx)
}

// AppID is the bot's Discord application, which commands are deployed to.
func (b *Bot) AppID(ctx context.Context) (discord.AppID, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.appID.IsValid() {
		return b.appID, nil
	}

	app, err := b.Client(ctx).CurrentApplication()
	if err != nil {
		return 0, fmt.Errorf("failed to get the Kite bot's application: %w", err)
	}
	b.appID = app.ID
	return b.appID, nil
}
