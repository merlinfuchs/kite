package eval

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/expr-lang/expr/ast"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type Context struct {
	Env      Env
	Patchers []ast.Visitor
}

type Env map[string]any

type InteractionEnv struct {
	interaction *discord.InteractionEvent

	ID         string                   `expr:"id" json:"id"`
	Channel    *CurrentChannelEnv       `expr:"channel" json:"channel"`
	Guild      *CurrentGuildEnv         `expr:"guild" json:"guild"`
	User       any                      `expr:"user" json:"user"`
	Member     any                      `expr:"member" json:"member"`
	Command    *CommandEnv              `expr:"command" json:"command"`
	Components map[string]*ComponentEnv `expr:"components" json:"components"`
	// Values are the options picked in a select menu, Value is the first of them.
	Values []string `expr:"values" json:"values"`
	Value  string   `expr:"value" json:"value"`
}

func NewInteractionEnv(i *discord.InteractionEvent) *InteractionEnv {
	return newInteractionEnv(i, nil)
}

// newInteractionEnv fills in what the interaction doesn't carry itself, like
// the name of the server, from the cache of session.
func newInteractionEnv(i *discord.InteractionEvent, session *state.State) *InteractionEnv {
	guild := guildLookup{session: session, guildID: i.GuildID}

	e := &InteractionEnv{
		interaction: i,

		ID:         i.ID.String(),
		Channel:    currentChannelEnv(session, i.ChannelID, i.Channel),
		Components: NewComponentsEnv(i),
	}

	if i.Member != nil {
		e.Member = guild.memberEnv(*i.Member)
		e.User = e.Member
	} else {
		e.User = NewUserEnv(*i.User)
		e.Member = e.User
	}

	if i.GuildID != 0 {
		e.Guild = guild.guildEnv()
	}

	if i.Data.InteractionType() == discord.CommandInteractionType {
		e.Command = newCommandEnv(i, guild)
	}

	if data, ok := i.Data.(*discord.StringSelectInteraction); ok {
		e.Values = data.Values
		if len(data.Values) > 0 {
			e.Value = data.Values[0]
		}
	}

	return e
}

func NewContextFromInteraction(i *discord.InteractionEvent, session *state.State) Context {
	interactionEnv := newInteractionEnv(i, session)

	return Context{
		Env: Env{
			"interaction": interactionEnv,
			"channel":     interactionEnv.Channel,
			"guild":       interactionEnv.Guild,
			"server":      interactionEnv.Guild,
			"user":        interactionEnv.User,
			"member":      interactionEnv.Member,
			"app":         NewAppEnv(session),

			"arg": func(name string) any {
				if interactionEnv.Command == nil {
					return nil
				}

				return interactionEnv.Command.Args[name]
			},
			"input": func(customID string) any {
				if interactionEnv.Components == nil {
					return nil
				}

				if component, ok := interactionEnv.Components[customID]; ok {
					return component.Value
				}
				return nil
			},
		},
	}
}

func (e InteractionEnv) String() string {
	return e.interaction.ID.String()
}

type CommandEnv struct {
	interaction *discord.InteractionEvent
	cmd         *discord.CommandInteraction

	ID   string         `expr:"id" json:"id"`
	Args map[string]any `expr:"args" json:"args"`
}

func NewCommandEnv(i *discord.InteractionEvent) *CommandEnv {
	return newCommandEnv(i, guildLookup{guildID: i.GuildID})
}

func newCommandEnv(i *discord.InteractionEvent, guild guildLookup) *CommandEnv {
	data, _ := i.Data.(*discord.CommandInteraction)

	args := make(map[string]any)

	var addArg func(option discord.CommandInteractionOption)
	addArg = func(option discord.CommandInteractionOption) {
		var value any
		_ = json.Unmarshal(option.Value, &value)

		switch option.Type {
		case discord.UserOptionType:
			userID, _ := strconv.ParseInt(value.(string), 10, 64)
			user := data.Resolved.Users[discord.UserID(userID)]

			if member, ok := data.Resolved.Members[discord.UserID(userID)]; ok {
				member.User = user
				args[option.Name] = guild.memberEnv(member)
			} else {
				args[option.Name] = NewUserEnv(user)
			}
		case discord.RoleOptionType:
			roleID, _ := strconv.ParseInt(value.(string), 10, 64)
			role := data.Resolved.Roles[discord.RoleID(roleID)]
			args[option.Name] = NewRoleEnv(role)
		case discord.ChannelOptionType:
			channelID, _ := strconv.ParseInt(value.(string), 10, 64)
			channel := data.Resolved.Channels[discord.ChannelID(channelID)]
			args[option.Name] = channelEnv(guild.session, channel.ID, &channel)
		case discord.MentionableOptionType:
			mentionableID, _ := strconv.ParseInt(value.(string), 10, 64)
			user, ok := data.Resolved.Users[discord.UserID(mentionableID)]
			if ok {
				args[option.Name] = NewUserEnv(user)
			} else {
				role := data.Resolved.Roles[discord.RoleID(mentionableID)]
				args[option.Name] = NewRoleEnv(role)
			}
		case discord.AttachmentOptionType:
			attachmentID, _ := strconv.ParseInt(value.(string), 10, 64)
			attachment := data.Resolved.Attachments[discord.AttachmentID(attachmentID)]
			args[option.Name] = NewAttachmentEnv(&attachment)
		case discord.SubcommandGroupOptionType:
			for _, subcommand := range option.Options {
				addArg(subcommand)
			}
		case discord.SubcommandOptionType:
			for _, option := range option.Options {
				addArg(option)
			}
		default:
			args[option.Name] = value
		}
	}

	for _, option := range data.Options {
		addArg(option)
	}

	return &CommandEnv{
		interaction: i,
		cmd:         data,

		ID:   data.ID.String(),
		Args: args,
	}
}

