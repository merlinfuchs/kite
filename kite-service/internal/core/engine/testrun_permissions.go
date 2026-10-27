package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// The checks of test runs warn about changes Discord would reject because of
// the bot's permissions or role position. They read the bot's roles and the
// channels from the cache, so a change the bot just made in Discord can take
// a moment to show.

var testPermissionNames = []struct {
	perm discord.Permissions
	name string
}{
	{discord.PermissionViewChannel, "View Channel"},
	{discord.PermissionSendMessages, "Send Messages"},
	{discord.PermissionSendMessagesInThreads, "Send Messages in Threads"},
	{discord.PermissionEmbedLinks, "Embed Links"},
	{discord.PermissionAttachFiles, "Attach Files"},
	{discord.PermissionReadMessageHistory, "Read Message History"},
	{discord.PermissionAddReactions, "Add Reactions"},
	{discord.PermissionManageMessages, "Manage Messages"},
	{discord.PermissionManageChannels, "Manage Channels"},
	{discord.PermissionManageThreads, "Manage Threads"},
	{discord.PermissionCreatePublicThreads, "Create Public Threads"},
	{discord.PermissionCreatePrivateThreads, "Create Private Threads"},
	{discord.PermissionManageRoles, "Manage Roles"},
	{discord.PermissionManageNicknames, "Manage Nicknames"},
	{discord.PermissionManageWebhooks, "Manage Webhooks"},
	{discord.PermissionManageGuild, "Manage Server"},
	{discord.PermissionKickMembers, "Kick Members"},
	{discord.PermissionBanMembers, "Ban Members"},
	{discord.PermissionModerateMembers, "Timeout Members"},
	{discord.PermissionMuteMembers, "Mute Members"},
	{discord.PermissionDeafenMembers, "Deafen Members"},
	{discord.PermissionMoveMembers, "Move Members"},
	{discord.PermissionConnect, "Connect"},
	{discord.PermissionCreateInstantInvite, "Create Invite"},
}

func missingPermissions(have discord.Permissions, need discord.Permissions) []string {
	var res []string
	for _, p := range testPermissionNames {
		if need.Has(p.perm) && !have.Has(p.perm) {
			res = append(res, p.name)
		}
	}
	return res
}

// testGuild is what the checks need to know about a server.
type testGuild struct {
	guild discord.Guild
	roles []discord.Role
	bot   discord.Member
}

func (s *testSession) testGuild(guildID discord.GuildID) (*testGuild, string) {
	guild, err := s.session.Guild(guildID)
	if err != nil {
		return nil, fmt.Sprintf("The bot isn't in server %s, or it can't be loaded.", guildID)
	}
	roles, err := s.session.Roles(guildID)
	if err != nil {
		return nil, fmt.Sprintf("Couldn't load the roles of server %s to check permissions.", guildID)
	}
	bot, err := s.session.Member(guildID, s.botID)
	if err != nil {
		return nil, fmt.Sprintf("Couldn't load the bot's member in server %s to check permissions.", guildID)
	}
	return &testGuild{guild: *guild, roles: roles, bot: *bot}, ""
}

func (g *testGuild) role(id discord.RoleID) (discord.Role, bool) {
	i := slices.IndexFunc(g.roles, func(r discord.Role) bool { return r.ID == id })
	if i == -1 {
		return discord.Role{}, false
	}
	return g.roles[i], true
}

// highestPosition is the position of the member's highest role.
func (g *testGuild) highestPosition(member discord.Member) int {
	highest := 0
	for _, id := range member.RoleIDs {
		if role, ok := g.role(id); ok && role.Position > highest {
			highest = role.Position
		}
	}
	return highest
}

func (g *testGuild) permissions(channel discord.Channel) discord.Permissions {
	return discord.CalcOverrides(g.guild, channel, g.bot, g.roles)
}

func (s *testSession) checkGuild(ctx context.Context, guildID discord.GuildID, need discord.Permissions) []string {
	g, problem := s.testGuild(guildID)
	if g == nil {
		return []string{problem}
	}
	if missing := missingPermissions(g.permissions(discord.Channel{}), need); len(missing) > 0 {
		return []string{fmt.Sprintf("The bot lacks %s in the server.", strings.Join(missing, ", "))}
	}
	return nil
}

// checkChannel checks that the bot has the permissions in a channel. In
// threads, Send Messages is Send Messages in Threads, and the overwrites are
// the ones of the thread's channel.
func (s *testSession) checkChannel(ctx context.Context, channelID discord.ChannelID, need discord.Permissions) []string {
	if _, ok := s.fakeChannel(channelID); ok {
		return nil
	}

	channel, err := s.session.Channel(channelID)
	if err != nil {
		return []string{fmt.Sprintf("Channel %s doesn't exist or the bot can't see it.", channelID)}
	}
	if !channel.GuildID.IsValid() {
		return nil
	}

	overwrites := *channel
	if isThread(channel.Type) {
		if need.Has(discord.PermissionSendMessages) {
			need = need&^discord.PermissionSendMessages | discord.PermissionSendMessagesInThreads
		}
		if parent, err := s.session.Channel(channel.ParentID); err == nil {
			overwrites = *parent
		}
	}

	g, problem := s.testGuild(channel.GuildID)
	if g == nil {
		return []string{problem}
	}
	if missing := missingPermissions(g.permissions(overwrites), need); len(missing) > 0 {
		return []string{fmt.Sprintf("The bot lacks %s in channel %s (%s).", strings.Join(missing, ", "), channel.Name, channelID)}
	}
	return nil
}

