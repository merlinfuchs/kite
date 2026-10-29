package plan

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
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
	// The role of each plan by its product. They're the roles users can lose.
	roleByProduct := map[string]discord.RoleID{}
	for _, plan := range m.plans {
		if roleID, ok := planRoleID(&plan); ok && plan.LemonSqueezyProductID != "" {
			roleByProduct[plan.LemonSqueezyProductID] = roleID
		}
	}
	if len(roleByProduct) == 0 {
		return nil
	}

	subscriptions, err := m.subscriptionStore.AllSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all subscriptions: %w", err)
	}

	// The roles each user with a subscription should have, by user ID.
	want := map[string]map[discord.RoleID]bool{}
	for _, sub := range subscriptions {
		if want[sub.UserID] == nil {
			want[sub.UserID] = map[discord.RoleID]bool{}
		}
		roleID, ok := roleByProduct[sub.LemonsqueezyProductID.String]
		if ok && subscriptionKeepsRole(sub, now) {
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
		r.sync(user.DiscordID, maps.Values(roleByProduct), roles, "Kite subscription")
	}

	return nil
}

// subscriptionKeepsRole reports whether the subscription still gives its plan's
// role, which is for planRoleGrace after the paid period ended.
func subscriptionKeepsRole(sub *model.Subscription, now time.Time) bool {
	paidUntil := sub.PaidUntil()
	return paidUntil.Valid && now.Before(paidUntil.Time.Add(planRoleGrace))
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
// the other managed roles they have, unless they aren't in the guild.
func (r roleGranter) sync(discordUserID string, managed iter.Seq[discord.RoleID], want map[discord.RoleID]bool, reason string) {
	userID, ok := parseUserID(discordUserID)
	if !ok {
		return
	}

	member, err := r.client.Member(r.guildID, userID)
	if err != nil {
		if !util.IsDiscordRestErrorCode(err, unknownMember) {
			slog.Error(
				"Failed to get member",
				slog.String("error", err.Error()),
				slog.String("discord_user_id", discordUserID),
			)
		}
		return
	}

	for roleID := range managed {
		has := slices.Contains(member.RoleIDs, roleID)
		switch {
		case want[roleID] && !has:
			r.grantID(userID, roleID, reason)
		case !want[roleID] && has:
			err := r.client.RemoveRole(r.guildID, userID, roleID, api.AuditLogReason(reason+" ended"))
			logRoleError("Failed to remove role", err, userID, roleID)
		}
	}
}

// grant gives the user the role, unless they aren't in the guild.
func (r roleGranter) grant(discordUserID string, roleID discord.RoleID, reason string) {
	if userID, ok := parseUserID(discordUserID); ok {
		r.grantID(userID, roleID, reason)
	}
}

func (r roleGranter) grantID(userID discord.UserID, roleID discord.RoleID, reason string) {
	err := r.client.AddRole(r.guildID, userID, roleID, api.AddRoleData{
		AuditLogReason: api.AuditLogReason(reason),
	})
	logRoleError("Failed to add role", err, userID, roleID)
}

// unknownMember is the error of users who aren't in the guild.
const unknownMember = 10007

// logRoleError logs errors other than the user not being in the guild, like an
// unknown role.
func logRoleError(msg string, err error, userID discord.UserID, roleID discord.RoleID) {
	if err != nil && !util.IsDiscordRestErrorCode(err, unknownMember) {
		slog.Error(
			msg,
			slog.String("error", err.Error()),
			slog.String("discord_user_id", userID.String()),
			slog.String("discord_role_id", roleID.String()),
		)
	}
}

func parseUserID(discordUserID string) (discord.UserID, bool) {
	userID, err := discord.ParseSnowflake(discordUserID)
	if err != nil {
		slog.Error(
			"Failed to parse discord user ID",
			slog.String("error", err.Error()),
			slog.String("discord_user_id", discordUserID),
		)
		return 0, false
	}
	return discord.UserID(userID), true
}