func (c CommandEnv) String() string {
	return c.ID
}

type ComponentEnv struct {
	CustomID string `expr:"custom_id" json:"custom_id"`
	Value    string `expr:"value" json:"value"`
}

func NewComponentsEnv(i *discord.InteractionEvent) map[string]*ComponentEnv {
	components := make(map[string]*ComponentEnv)

	data, ok := i.Data.(*discord.ModalInteraction)
	if !ok {
		return components
	}

	for _, row := range data.Components {
		actionRow, ok := row.(*discord.ActionRowComponent)
		if !ok {
			continue
		}

		for _, component := range *actionRow {
			c := NewComponentEnv(component)
			if c != nil {
				components[c.CustomID] = c
			}
		}
	}

	return components
}

func NewComponentEnv(component discord.InteractiveComponent) *ComponentEnv {
	switch c := component.(type) {
	case *discord.TextInputComponent:
		return &ComponentEnv{
			CustomID: string(c.CustomID),
			Value:    c.Value,
		}
	}

	return nil
}

func (c ComponentEnv) String() string {
	return c.Value
}

type EventEnv struct {
	event ws.Event

	User    any                `expr:"user" json:"user"`
	Member  any                `expr:"member" json:"member"`
	Channel *CurrentChannelEnv `expr:"channel" json:"channel"`
	Message *MessageEnv        `expr:"message" json:"message"`
	Guild   any                `expr:"guild" json:"guild"`
	Emoji   *EmojiEnv          `expr:"emoji" json:"emoji"`

	Schedule *ScheduleEnv `expr:"schedule" json:"schedule"`
}

type ScheduleEnv struct {
	Time string `expr:"time" json:"time"`
	Unix int64  `expr:"unix" json:"unix"`
}

func NewScheduleEnv(e *schedule.Event) *ScheduleEnv {
	t := e.Time.UTC()
	return &ScheduleEnv{
		Time: t.Format(time.RFC3339),
		Unix: t.Unix(),
	}
}

func (s ScheduleEnv) String() string {
	return s.Time
}

func NewEventEnv(event ws.Event) *EventEnv {
	return newEventEnv(event, nil)
}

// newEventEnv fills in what the event doesn't carry itself, like the name of
// the server, from the cache of session.
func newEventEnv(event ws.Event, session *state.State) *EventEnv {
	env := &EventEnv{
		event: event,
	}

	guild := func(id discord.GuildID) guildLookup {
		return guildLookup{session: session, guildID: id}
	}

	switch e := event.(type) {
	case *gateway.MessageCreateEvent:
		if e.Member != nil {
			env.Member = guild(e.GuildID).memberEnv(*e.Member)
			env.User = env.Member
		} else {
			env.User = NewUserEnv(e.Author)
			env.Member = env.User
		}
		env.Channel = currentChannelEnv(session, e.ChannelID, nil)
		if e.GuildID != 0 {
			env.Guild = guild(e.GuildID).guildEnv()
		}
		env.Message = NewMessageEnv(e.Message)
	case *gateway.MessageUpdateEvent:
		if e.Member != nil {
			env.Member = guild(e.GuildID).memberEnv(*e.Member)
			env.User = env.Member
		} else {
			env.User = NewUserEnv(e.Author)
			env.Member = env.User
		}
		env.Channel = currentChannelEnv(session, e.ChannelID, nil)
		if e.GuildID != 0 {
			env.Guild = guild(e.GuildID).guildEnv()
		}
		env.Message = NewMessageEnv(e.Message)
	case *gateway.MessageDeleteEvent:
		env.Message = NewMessageEnv(discord.Message{
			ID: e.ID,
		})
		env.Channel = currentChannelEnv(session, e.ChannelID, nil)
		if e.GuildID != 0 {
			env.Guild = guild(e.GuildID).guildEnv()
		}
	case *gateway.GuildMemberAddEvent:
		env.Member = guild(e.GuildID).memberEnv(e.Member)
		env.User = env.Member
		env.Guild = guild(e.GuildID).guildEnv()
	case *gateway.GuildMemberRemoveEvent:
		env.User = NewUserEnv(e.User)
		env.Member = env.User
		env.Guild = guild(e.GuildID).guildEnv()
	case *gateway.MessageReactionAddEvent:
		if e.Member != nil {
			env.Member = guild(e.GuildID).memberEnv(*e.Member)
			env.User = env.Member
		} else {
			// DM reactions (and, in principle, a guild reaction delivered
			// without member data) only give us the user's ID.
			env.User = NewUserIDEnv(e.UserID)
			env.Member = env.User
		}
		env.Channel = currentChannelEnv(session, e.ChannelID, nil)
		if e.GuildID != 0 {
			env.Guild = guild(e.GuildID).guildEnv()
		}
		env.Message = NewMessageEnv(discord.Message{ID: e.MessageID})
		env.Emoji = NewEmojiEnv(e.Emoji)
	case *gateway.MessageReactionRemoveEvent:
		// MESSAGE_REACTION_REMOVE never includes member/user data, only the ID.
		env.User = NewUserIDEnv(e.UserID)
		env.Member = env.User
		env.Channel = currentChannelEnv(session, e.ChannelID, nil)
		if e.GuildID != 0 {
			env.Guild = guild(e.GuildID).guildEnv()
		}
		env.Message = NewMessageEnv(discord.Message{ID: e.MessageID})
		env.Emoji = NewEmojiEnv(e.Emoji)
	case *state.GuildJoinEvent:
		// The cached server knows its member count, the event doesn't.
		joined := e.Guild
		if cached, ok := guild(e.ID).guild(); ok {
			joined = *cached
		}
		env.Guild = guild(e.ID).newGuildEnv(joined)
	case *state.GuildLeaveEvent:
		// The server is gone from the cache by now, so only its ID is left.
		env.Guild = guild(e.ID).guildEnv()
	case *schedule.Event:
		env.Schedule = NewScheduleEnv(e)
	}

	if env.Guild == nil {
		// Typed, so its fields are empty instead of failing, like they are
		// for an interaction outside of a server.
		env.Guild = (*CurrentGuildEnv)(nil)
	}

	return env
}

