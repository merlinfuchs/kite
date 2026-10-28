package plan

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
)

func (m *PlanManager) Run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Minute)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := m.handOutRoles(ctx); err != nil {
					slog.With("error", err).Error("failed to hand out roles")
				}
			}
		}
	}()
}

// handOutRoles gives users the role of their plan, and creators of templates
// in the marketplace the creator role, in the Kite server. Users who aren't in
// it get theirs once they join. Roles are never taken away.
func (m *PlanManager) handOutRoles(ctx context.Context) error {
	if m.config.DiscordBotToken == "" || m.config.DiscordGuildID == "" {
		return nil
	}

	guildID, err := discord.ParseSnowflake(m.config.DiscordGuildID)
	if err != nil {
		return fmt.Errorf("failed to parse discord guild ID: %w", err)
	}
	r := roleGranter{
		client:  api.NewClient("Bot " + m.config.DiscordBotToken),
		guildID: discord.GuildID(guildID),
	}

	subscriptions, err := m.subscriptionStore.AllSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all subscriptions: %w", err)
	}

	for _, sub := range subscriptions {
		if sub.Status != "active" {
			continue
		}

		if !sub.LemonsqueezyProductID.Valid {
			continue
		}

		plan := m.PlanByLemonSqueezyProductID(sub.LemonsqueezyProductID.String)
		if plan == nil || plan.DiscordRoleID == "" {
			continue
		}

		user, err := m.userStore.User(ctx, sub.UserID)
		if err != nil {
			slog.Error(
				"Failed to get user by ID",
				slog.String("error", err.Error()),
				slog.String("subscription_id", sub.ID),
			)
			continue
		}

		r.grant(user.DiscordID, plan.DiscordRoleID, "Kite subscription")
	}

	if m.config.DiscordCreatorRoleID != "" {
		creators, err := m.bundleStore.ListedBundleOwnerDiscordIDs(ctx)
		if err != nil {
			return fmt.Errorf("failed to get creators: %w", err)
		}
		for _, discordID := range creators {
			r.grant(discordID, m.config.DiscordCreatorRoleID, "Kite template creator")
		}
	}

	return nil
}

type roleGranter struct {
	client  *api.Client
	guildID discord.GuildID
}

// grant gives the user the role, unless they aren't in the guild.
func (r roleGranter) grant(discordUserID, roleID, reason string) {
	userID, err := discord.ParseSnowflake(discordUserID)
	if err != nil {
		slog.Error("Failed to parse discord user ID", slog.String("discord_user_id", discordUserID))
		return
	}
	role, err := discord.ParseSnowflake(roleID)
	if err != nil {
		slog.Error("Failed to parse discord role ID", slog.String("discord_role_id", roleID))
		return
	}

	err = r.client.AddRole(r.guildID, discord.UserID(userID), discord.RoleID(role), api.AddRoleData{
		AuditLogReason: api.AuditLogReason(reason),
	})
	var dErr *httputil.HTTPError
	if err != nil && (!errors.As(err, &dErr) || dErr.Status != 404) {
		slog.Error(
			"Failed to add role",
			slog.String("error", err.Error()),
			slog.String("discord_user_id", discordUserID),
			slog.String("discord_role_id", roleID),
		)
	}
}