func isThread(t discord.ChannelType) bool {
	return t == discord.GuildPublicThread || t == discord.GuildPrivateThread || t == discord.GuildAnnouncementThread
}

func (s *testSession) checkChannelManage(ctx context.Context, channelID discord.ChannelID) []string {
	if channel, err := s.session.Channel(channelID); err == nil && isThread(channel.Type) {
		return s.checkChannel(ctx, channelID, discord.PermissionViewChannel|discord.PermissionManageThreads)
	}
	return s.checkChannel(ctx, channelID, discord.PermissionViewChannel|discord.PermissionManageChannels)
}

// checkOwnMessage warns about edits of messages the bot didn't send, which
// Discord doesn't allow.
func (s *testSession) checkOwnMessage(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID) []string {
	if _, ok := s.fakeMessage(messageID); ok {
		return nil
	}
	msg, err := s.session.Message(channelID, messageID)
	if err != nil {
		return []string{fmt.Sprintf("Message %s doesn't exist in channel %s or the bot can't see it.", messageID, channelID)}
	}
	if msg.Author.ID != s.botID {
		return []string{fmt.Sprintf("Message %s was sent by someone else, the bot can only edit its own messages.", messageID)}
	}
	return nil
}

// checkMessageDelete checks that the bot can delete a message, which needs
// Manage Messages unless the bot sent it.
func (s *testSession) checkMessageDelete(ctx context.Context, channelID discord.ChannelID, messageID discord.MessageID) []string {
	if _, ok := s.fakeMessage(messageID); ok {
		return nil
	}
	if msg, err := s.session.Message(channelID, messageID); err == nil && msg.Author.ID == s.botID {
		return nil
	}
	return s.checkChannel(ctx, channelID, discord.PermissionViewChannel|discord.PermissionManageMessages)
}

// checkRole checks that the bot can assign or change a role, which has to be
// below its own highest role.
func (s *testSession) checkRole(ctx context.Context, guildID discord.GuildID, roleID discord.RoleID) []string {
	g, problem := s.testGuild(guildID)
	if g == nil {
		return []string{problem}
	}

	role, ok := g.role(roleID)
	switch {
	case !ok:
		return []string{fmt.Sprintf("Role %s doesn't exist in the server.", roleID)}
	case discord.GuildID(roleID) == guildID:
		return []string{"The @everyone role can't be assigned."}
	case role.Managed:
		return []string{fmt.Sprintf("Role %s is managed by an integration, like a bot's own role, and can't be assigned.", role.Name)}
	case g.guild.OwnerID != s.botID && role.Position >= g.highestPosition(g.bot):
		return []string{fmt.Sprintf("Role %s isn't below the bot's highest role, so the bot can't manage it. Move the bot's role above it in the server settings.", role.Name)}
	}
	return nil
}

// checkModerate checks that the bot can act on a member, like kicking them,
// which needs the member's highest role to be below the bot's.
func (s *testSession) checkModerate(ctx context.Context, guildID discord.GuildID, userID discord.UserID, need discord.Permissions, mustBeMember bool) []string {
	warnings := s.checkGuild(ctx, guildID, need)

	g, _ := s.testGuild(guildID)
	if g == nil {
		return warnings
	}

	member, err := s.session.Member(guildID, userID)
	if err != nil {
		if mustBeMember {
			warnings = append(warnings, fmt.Sprintf("User %s isn't a member of the server.", userID))
		}
		return warnings
	}

	switch {
	case userID == s.botID:
		warnings = append(warnings, "The bot can't do this to itself.")
	case userID == g.guild.OwnerID:
		warnings = append(warnings, "The member is the owner of the server, the bot can't act on them.")
	case g.guild.OwnerID != s.botID && g.highestPosition(*member) >= g.highestPosition(g.bot):
		warnings = append(warnings, "The member's highest role isn't below the bot's highest role, so the bot can't act on them.")
	}
	return warnings
}

func (s *testSession) checkMemberEdit(ctx context.Context, guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) []string {
	var warnings []string
	var need discord.Permissions
	moderates := false

	if data.Nick != nil {
		need |= discord.PermissionManageNicknames
		moderates = true
	}
	if data.Mute != nil {
		need |= discord.PermissionMuteMembers
	}
	if data.Deaf != nil {
		need |= discord.PermissionDeafenMembers
	}
	if data.VoiceChannel.IsValid() {
		need |= discord.PermissionMoveMembers
	}
	if data.CommunicationDisabledUntil != nil {
		need |= discord.PermissionModerateMembers
		moderates = true
	}

	if data.Roles != nil {
		need |= discord.PermissionManageRoles

		// Only the roles that change have to be below the bot's.
		var current []discord.RoleID
		if member, err := s.session.Member(guildID, userID); err == nil {
			current = member.RoleIDs
		}
		for _, id := range *data.Roles {
			if !slices.Contains(current, id) {
				warnings = append(warnings, s.checkRole(ctx, guildID, id)...)
			}
		}
		for _, id := range current {
			if !slices.Contains(*data.Roles, id) {
				warnings = append(warnings, s.checkRole(ctx, guildID, id)...)
			}
		}
	}

	if moderates {
		return append(s.checkModerate(ctx, guildID, userID, need, true), warnings...)
	}
	return append(s.checkGuild(ctx, guildID, need), warnings...)
}