// SetResumeContext makes the interactions or events from before a resume point
// available to the resumed flow, oldest first. Command args and modal inputs
// only exist on one kind of interaction, so arg() and input() fall back to
// earlier ones, newest first.
func (c Context) SetResumeContext(earlier []Context) {
	if len(earlier) == 0 {
		return
	}

	c.Env["origin"] = map[string]any(earlier[0].Env)
	c.Env["previous"] = map[string]any(earlier[len(earlier)-1].Env)

	for _, name := range []string{"arg", "input"} {
		var lookups []func(string) any
		if fn, ok := c.Env[name].(func(string) any); ok {
			lookups = append(lookups, fn)
		}
		for i := len(earlier) - 1; i >= 0; i-- {
			if fn, ok := earlier[i].Env[name].(func(string) any); ok {
				lookups = append(lookups, fn)
			}
		}

		c.Env[name] = func(key string) any {
			for _, lookup := range lookups {
				if v := lookup(key); v != nil {
					return v
				}
			}
			return nil
		}
	}
}

func NewContext(env Env) Context {
	return Context{
		Env: env,
	}
}

func NewContextFromEvent(event ws.Event, session *state.State) Context {
	env := newEventEnv(event, session)
	return Context{
		Env: Env{
			"event":    env,
			"user":     env.User,
			"member":   env.Member,
			"channel":  env.Channel,
			"guild":    env.Guild,
			"server":   env.Guild,
			"message":  env.Message,
			"emoji":    env.Emoji,
			"schedule": env.Schedule,
			"app":      NewAppEnv(session),
		},
	}
}

type UserEnv struct {
	og discord.User

	ID            string `expr:"id" json:"id"`
	Username      string `expr:"username" json:"username"`
	Discriminator string `expr:"discriminator" json:"discriminator"`
	DisplayName   string `expr:"display_name" json:"display_name"`
	Name          string `expr:"name" json:"name"`
	Mention       string `expr:"mention" json:"mention"`
	AvatarURL     string `expr:"avatar_url" json:"avatar_url"`
	BannerURL     string `expr:"banner_url" json:"banner_url"`
	IsBot         bool   `expr:"is_bot" json:"is_bot"`
	// CreatedAt and JoinedAt are Unix timestamps in seconds, so they can be
	// compared with now().Unix() and put into a Discord timestamp.
	CreatedAt int64 `expr:"created_at" json:"created_at"`

	// The fields below are only filled for members of a server. They are part
	// of every user so a placeholder using them is empty outside of a server
	// instead of failing the flow. The ones that need the roles of the server
	// are in lazyField.
	JoinedAt  int64 `expr:"joined_at" json:"joined_at"`
	RoleCount int   `expr:"role_count" json:"role_count"`

	IsBooster     bool  `expr:"is_booster" json:"is_booster"`
	BoostingSince int64 `expr:"boosting_since" json:"boosting_since"`
	// TimeoutUntil is 0 unless the member is timed out right now.
	IsTimedOut   bool  `expr:"is_timed_out" json:"is_timed_out"`
	TimeoutUntil int64 `expr:"timeout_until" json:"timeout_until"`
}

func (u UserEnv) Thing() thing.Thing {
	return thing.NewDiscordUser(u.og)
}

func (u UserEnv) String() string {
	return u.ID
}

// lazyField has the member fields that need the roles of the server, which
// are empty for a user outside of a server.
func (u *UserEnv) lazyField(name string) (any, bool, error) {
	switch name {
	case "top_role":
		return &RoleEnv{}, true, nil
	case "role_mentions", "role_names", "color":
		return "", true, nil
	case "is_owner", "is_admin":
		return false, true, nil
	case "permissions":
		return []string{}, true, nil
	}
	return nil, false, nil
}

func NewUserEnv(user discord.User) *UserEnv {
	displayName := user.DisplayName
	if displayName == "" {
		displayName = user.Username
	}

	return &UserEnv{
		og: user,

		ID:            user.ID.String(),
		Username:      user.Username,
		Discriminator: user.Discriminator,
		DisplayName:   displayName,
		Name:          displayName,
		Mention:       fmt.Sprintf("<@%s>", user.ID.String()),
		AvatarURL:     user.AvatarURL(),
		BannerURL:     user.BannerURL(),
		IsBot:         user.Bot,
		CreatedAt:     snowflakeUnix(discord.Snowflake(user.ID)),
	}
}

