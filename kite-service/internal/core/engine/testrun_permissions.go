package engine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// The checks of test runs warn about changes Discord would reject because of
// the bot's permissions or role position. They read the bot's roles and the
// channels once per test, so a change the bot just made in Discord can take a
// moment to show.

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
	// Why the server couldn't be loaded, then the rest is empty.
	problem string
}

// testGuild loads a server for the checks, once per test.
func (s *testSession) testGuild(guildID discord.GuildID) *testGuild {
	s.mu.Lock()
	g, ok := s.guilds[guildID]
	s.mu.Unlock()
	if ok {
		return g
	}

	g = &testGuild{}
	if guild, err := s.session.Guild(guildID); err != nil {
		g.problem = fmt.Sprintf("The bot isn't in server %s, or it can't be loaded.", guildID)
	} else if roles, err := s.session.Roles(guildID); err != nil {
		g.problem = fmt.Sprintf("Couldn't load the roles of server %s to check permissions.", guildID)
	} else if bot, err := s.member(guildID, s.botID); err != nil {
		g.problem = fmt.Sprintf("Couldn't load the bot's member in server %s to check permissions.", guildID)
	} else {
		g.guild, g.roles, g.bot = *guild, roles, *bot
	}

	s.mu.Lock()
	s.guilds[guildID] = g
	s.mu.Unlock()
	return g
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

func (g *testGuild) checkPermissions(need discord.Permissions) []string {
	if g.problem != "" {
		return []string{g.problem}
	}
	if missing := missingPermissions(g.permissions(discord.Channel{}), need); len(missing) > 0 {
		return []string{fmt.Sprintf("The bot lacks %s in the server.", strings.Join(missing, ", "))}
	}
	return nil
}

func (s *testSession) checkGuild(guildID discord.GuildID, need discord.Permissions) []string {
	return s.testGuild(guildID).checkPermissions(need)
}

// checkChannel checks that the bot has the permissions in a channel. In
// threads, Send Messages is Send Messages in Threads, and the overwrites are
// the ones of the thread's channel.
func (s *testSession) checkChannel(channelID discord.ChannelID, need discord.Permissions) []string {
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
	if eval.IsThread(channel.Type) {
		if need.Has(discord.PermissionSendMessages) {
			need = need&^discord.PermissionSendMessages | discord.PermissionSendMessagesInThreads
		}
		if parent, err := s.session.Channel(channel.ParentID); err == nil {
			overwrites = *parent
		}
	}

	g := s.testGuild(channel.GuildID)
	if g.problem != "" {
		return []string{g.problem}
	}
	if missing := missingPermissions(g.permissions(overwrites), need); len(missing) > 0 {
		return []string{fmt.Sprintf("The bot lacks %s in channel %s (%s).", strings.Join(missing, ", "), channel.Name, channelID)}
	}
	return nil
}

func (s *testSession) checkChannelManage(channelID discord.ChannelID) []string {
	if channel, err := s.session.Channel(channelID); err == nil && eval.IsThread(channel.Type) {
		return s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionManageThreads)
	}
	return s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionManageChannels)
}

