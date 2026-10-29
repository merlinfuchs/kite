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
	"github.com/kitecloud/kite/kite-service/internal/model"
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
// server. Users who aren't in it get theirs once they join. Plan roles are
// taken away again once the subscription ended, the creator role never is.
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
	return errors.Join(m.syncPlanRoles(ctx, r, time.Now()), m.handOutCreatorRoles(ctx, r))
}

// planRoleGrace is how long users keep the role of their plan after their
// subscription ended.
const planRoleGrace = 3 * 24 * time.Hour

// syncPlanRoles gives every user who ever subscribed the roles of the plans
// they pay for, and takes away the roles of plans they no longer pay for.
// Roles of users who never subscribed aren't touched.
func (m *PlanManager) syncPlanRoles(ctx context.Context, r roleGranter, now time.Time) error {
	subscriptions, err := m.subscriptionStore.AllSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all subscriptions: %w", err)
	}

	// The roles of all plans, which are the ones users can lose.
	planRoles := map[discord.RoleID]bool{}
	for _, plan := range m.plans {
		if roleID, ok := planRoleID(&plan); ok {
			planRoles[roleID] = true
		}
	}
	if len(planRoles) == 0 {
		return nil
	}

	// The roles each user with a subscription should have, by user ID.
	want := map[string]map[discord.RoleID]bool{}
	for _, sub := range subscriptions {
		if want[sub.UserID] == nil {
			want[sub.UserID] = map[discord.RoleID]bool{}
		}
		if !subscriptionKeepsRole(sub, now) || !sub.LemonsqueezyProductID.Valid {
			continue
		}
		plan := m.PlanByLemonSqueezyProductID(sub.LemonsqueezyProductID.String)
		if plan == nil {
			continue
		}
		if roleID, ok := planRoleID(plan); ok {
			want[sub.UserID][roleID] = true
		}
	}

	for userID, roles := range want {
		user, err := m.userStore.User(ctx, userID)
		if err != nil {
			slog.Error(
				"Failed to get user by ID",
				slog.String("error", err.Error()),
				slog.String("user_id", userID),
			)
			continue
		}
		r.sync(user.DiscordID, planRoles, roles, "Kite subscription")
	}

	return nil
}

// subscriptionKeepsRole reports whether the subscription still gives its plan's
// role: while it gives access, and for planRoleGrace after it ended.
func subscriptionKeepsRole(sub *model.Subscription, now time.Time) bool {
	if sub.IsActive() {
		return true
	}
	// Cancelled subscriptions end when the paid period does.
	return sub.EndsAt.Valid && now.Before(sub.EndsAt.Time.Add(planRoleGrace))
}

func planRoleID(plan *model.Plan) (discord.RoleID, bool) {
	if plan.DiscordRoleID == "" {
		return 0, false
	}
	roleID, err := discord.ParseSnowflake(plan.DiscordRoleID)
	if err != nil {
		slog.Error(
			"Failed to parse discord role ID from plan",
			slog.String("error", err.Error()),
			slog.String("discord_role_id", plan.DiscordRoleID),
			slog.String("plan_id", plan.ID),
		)
		return 0, false
	}
	return discord.RoleID(roleID), true
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

// sync gives the user the roles of want they don't have yet, and takes away
// the roles of managed they have but shouldn't, unless they aren't in the
// guild.
func (r roleGranter) sync(discordUserID string, managed, want map[discord.RoleID]bool, reason string) {
	userID, err := discord.ParseSnowflake(discordUserID)
	if err != nil {
		slog.Error(
			"Failed to parse discord user ID",
			slog.String("error", err.Error()),
			slog.String("discord_user_id", discordUserID),
		)
		return
	}

	member, err := r.client.Member(r.guildID, discord.UserID(userID))
	if err != nil {
		if !isUnknownMember(err) {
			slog.Error(
				"Failed to get member",
				slog.String("error", err.Error()),
				slog.String("discord_user_id", discordUserID),
			)
		}
		return
	}

	has := map[discord.RoleID]bool{}
	for _, roleID := range member.RoleIDs {
		has[roleID] = true
	}
	for roleID := range want {
		if !has[roleID] {
			r.grant(discordUserID, roleID, reason)
		}
	}
	for roleID := range managed {
		if has[roleID] && !want[roleID] {
			err := r.client.RemoveRole(r.guildID, discord.UserID(userID), roleID, api.AuditLogReason(reason+" ended"))
			if err != nil && !isUnknownMember(err) {
				slog.Error(
					"Failed to remove role",
					slog.String("error", err.Error()),
					slog.String("discord_user_id", discordUserID),
					slog.String("discord_role_id", roleID.String()),
				)
			}
		}
	}
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
	// Other errors, like an unknown role, are logged.
	if err != nil && !isUnknownMember(err) {
		slog.Error(
			"Failed to add role",
			slog.String("error", err.Error()),
			slog.String("discord_user_id", discordUserID),
			slog.String("discord_role_id", roleID.String()),
			slog.String("reason", reason),
		)
	}
}

// isUnknownMember reports whether the user isn't in the guild.
func isUnknownMember(err error) bool {
	var dErr *httputil.HTTPError
	return errors.As(err, &dErr) && dErr.Code == 10007
}