// NewUserIDEnv is a user of whom only the ID is known, like the user of a
// reaction remove event. Everything but id, mention and created_at is empty.
func NewUserIDEnv(id discord.UserID) *UserEnv {
	return &UserEnv{
		og: discord.User{ID: id},

		ID:        id.String(),
		Mention:   fmt.Sprintf("<@%s>", id.String()),
		CreatedAt: snowflakeUnix(discord.Snowflake(id)),
	}
}

// snowflakeUnix returns when a Discord ID was created, or 0 without an ID.
func snowflakeUnix(id discord.Snowflake) int64 {
	if !id.IsValid() {
		return 0
	}
	return id.Time().Unix()
}

type MemberEnv struct {
	og discord.Member

	UserEnv

	Nick    string   `expr:"nick" json:"nick"`
	RoleIDs []string `expr:"role_ids" json:"role_ids"`

	// roles is only called when a placeholder needs the roles of the server,
	// as it copies all of them from the cache.
	roles func() (memberRoles, error)
}

// memberRoles is what the roles of the server tell about a member.
type memberRoles struct {
	// roles are those of the member, highest first.
	roles       []discord.Role
	everyone    *discord.Role
	permissions discord.Permissions
	isOwner     bool
}

var (
	errMemberWithoutServer = errors.New("for a member from a block result or variable")
	errServerNotCached     = errors.New("because the server isn't cached")
)

func (m MemberEnv) String() string {
	return m.UserEnv.String()
}

// NewMemberEnv doesn't know the server of the member, like the result of a
// Get Member block. The fields that need the roles of the server, like
// is_admin, fail instead of being empty or false.
func NewMemberEnv(member discord.Member) *MemberEnv {
	return newMemberEnv(member, func() (memberRoles, error) {
		return memberRoles{}, errMemberWithoutServer
	})
}

func newMemberEnv(member discord.Member, roles func() (memberRoles, error)) *MemberEnv {
	roleIDs := make([]string, len(member.RoleIDs))
	for i, role := range member.RoleIDs {
		roleIDs[i] = role.String()
	}

	env := &MemberEnv{
		og: member,

		UserEnv: *NewUserEnv(member.User),

		Nick:    member.Nick,
		RoleIDs: roleIDs,

		roles: roles,
	}

	if member.Joined.IsValid() {
		env.JoinedAt = member.Joined.Time().Unix()
	}
	env.RoleCount = len(member.RoleIDs)

	if member.BoostedSince.IsValid() {
		env.IsBooster = true
		env.BoostingSince = member.BoostedSince.Time().Unix()
	}
	// Discord keeps the time of a timeout that has ended.
	if until := member.CommunicationDisabledUntil; until.IsValid() && until.Time().After(time.Now()) {
		env.IsTimedOut = true
		env.TimeoutUntil = until.Time().Unix()
	}

	return env
}

// lazyField has the fields that need the roles of the member's server. The
// top role of a member without roles is @everyone. The owner of the server and
// administrators have every permission, whatever their roles are. Permissions
// that a channel grants or denies aren't part of permissions.
func (m *MemberEnv) lazyField(name string) (any, bool, error) {
	switch name {
	case "top_role", "role_mentions", "role_names", "color", "is_owner", "is_admin", "permissions":
	default:
		return nil, false, nil
	}

	res, err := memberRoles{}, errMemberWithoutServer
	if m.roles != nil {
		res, err = m.roles()
	}
	if err != nil {
		// Mentions don't need the server, only their order does.
		if name == "role_mentions" {
			mentions := make([]string, len(m.og.RoleIDs))
			for i, id := range m.og.RoleIDs {
				mentions[i] = id.Mention()
			}
			return strings.Join(mentions, " "), true, nil
		}
		return nil, true, fmt.Errorf("%s isn't known %w", name, err)
	}

	isAdmin := res.isOwner || res.permissions.Has(discord.PermissionAdministrator)

	switch name {
	case "top_role":
		if len(res.roles) > 0 {
			return NewRoleEnv(res.roles[0]), true, nil
		}
		if res.everyone != nil {
			return NewRoleEnv(*res.everyone), true, nil
		}
		return &RoleEnv{}, true, nil
	case "role_mentions":
		mentions := make([]string, len(res.roles))
		for i, role := range res.roles {
			mentions[i] = role.ID.Mention()
		}
		return strings.Join(mentions, " "), true, nil
	case "role_names":
		// The names, highest first, separated by commas.
		names := make([]string, 0, len(res.roles))
		for _, role := range res.roles {
			if role.Name != "" {
				names = append(names, role.Name)
			}
		}
		return strings.Join(names, ", "), true, nil
	case "color":
		// The color of the highest role that has one, like #5865f2.
		for _, role := range res.roles {
			if role.Color > 0 {
				return fmt.Sprintf("#%06x", role.Color.Uint32()), true, nil
			}
		}
		return "", true, nil
	case "is_owner":
		return res.isOwner, true, nil
	case "is_admin":
		return isAdmin, true, nil
	default:
		permissions := res.permissions
		if isAdmin {
			permissions = ^discord.Permissions(0)
		}
		return permissionNames(permissions), true, nil
	}
}