// checkOwnMessage warns about edits of messages the bot didn't send, which
// Discord doesn't allow.
func (s *testSession) checkOwnMessage(channelID discord.ChannelID, messageID discord.MessageID) []string {
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
func (s *testSession) checkMessageDelete(channelID discord.ChannelID, messageID discord.MessageID) []string {
	if _, ok := s.fakeMessage(messageID); ok {
		return nil
	}
	if msg, err := s.session.Message(channelID, messageID); err == nil && msg.Author.ID == s.botID {
		return nil
	}
	return s.checkChannel(channelID, discord.PermissionViewChannel|discord.PermissionManageMessages)
}

// checkRole checks that the bot can assign or change a role, which has to be
// below its own highest role.
func (s *testSession) checkRole(guildID discord.GuildID, roleID discord.RoleID) []string {
	g := s.testGuild(guildID)
	if g.problem != "" {
		return []string{g.problem}
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
func (s *testSession) checkModerate(guildID discord.GuildID, userID discord.UserID, need discord.Permissions, mustBeMember bool) []string {
	g := s.testGuild(guildID)
	warnings := g.checkPermissions(need)
	if g.problem != "" {
		return warnings
	}

	member, err := s.member(guildID, userID)
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

func (s *testSession) checkMemberEdit(guildID discord.GuildID, userID discord.UserID, data api.ModifyMemberData) []string {
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
		if member, err := s.member(guildID, userID); err == nil {
			current = member.RoleIDs
		}
		for _, id := range *data.Roles {
			if !slices.Contains(current, id) {
				warnings = append(warnings, s.checkRole(guildID, id)...)
			}
		}
		for _, id := range current {
			if !slices.Contains(*data.Roles, id) {
				warnings = append(warnings, s.checkRole(guildID, id)...)
			}
		}
	}

	if moderates {
		return append(s.checkModerate(guildID, userID, need, true), warnings...)
	}
	return append(s.checkGuild(guildID, need), warnings...)
}

type testAPICheck func(s *testSession, ids []discord.Snowflake, body []byte) []string

func inGuild(need discord.Permissions) testAPICheck {
	return func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return s.checkGuild(discord.GuildID(ids[0]), need)
	}
}

func inChannel(need discord.Permissions) testAPICheck {
	return func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannel(discord.ChannelID(ids[0]), discord.PermissionViewChannel|need)
	}
}

// testAPIRules are the permissions of common endpoints of the Discord API,
// which request blocks like Ban Member use. The first ID in the path is the
// server or channel. Other endpoints aren't checked.
var testAPIRules = []struct {
	methods []string
	path    *regexp.Regexp
	check   testAPICheck
}{
	{[]string{"PUT"}, regexp.MustCompile(`^/guilds/(\d+)/bans/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return s.checkModerate(discord.GuildID(ids[0]), discord.UserID(ids[1]), discord.PermissionBanMembers, false)
	}},
	{[]string{"DELETE"}, regexp.MustCompile(`^/guilds/(\d+)/bans/(\d+)$`), inGuild(discord.PermissionBanMembers)},
	{[]string{"DELETE"}, regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return s.checkModerate(discord.GuildID(ids[0]), discord.UserID(ids[1]), discord.PermissionKickMembers, true)
	}},
	{[]string{"PATCH"}, regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		var data api.ModifyMemberData
		if err := json.Unmarshal(body, &data); err != nil {
			return nil
		}
		return s.checkMemberEdit(discord.GuildID(ids[0]), discord.UserID(ids[1]), data)
	}},
	{[]string{"PUT", "DELETE"}, regexp.MustCompile(`^/guilds/(\d+)/members/(\d+)/roles/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return append(s.checkGuild(discord.GuildID(ids[0]), discord.PermissionManageRoles), s.checkRole(discord.GuildID(ids[0]), discord.RoleID(ids[2]))...)
	}},
	{[]string{"POST"}, regexp.MustCompile(`^/guilds/(\d+)/roles$`), inGuild(discord.PermissionManageRoles)},
	{[]string{"PATCH", "DELETE"}, regexp.MustCompile(`^/guilds/(\d+)/roles/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return append(s.checkGuild(discord.GuildID(ids[0]), discord.PermissionManageRoles), s.checkRole(discord.GuildID(ids[0]), discord.RoleID(ids[1]))...)
	}},
	{[]string{"POST"}, regexp.MustCompile(`^/guilds/(\d+)/channels$`), inGuild(discord.PermissionManageChannels)},
	{[]string{"PATCH"}, regexp.MustCompile(`^/guilds/(\d+)$`), inGuild(discord.PermissionManageGuild)},
	{[]string{"POST"}, regexp.MustCompile(`^/channels/(\d+)/messages$`), inChannel(discord.PermissionSendMessages)},
	{[]string{"PATCH"}, regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return s.checkOwnMessage(discord.ChannelID(ids[0]), discord.MessageID(ids[1]))
	}},
	{[]string{"DELETE"}, regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return s.checkMessageDelete(discord.ChannelID(ids[0]), discord.MessageID(ids[1]))
	}},
	{[]string{"POST"}, regexp.MustCompile(`^/channels/(\d+)/messages/bulk-delete$`), inChannel(discord.PermissionManageMessages | discord.PermissionReadMessageHistory)},
	{[]string{"PUT"}, regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)/reactions/[^/]+/@me$`), inChannel(discord.PermissionReadMessageHistory | discord.PermissionAddReactions)},
	{[]string{"DELETE"}, regexp.MustCompile(`^/channels/(\d+)/messages/(\d+)/reactions(/[^/]+/(\d+))?$`), inChannel(discord.PermissionManageMessages)},
	{[]string{"PATCH", "DELETE"}, regexp.MustCompile(`^/channels/(\d+)$`), func(s *testSession, ids []discord.Snowflake, body []byte) []string {
		return s.checkChannelManage(discord.ChannelID(ids[0]))
	}},
	{[]string{"PUT", "DELETE"}, regexp.MustCompile(`^/channels/(\d+)/permissions/(\d+)$`), inChannel(discord.PermissionManageRoles)},
	{[]string{"POST"}, regexp.MustCompile(`^/channels/(\d+)/invites$`), inChannel(discord.PermissionCreateInstantInvite)},
	{[]string{"POST"}, regexp.MustCompile(`^/channels/(\d+)/(messages/(\d+)/)?threads$`), inChannel(discord.PermissionCreatePublicThreads)},
	{[]string{"POST"}, regexp.MustCompile(`^/channels/(\d+)/webhooks$`), inChannel(discord.PermissionManageWebhooks)},
}

func (s *testSession) checkAPIRequest(req provider.DiscordAPIRequest) []string {
	path, _, _ := strings.Cut(req.Path, "?")
	for _, rule := range testAPIRules {
		if !slices.Contains(rule.methods, req.Method) {
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
		return rule.check(s, ids, req.Body)
	}
	return nil
}