// testAPIRules are the permissions of common endpoints of the Discord API,
// which request blocks like Ban Member use. Other endpoints aren't checked.
var testAPIRules = []struct {
	methods string
	path    *regexp.Regexp
	check   func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string
}{
	{"PUT", regexp.MustCompile(`^/guilds/(\d+)/bans/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkModerate(ctx, discord.GuildID(ids[0]), discord.UserID(ids[1]), discord.PermissionBanMembers, false)
	}},
	{"DELETE", regexp.MustCompile(`^/guilds/(\d+)/bans/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkGuild(ctx, discord.GuildID(ids[0]), discord.PermissionBanMembers)
	}},
	{"DELETE", regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkModerate(ctx, discord.GuildID(ids[0]), discord.UserID(ids[1]), discord.PermissionKickMembers, true)
	}},
	{"PATCH", regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		var data api.ModifyMemberData
		if err := json.Unmarshal(body, &data); err != nil {
			return nil
		}
		return s.checkMemberEdit(ctx, discord.GuildID(ids[0]), discord.UserID(ids[1]), data)
	}},
	{"PUT DELETE", regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)/roles/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return append(s.checkGuild(ctx, discord.GuildID(ids[0]), discord.PermissionManageRoles), s.checkRole(ctx, discord.GuildID(ids[0]), discord.RoleID(ids[2]))...)
	}},
	{"POST", regexp.MustCompile(`^/guilds/(\d+)/roles$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkGuild(ctx, discord.GuildID(ids[0]), discord.PermissionManageRoles)
	}},
	{"PATCH DELETE", regexp.MustCompile(`^/guilds/(\d+)/roles/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return append(s.checkGuild(ctx, discord.GuildID(ids[0]), discord.PermissionManageRoles), s.checkRole(ctx, discord.GuildID(ids[0]), discord.RoleID(ids[1]))...)
	}},
	{"POST", regexp.MustCompile(`^/guilds/(\d+)/channels$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkGuild(ctx, discord.GuildID(ids[0]), discord.PermissionManageChannels)
	}},
	{"PATCH", regexp.MustCompile(`^/guilds/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkGuild(ctx, discord.GuildID(ids[0]), discord.PermissionManageGuild)
	}},
	{"POST", regexp.MustCompile(`^/channels/(\d+)/messages$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionSendMessages)
	}},
	{"PATCH", regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkOwnMessage(ctx, discord.ChannelID(ids[0]), discord.MessageID(ids[1]))
	}},
	{"DELETE", regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkMessageDelete(ctx, discord.ChannelID(ids[0]), discord.MessageID(ids[1]))
	}},
	{"POST", regexp.MustCompile(`^/channels/(\d+)/messages/bulk-delete$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionManageMessages|discord.PermissionReadMessageHistory)
	}},
	{"PUT", regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)/reactions/[^/]+/@me$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionReadMessageHistory|discord.PermissionAddReactions)
	}},
	{"DELETE", regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)/reactions(/[^/]+/(\d+))?$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionManageMessages)
	}},
	{"PATCH DELETE", regexp.MustCompile(`^/channels/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannelManage(ctx, discord.ChannelID(ids[0]))
	}},
	{"PUT DELETE", regexp.MustCompile(`^/channels/(\d+)/permissions/(\d+)$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionManageRoles)
	}},
	{"POST", regexp.MustCompile(`^/channels/(\d+)/invites$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionCreateInstantInvite)
	}},
	{"POST", regexp.MustCompile(`^/channels/(\d+)/(messages/(\d+)/)?threads$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionCreatePublicThreads)
	}},
	{"POST", regexp.MustCompile(`^/channels/(\d+)/webhooks$`), func(s *testSession, ctx context.Context, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(ctx, discord.ChannelID(ids[0]), discord.PermissionViewChannel|discord.PermissionManageWebhooks)
	}},
}

func (s *testSession) checkAPIRequest(ctx context.Context, req provider.DiscordAPIRequest) []string {
	path, _, _ := strings.Cut(req.Path, "?")
	for _, rule := range testAPIRules {
		if !slices.Contains(strings.Fields(rule.methods), req.Method) {
			continue
		}
		match := rule.path.FindStringSubmatch(path)
		if match == nil {
			continue
		}

		var ids []discord.Snowflake
		for _, m := range match[1:] {
			if id, err := discord.ParseSnowflake(m); err == nil {
				ids = append(ids, id)
			}
		}
		return rule.check(s, ctx, ids, req.Body)
	}
	return nil
}