// permissionNameByBit has the names Discord documents for the permission bits,
// lowercased. Bits without a name are left out.
var permissionNameByBit = map[int]string{
	0:  "create_instant_invite",
	1:  "kick_members",
	2:  "ban_members",
	3:  "administrator",
	4:  "manage_channels",
	5:  "manage_guild",
	6:  "add_reactions",
	7:  "view_audit_log",
	8:  "priority_speaker",
	9:  "stream",
	10: "view_channel",
	11: "send_messages",
	12: "send_tts_messages",
	13: "manage_messages",
	14: "embed_links",
	15: "attach_files",
	16: "read_message_history",
	17: "mention_everyone",
	18: "use_external_emojis",
	19: "view_guild_insights",
	20: "connect",
	21: "speak",
	22: "mute_members",
	23: "deafen_members",
	24: "move_members",
	25: "use_vad",
	26: "change_nickname",
	27: "manage_nicknames",
	28: "manage_roles",
	29: "manage_webhooks",
	30: "manage_guild_expressions",
	31: "use_application_commands",
	32: "request_to_speak",
	33: "manage_events",
	34: "manage_threads",
	35: "create_public_threads",
	36: "create_private_threads",
	37: "use_external_stickers",
	38: "send_messages_in_threads",
	39: "use_embedded_activities",
	40: "moderate_members",
	41: "view_creator_monetization_analytics",
	42: "use_soundboard",
	43: "create_guild_expressions",
	44: "create_events",
	45: "use_external_sounds",
	46: "send_voice_messages",
	48: "set_voice_channel_status",
	49: "send_polls",
	50: "use_external_apps",
	51: "pin_messages",
	52: "bypass_slowmode",
}

func permissionNames(permissions discord.Permissions) []string {
	names := []string{}
	for bit := 0; bit < 64; bit++ {
		name, ok := permissionNameByBit[bit]
		if ok && permissions&(1<<bit) != 0 {
			names = append(names, name)
		}
	}
	return names
}

// guildLookup reads what is known about a server from the cache of the
// session. It never asks Discord, since it runs for every flow execution.
type guildLookup struct {
	session *state.State
	guildID discord.GuildID
}

func (l guildLookup) cached() bool {
	return l.session != nil && l.session.Cabinet != nil && l.guildID.IsValid()
}

func (l guildLookup) guild() (*discord.Guild, bool) {
	if !l.cached() {
		return nil, false
	}
	guild, err := l.session.Cabinet.Guild(l.guildID)
	return guild, err == nil
}

// guildEnv falls back to only the ID if the server isn't cached.
func (l guildLookup) guildEnv() *CurrentGuildEnv {
	guild, ok := l.guild()
	if !ok {
		guild = &discord.Guild{ID: l.guildID}
	}
	return &CurrentGuildEnv{GuildEnv: *l.newGuildEnv(*guild)}
}

// guildExtra adds what the cache keeps apart from the server itself: its
// roles, channels and emojis.
func (l guildLookup) guildExtra(guild discord.Guild) guildExtra {
	res := guildExtra{
		roleCount:     countRoles(guild.Roles),
		emojiCount:    len(guild.Emojis),
		rulesChannel:  l.guildChannelEnv(guild.RulesChannelID),
		systemChannel: l.guildChannelEnv(guild.SystemChannelID),
	}
	if !l.cached() {
		return res
	}

	if roles, err := l.session.Cabinet.Roles(l.guildID); err == nil {
		res.roleCount = countRoles(roles)
	}
	if emojis, err := l.session.Cabinet.Emojis(l.guildID); err == nil {
		res.emojiCount = len(emojis)
	}
	if channels, err := l.session.Cabinet.Channels(l.guildID); err == nil {
		for _, channel := range channels {
			// Categories and threads aren't what people count as channels.
			if channel.Type != discord.GuildCategory && !isThread(channel.Type) {
				res.channelCount++
			}
		}
	}
	return res
}

func (l guildLookup) guildChannelEnv(id discord.ChannelID) *ChannelEnv {
	if !id.IsValid() {
		return &ChannelEnv{}
	}
	return channelEnv(l.session, id, nil)
}

func (l guildLookup) memberEnv(member discord.Member) *MemberEnv {
	return newMemberEnv(member, sync.OnceValues(func() (memberRoles, error) {
		return l.memberRoles(member)
	}))
}

func (l guildLookup) memberRoles(member discord.Member) (memberRoles, error) {
	guild, ok := l.guild()
	if !ok {
		return memberRoles{}, errServerNotCached
	}
	guildRoles, err := l.session.Cabinet.Roles(l.guildID)
	if err != nil {
		return memberRoles{}, errServerNotCached
	}

	known := make(map[discord.RoleID]discord.Role, len(guildRoles))
	for _, role := range guildRoles {
		known[role.ID] = role
	}

	res := memberRoles{
		roles:   make([]discord.Role, 0, len(member.RoleIDs)),
		isOwner: guild.OwnerID.IsValid() && member.User.ID == guild.OwnerID,
	}
	// Everyone has the permissions of @everyone.
	if everyone, ok := known[discord.RoleID(l.guildID)]; ok {
		res.everyone = &everyone
		res.permissions = everyone.Permissions
	}
	for _, id := range member.RoleIDs {
		role, ok := known[id]
		if !ok {
			role = discord.Role{ID: id}
		}
		res.roles = append(res.roles, role)
		res.permissions |= role.Permissions
	}
	// Roles at the same position are ordered by age, like Discord does.
	sort.SliceStable(res.roles, func(a, b int) bool {
		if res.roles[a].Position != res.roles[b].Position {
			return res.roles[a].Position > res.roles[b].Position
		}
		return res.roles[a].ID < res.roles[b].ID
	})
	return res, nil
}

