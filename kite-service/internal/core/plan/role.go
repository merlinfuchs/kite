package plan

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
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

func (m *PlanManager) handOutRoles(ctx context.Context) error {
	if m.config.DiscordBotToken == "" || m.config.DiscordGuildID == "" {
		return nil
	}

	client := api.NewClient("Bot " + m.config.DiscordBotToken)

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

		discordGuildID, err := strconv.ParseUint(m.config.DiscordGuildID, 10, 64)
		if err != nil {
			slog.Error(
				"Failed to parse discord guild ID from config",
				slog.String("error", err.Error()),
				slog.String("discord_guild_id", m.config.DiscordGuildID),
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

		discordUserID, err := strconv.ParseUint(user.DiscordID, 10, 64)
		if err != nil {
			slog.Error(
				"Failed to parse discord user ID from subscription",
				slog.String("error", err.Error()),
				slog.String("discord_user_id", sub.UserID),
				slog.String("subscription_id", sub.ID),
			)
			continue
		}

		discordRoleID, err := strconv.ParseUint(plan.DiscordRoleID, 10, 64)
		if err != nil {
			slog.Error(
				"Failed to parse discord role ID from plan",
				slog.String("error", err.Error()),
				slog.String("discord_role_id", plan.DiscordRoleID),
				slog.String("subscription_id", sub.ID),
			)
			continue
		}

		err = client.AddRole(
			discord.GuildID(discordGuildID),
			discord.UserID(discordUserID),
			discord.RoleID(discordRoleID),
			api.AddRoleData{
				AuditLogReason: api.AuditLogReason("Kite subscription"),
			},
		)
		if err != nil {
			// TODO: ignore member not found error
			slog.Error(
				"Failed to add role",
				slog.String("error", err.Error()),
				slog.String("subscription_id", sub.ID),
			)
			continue
		}
	}

	return nil
}
