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

// creatorMinInstallers is how many other users must have installed a template
// before its creator gets the creator role, so listing one isn't enough.
const creatorMinInstallers = 10

// handOutRoles gives users the role of their plan, and creators of templates
// in the marketplace that others installed the creator role, in the Kite
// server. Users who aren't in it get theirs once they join. Roles are never
// taken away.
func (m *PlanManager) handOutRoles(ctx context.Context) error {
	if m.config.DiscordBotToken == "" || m.config.DiscordGuildID == "" {
		return nil
	}

	guildID, err := discord.ParseSnowflake(m.config.DiscordGuildID)
	if err != nil {
		return fmt.Errorf("failed to parse discord guild ID: %w", err)
	}
	r := roleGranter{
		client:  api.NewClient("Bot " + m.config.DiscordBotToken).WithContext(ctx),
		guildID: discord.GuildID(guildID),
	}

	// One failing doesn't keep the other from running.
	return errors.Join(m.handOutPlanRoles(ctx, r), m.handOutCreatorRoles(ctx, r))
}

func (m *PlanManager) handOutPlanRoles(ctx context.Context, r roleGranter) error {
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

		roleID, err := discord.ParseSnowflake(plan.DiscordRoleID)
		if err != nil {
			slog.Error(
				"Failed to parse discord role ID from plan",
				slog.String("error", err.Error()),
				slog.String("discord_role_id", plan.DiscordRoleID),
				slog.String("subscription_id", sub.ID),
			)
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

		r.grant(user.DiscordID, discord.RoleID(roleID), "Kite subscription")
	}

	return nil
}

func (m *PlanManager) handOutCreatorRoles(ctx context.Context, r roleGranter) error {
	if m.config.DiscordCreatorRoleID == "" {
		return nil
	}

	roleID, err := discord.ParseSnowflake(m.config.DiscordCreatorRoleID)
	if err != nil {
		return fmt.Errorf("failed to parse discord creator role ID: %w", err)
	}

	creators, err := m.bundleStore.CreatorDiscordIDs(ctx, creatorMinInstallers)
	if err != nil {
		return fmt.Errorf("failed to get creators: %w", err)
	}
	for _, discordID := range creators {
		r.grant(discordID, discord.RoleID(roleID), "Kite template creator")
	}

	return nil
}

type roleGranter struct {
	client  *api.Client
	guildID discord.GuildID
}

// grant gives the user the role, unless they aren't in the guild.
func (r roleGranter) grant(discordUserID string, roleID discord.RoleID, reason string) {
	userID, err := discord.ParseSnowflake(discordUserID)
	if err != nil {
		slog.Error(
			"Failed to parse discord user ID",
			slog.String("error", err.Error()),
			slog.String("discord_user_id", discordUserID),
		)
		return
	}

	err = r.client.AddRole(r.guildID, discord.UserID(userID), roleID, api.AddRoleData{
		AuditLogReason: api.AuditLogReason(reason),
	})
	var dErr *httputil.HTTPError
	// 10007 is Unknown Member. Other errors, like an unknown role, are logged.
	if err != nil && (!errors.As(err, &dErr) || dErr.Code != 10007) {
		slog.Error(
			"Failed to add role",
			slog.String("error", err.Error()),
			slog.String("discord_user_id", discordUserID),
			slog.String("discord_role_id", roleID.String()),
			slog.String("reason", reason),
		)
	}
}