func (m MemberEnv) Thing() thing.Thing {
	return thing.NewDiscordMember(m.og)
}

type ChannelEnv struct {
	og discord.Channel

	ID      string `expr:"id" json:"id"`
	Name    string `expr:"name" json:"name"`
	Mention string `expr:"mention" json:"mention"`
	// Type is a word like text, voice or thread, and empty if it isn't known.
	Type string `expr:"type" json:"type"`

	// category is only called when a placeholder needs the category, as it
	// can take two more cache lookups.
	category func() channelCategory
}

// channelCategory is empty for channels outside of a category. For a thread
// it's the category of the channel the thread is in.
type channelCategory struct {
	id   string
	name string
}

// NewChannelEnv only knows what the channel itself carries, so the name of the
// category is missing and a thread has no category.
func NewChannelEnv(channel discord.Channel) *ChannelEnv {
	return newChannelEnv(nil, channel)
}

func newChannelEnv(session *state.State, channel discord.Channel) *ChannelEnv {
	env := &ChannelEnv{
		og: channel,

		ID:      channel.ID.String(),
		Name:    channel.Name,
		Mention: fmt.Sprintf("<#%s>", channel.ID.String()),

		category: sync.OnceValue(func() channelCategory {
			return lookupCategory(session, channel)
		}),
	}

	// A channel that only has an ID would otherwise claim to be a text
	// channel, as that's type 0.
	if channel.Name != "" || channel.Type != discord.GuildText {
		env.Type = channelTypeName(channel.Type)
	}
	return env
}

func (c *ChannelEnv) lazyField(name string) (any, bool, error) {
	if name != "category_id" && name != "category_name" {
		return nil, false, nil
	}

	var category channelCategory
	if c.category != nil {
		category = c.category()
	}
	if name == "category_id" {
		return category.id, true, nil
	}
	return category.name, true, nil
}

func isThread(t discord.ChannelType) bool {
	switch t {
	case discord.GuildAnnouncementThread, discord.GuildPublicThread, discord.GuildPrivateThread:
		return true
	}
	return false
}

// guildMedia is like a forum, but for images and videos. arikawa has no
// constant for it.
const guildMedia discord.ChannelType = 16

func channelTypeName(t discord.ChannelType) string {
	switch t {
	case discord.GuildText:
		return "text"
	case discord.DirectMessage:
		return "dm"
	case discord.GuildVoice:
		return "voice"
	case discord.GroupDM:
		return "group_dm"
	case discord.GuildCategory:
		return "category"
	case discord.GuildAnnouncement:
		return "announcement"
	case discord.GuildAnnouncementThread, discord.GuildPublicThread, discord.GuildPrivateThread:
		return "thread"
	case discord.GuildStageVoice:
		return "stage"
	case discord.GuildForum:
		return "forum"
	case guildMedia:
		return "media"
	}
	return "unknown"
}

func cachedChannel(session *state.State, id discord.ChannelID) (*discord.Channel, bool) {
	if session == nil || session.Cabinet == nil || !id.IsValid() {
		return nil, false
	}
	channel, err := session.Cabinet.Channel(id)
	return channel, err == nil
}

// channelEnv reads the channel from the cache of session, to know its
// category. It never asks Discord, since it runs for every flow execution.
// Without the channel in the cache it falls back to what Discord sent along,
// or only the ID.
func channelEnv(session *state.State, id discord.ChannelID, fallback *discord.Channel) *ChannelEnv {
	channel, ok := cachedChannel(session, id)
	if !ok {
		// A copy, so the channel of the interaction stays untouched.
		partial := discord.Channel{}
		if fallback != nil {
			partial = *fallback
		}
		partial.ID = id
		channel = &partial
	}
	return newChannelEnv(session, *channel)
}

func lookupCategory(session *state.State, channel discord.Channel) channelCategory {
	categoryID := channel.ParentID
	if isThread(channel.Type) {
		categoryID = 0
		if parent, ok := cachedChannel(session, channel.ParentID); ok {
			categoryID = parent.ParentID
		}
	}
	if !categoryID.IsValid() {
		return channelCategory{}
	}

	res := channelCategory{id: categoryID.String()}
	if category, ok := cachedChannel(session, categoryID); ok {
		res.name = category.Name
	}
	return res
}

// CurrentChannelEnv is the channel a flow runs in. A placeholder of just the
// channel is its ID, like it was before the channel had other fields.
type CurrentChannelEnv struct {
	ChannelEnv
}

func (c CurrentChannelEnv) Thing() thing.Thing {
	return thing.NewString(c.ID)
}

func currentChannelEnv(session *state.State, id discord.ChannelID, fallback *discord.Channel) *CurrentChannelEnv {
	return &CurrentChannelEnv{ChannelEnv: *channelEnv(session, id, fallback)}
}

func (c ChannelEnv) Thing() thing.Thing {
	// A channel that isn't set, like the rules channel of a server without one.
	if c.ID == "" {
		return thing.Null
	}
	return thing.NewDiscordChannel(c.og)
}

func (c ChannelEnv) String() string {
	return c.ID
}

type RoleEnv struct {
	og discord.Role

	ID      string `expr:"id" json:"id"`
	Name    string `expr:"name" json:"name"`
	Mention string `expr:"mention" json:"mention"`
}

func NewRoleEnv(role discord.Role) *RoleEnv {
	return &RoleEnv{
		og: role,

		ID:      role.ID.String(),
		Name:    role.Name,
		Mention: fmt.Sprintf("<@&%s>", role.ID.String()),
	}
}

func (r RoleEnv) Thing() thing.Thing {
	// A role that isn't known, like the top role of a user outside of a server.
	if r.ID == "" {
		return thing.Null
	}
	return thing.NewDiscordRole(r.og)
}

func (r RoleEnv) String() string {
	return r.ID
}

type MessageEnv struct {
	og discord.Message

	ID      string `expr:"id" json:"id"`
	Content string `expr:"content" json:"content"`
}

func NewMessageEnv(msg discord.Message) *MessageEnv {
	return &MessageEnv{
		og: msg,

		ID:      msg.ID.String(),
		Content: msg.Content,
	}
}

func (m MessageEnv) Thing() thing.Thing {
	return thing.NewDiscordMessage(m.og)
}

func (m MessageEnv) String() string {
	return m.ID
}

type GuildEnv struct {
	og discord.Guild

	ID      string `expr:"id" json:"id"`
	Name    string `expr:"name" json:"name"`
	IconURL string `expr:"icon_url" json:"icon_url"`
	OwnerID string `expr:"owner_id" json:"owner_id"`
	// MemberCount is 0 if the count isn't known.
	MemberCount int `expr:"member_count" json:"member_count"`
	BoostCount  int `expr:"boost_count" json:"boost_count"`
	// BoostLevel is 0 to 3.
	BoostLevel  int    `expr:"boost_level" json:"boost_level"`
	CreatedAt   int64  `expr:"created_at" json:"created_at"`
	BannerURL   string `expr:"banner_url" json:"banner_url"`
	Description string `expr:"description" json:"description"`
	// VanityURL is the custom invite link of the server, if it has one.
	VanityURL string `expr:"vanity_url" json:"vanity_url"`

	// extra is only called when a placeholder needs it, as it copies all
	// roles and channels of the server from the cache.
	extra func() guildExtra
}

type guildExtra struct {
	// roleCount doesn't count @everyone, and channelCount neither categories
	// nor threads. They are 0 if they aren't known.
	roleCount    int
	channelCount int
	emojiCount   int
	// The channels are empty if the server hasn't set them.
	rulesChannel  *ChannelEnv
	systemChannel *ChannelEnv
}

// NewGuildEnv only knows what the server itself carries, so the names of its
// rules and system channel are missing.
func NewGuildEnv(guild discord.Guild) *GuildEnv {
	return guildLookup{guildID: guild.ID}.newGuildEnv(guild)
}

func (l guildLookup) newGuildEnv(guild discord.Guild) *GuildEnv {
	env := &GuildEnv{
		og: guild,

		ID:          guild.ID.String(),
		Name:        guild.Name,
		IconURL:     guild.IconURL(),
		MemberCount: int(guild.ApproximateMembers),
		BoostCount:  int(guild.NitroBoosters),
		BoostLevel:  int(guild.NitroBoost),
		CreatedAt:   snowflakeUnix(discord.Snowflake(guild.ID)),
		BannerURL:   guild.BannerURL(),
		Description: guild.Description,

		extra: sync.OnceValue(func() guildExtra {
			return l.guildExtra(guild)
		}),
	}

	if guild.OwnerID.IsValid() {
		env.OwnerID = guild.OwnerID.String()
	}
	if guild.VanityURLCode != "" {
		env.VanityURL = "https://discord.gg/" + guild.VanityURLCode
	}
	return env
}

func (g *GuildEnv) lazyField(name string) (any, bool, error) {
	extra := func() guildExtra {
		if g.extra == nil {
			return guildExtra{rulesChannel: &ChannelEnv{}, systemChannel: &ChannelEnv{}}
		}
		return g.extra()
	}

	switch name {
	case "role_count":
		return extra().roleCount, true, nil
	case "channel_count":
		return extra().channelCount, true, nil
	case "emoji_count":
		return extra().emojiCount, true, nil
	case "rules_channel":
		return extra().rulesChannel, true, nil
	case "system_channel":
		return extra().systemChannel, true, nil
	}
	return nil, false, nil
}

// countRoles leaves out @everyone, which every server has.
func countRoles(roles []discord.Role) int {
	if len(roles) == 0 {
		return 0
	}
	return len(roles) - 1
}

func (g GuildEnv) Thing() thing.Thing {
	return thing.NewDiscordGuild(g.og)
}

func (g GuildEnv) String() string {
	return g.ID
}

// CurrentGuildEnv is the server a flow runs in. A placeholder of just the
// server is its ID, like it was before the server had other fields. Outside
// of a server it's nil, but its fields are empty instead of failing the flow,
// see fetchField.
type CurrentGuildEnv struct {
	GuildEnv
}

func (g CurrentGuildEnv) Thing() thing.Thing {
	return thing.NewString(g.ID)
}

type AttachmentEnv struct {
	ID       string `expr:"id" json:"id"`
	URL      string `expr:"url" json:"url"`
	ProxyURL string `expr:"proxy_url" json:"proxy_url"`
	Filename string `expr:"filename" json:"filename"`
}

func NewAttachmentEnv(attachment *discord.Attachment) *AttachmentEnv {
	return &AttachmentEnv{
		ID:       attachment.ID.String(),
		URL:      attachment.URL,
		ProxyURL: attachment.Proxy,
		Filename: attachment.Filename,
	}
}

func (a AttachmentEnv) String() string {
	return a.URL
}

type EmojiEnv struct {
	og discord.Emoji

	ID       string `expr:"id" json:"id"`
	Name     string `expr:"name" json:"name"`
	Animated bool   `expr:"animated" json:"animated"`
	// Mention is the emoji formatted the way Discord clients render it, e.g.
	// "🔥" for a unicode emoji or "<:name:id>" for a custom one.
	Mention string `expr:"mention" json:"mention"`
}

func NewEmojiEnv(emoji discord.Emoji) *EmojiEnv {
	id := ""
	if emoji.ID.IsValid() {
		id = emoji.ID.String()
	}

	return &EmojiEnv{
		og: emoji,

		ID:       id,
		Name:     emoji.Name,
		Animated: emoji.Animated,
		Mention:  emoji.String(),
	}
}

func (e EmojiEnv) String() string {
	return e.og.String()
}

type SnowflakeEnv struct {
	ID string `expr:"id" json:"id"`
}

func NewSnowflakeEnv[T fmt.Stringer](id T) *SnowflakeEnv {
	return &SnowflakeEnv{
		ID: id.String(),
	}
}

func (s SnowflakeEnv) String() string {
	return s.ID
}

type HTTPResponseEnv struct {
	og thing.HTTPResponseValue

	Status     string                 `expr:"status" json:"status"`
	StatusCode int                    `expr:"status_code" json:"status_code"`
	BodyFunc   func() (string, error) `expr:"body" json:"-"`
	DataFunc   func() (any, error)    `expr:"data" json:"-"`
}

func NewHTTPResponseEnv(resp thing.HTTPResponseValue) *HTTPResponseEnv {
	res := &HTTPResponseEnv{
		og: resp,

		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		BodyFunc: func() (string, error) {
			return string(resp.Body), nil
		},
		DataFunc: func() (any, error) {
			var v any
			if err := json.Unmarshal(resp.Body, &v); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
			}
			return v, nil
		},
	}

	return res
}

func (h HTTPResponseEnv) Thing() thing.Thing {
	return thing.NewHTTPResponse(h.og)
}

func (h HTTPResponseEnv) String() string {
	return h.Status
}

func NewThingEnv(t thing.Thing) any {
	switch t.Type {
	case thing.TypeString:
		return t.String()
	case thing.TypeInt:
		return t.Int()
	case thing.TypeFloat:
		return t.Float()
	case thing.TypeBool:
		return t.Bool()
	case thing.TypeDiscordMessage:
		return NewMessageEnv(t.DiscordMessage())
	case thing.TypeDiscordUser:
		return NewUserEnv(t.DiscordUser())
	case thing.TypeDiscordMember:
		return NewMemberEnv(t.DiscordMember())
	case thing.TypeDiscordChannel:
		return NewChannelEnv(t.DiscordChannel())
	case thing.TypeDiscordGuild:
		return NewGuildEnv(t.DiscordGuild())
	case thing.TypeDiscordRole:
		return NewRoleEnv(t.DiscordRole())
	case thing.TypeHTTPResponse:
		return NewHTTPResponseEnv(t.HTTPResponse())
	case thing.TypeRobloxUser:
		return NewRobloxUserEnv(t.RobloxUser())
	case thing.TypeArray:
		res := make([]any, len(t.Array()))
		for i, v := range t.Array() {
			res[i] = NewThingEnv(v)
		}
		return res
	case thing.TypeObject:
		res := make(map[string]any, len(t.Object()))
		for k, v := range t.Object() {
			res[k] = NewThingEnv(v)
		}
		return res
	default:
		return t.Value
	}
}

type AppEnv struct {
	User *UserEnv `expr:"user" json:"user"`
}

func NewAppEnv(session *state.State) *AppEnv {
	user := session.Ready().User

	return &AppEnv{
		User: NewUserEnv(user),
	}
}

type RobloxUserEnv struct {
	og thing.RobloxUserValue

	ID                     int64  `expr:"id" json:"id"`
	Name                   string `expr:"name" json:"name"`
	DisplayName            string `expr:"display_name" json:"display_name"`
	Description            string `expr:"description" json:"description"`
	CreatedAt              string `expr:"created_at" json:"created_at"`
	IsBanned               bool   `expr:"is_banned" json:"is_banned"`
	HasVerifiedBadge       bool   `expr:"has_verified_badge" json:"has_verified_badge"`
	ExternalAppDisplayName string `expr:"external_app_display_name" json:"external_app_display_name"`
}

func (r RobloxUserEnv) String() string {
	return strconv.FormatInt(r.ID, 10)
}

func NewRobloxUserEnv(user thing.RobloxUserValue) *RobloxUserEnv {
	return &RobloxUserEnv{
		og: user,

		ID:                     user.ID,
		Name:                   user.Name,
		DisplayName:            user.DisplayName,
		Description:            user.Description,
		CreatedAt:              user.CreatedAt,
		IsBanned:               user.IsBanned,
		HasVerifiedBadge:       user.HasVerifiedBadge,
		ExternalAppDisplayName: user.ExternalAppDisplayName,
	}
}
