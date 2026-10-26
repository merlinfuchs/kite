package eval

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
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
	"github.com/kitecloud/kite/kite-service/pkg/auditlog"
	"github.com/kitecloud/kite/kite-service/pkg/guildevent"
	"github.com/kitecloud/kite/kite-service/pkg/schedule"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
)

type Context struct {
	Env      Env
	Patchers []ast.Visitor
	// Raw keeps results as they are, without turning \n into a newline, for
	// templates that have to match exactly, like the message a signature signs.
	Raw bool
	// KeepUnresolved keeps placeholders that fail or are empty as they are
	// written, for messages sent without an interaction or event to fill them.
	KeepUnresolved bool
}

type Env map[string]any

type InteractionEnv struct {
	interaction *discord.InteractionEvent
	picked      map[string]pickedInput

	ID         string                   `expr:"id" json:"id"`
	Channel    *CurrentChannelEnv       `expr:"channel" json:"channel"`
	Guild      *CurrentGuildEnv         `expr:"guild" json:"guild"`
	User       any                      `expr:"user" json:"user"`
	Member     any                      `expr:"member" json:"member"`
	Command    *CommandEnv              `expr:"command" json:"command"`
	Components map[string]*ComponentEnv `expr:"components" json:"components"`
	// Values are the options picked in a select menu, or the IDs of the users,
	// roles or channels picked in the other selects. Value is the first of them.
	Values []string `expr:"values" json:"values"`
	Value  string   `expr:"value" json:"value"`
}

// NewInteractionEnv fills in what the interaction doesn't carry itself, like
// the name of the server, from the cache of session.
func NewInteractionEnv(i *discord.InteractionEvent, session *state.State) *InteractionEnv {
	guild := guildLookup{session: session, guildID: i.GuildID}

	e := &InteractionEnv{
		interaction: i,
		picked:      newPickedInputs(i, guild),

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

	if values, ok := selectedValues(i.Data); ok {
		e.Values = values
		if len(values) > 0 {
			e.Value = values[0]
		}
	}

	return e
}

// selectedValues returns what was picked in a select menu: the option values
// of a string select, or the IDs of the users, roles or channels of the others.
func selectedValues(data discord.InteractionData) ([]string, bool) {
	switch d := data.(type) {
	case *discord.StringSelectInteraction:
		return d.Values, true
	case *discord.UserSelectInteraction:
		return stringValues(d.Values), true
	case *discord.RoleSelectInteraction:
		return stringValues(d.Values), true
	case *discord.MentionableSelectInteraction:
		return stringValues(d.Values), true
	case *discord.ChannelSelectInteraction:
		return stringValues(d.Values), true
	}
	return nil, false
}

func stringValues[T fmt.Stringer](values []T) []string {
	res := make([]string, len(values))
	for i, v := range values {
		res[i] = v.String()
	}
	return res
}

func NewContextFromInteraction(i *discord.InteractionEvent, session *state.State) Context {
	interactionEnv := NewInteractionEnv(i, session)

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
			"inputs": func(customID string) any {
				if interactionEnv.Components == nil {
					return nil
				}

				if component, ok := interactionEnv.Components[customID]; ok {
					return component.Values
				}
				return nil
			},
			// picked() and pickedAll() are like input() and inputs(), but return
			// the users, roles or channels picked in a select, or the uploaded
			// files, instead of their IDs.
			"picked": func(customID string) any {
				p, ok := interactionEnv.picked[customID]
				if !ok {
					return nil
				}
				if len(p.envs) == 0 {
					return p.none
				}
				return p.envs[0]
			},
			"pickedAll": func(customID string) any {
				if p, ok := interactionEnv.picked[customID]; ok {
					return p.envs
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
	// Values, which inputs() returns, are the options picked in a select menu
	// or checkbox group, the IDs of uploaded files, or the one value of other
	// inputs, if any. Value, which input() returns, is the first of them, like
	// interaction.value.
	Values []string `expr:"values" json:"values"`
}

func NewComponentsEnv(i *discord.InteractionEvent) map[string]*ComponentEnv {
	components := make(map[string]*ComponentEnv)

	data, ok := i.Data.(*discord.ModalInteraction)
	if !ok {
		return components
	}

	add := func(component discord.Component) {
		c := NewComponentEnv(component)
		if c != nil {
			components[c.CustomID] = c
		}
	}

	for _, row := range data.Components {
		switch row := row.(type) {
		case *discord.ActionRowComponent:
			for _, component := range *row {
				add(component)
			}
		case *discord.LabelComponent:
			if row.Component != nil {
				add(row.Component)
			}
		}
	}

	return components
}

// pickedInput is what was picked in a user, role, mentionable or channel select
// of a modal, or uploaded to a file upload. none is what picked() returns if nothing was picked: a typed nil,
// so its fields are empty and earlier modals aren't looked at instead.
type pickedInput struct {
	envs []any
	none any
}

// newPickedInputs returns the picks of a modal's selects by custom ID, filled
// in from the resolved data Discord sends with the submission.
func newPickedInputs(i *discord.InteractionEvent, guild guildLookup) map[string]pickedInput {
	res := make(map[string]pickedInput)

	data, ok := i.Data.(*discord.ModalInteraction)
	if !ok {
		return res
	}
	resolved := data.Resolved

	user := func(id discord.UserID) any {
		user, ok := resolved.Users[id]
		if !ok {
			user = discord.User{ID: id}
		}
		if member, ok := resolved.Members[id]; ok {
			member.User = user
			return guild.memberEnv(member)
		}
		return NewUserEnv(user)
	}
	role := func(id discord.RoleID) any {
		role, ok := resolved.Roles[id]
		if !ok {
			role = discord.Role{ID: id}
		}
		return NewRoleEnv(role)
	}
	channel := func(id discord.ChannelID) any {
		var fallback *discord.Channel
		if channel, ok := resolved.Channels[id]; ok {
			fallback = &channel
		}
		return channelEnv(guild.session, id, fallback)
	}
	attachment := func(id discord.Snowflake) any {
		attachment, ok := resolved.Attachments[discord.AttachmentID(id)]
		if !ok {
			attachment = discord.Attachment{ID: discord.AttachmentID(id)}
		}
		return NewAttachmentEnv(&attachment)
	}
	// A mentionable select picks users and roles, which Discord tells apart
	// only by where it puts them in the resolved data.
	mentionable := func(id discord.Snowflake) any {
		if _, ok := resolved.Roles[discord.RoleID(id)]; ok {
			return role(discord.RoleID(id))
		}
		return user(discord.UserID(id))
	}

	add := func(component discord.Component) {
		switch c := component.(type) {
		case *discord.UserSelectComponent:
			res[string(c.CustomID)] = pickedInput{envs: mapEnvs(c.Values, user), none: (*UserEnv)(nil)}
		case *discord.RoleSelectComponent:
			res[string(c.CustomID)] = pickedInput{envs: mapEnvs(c.Values, role), none: (*RoleEnv)(nil)}
		case *discord.MentionableSelectComponent:
			res[string(c.CustomID)] = pickedInput{envs: mapEnvs(c.Values, mentionable), none: (*UserEnv)(nil)}
		case *discord.ChannelSelectComponent:
			res[string(c.CustomID)] = pickedInput{envs: mapEnvs(c.Values, channel), none: (*ChannelEnv)(nil)}
		case *discord.FileUploadComponent:
			res[string(c.CustomID)] = pickedInput{envs: mapEnvs(c.Values, attachment), none: (*AttachmentEnv)(nil)}
		}
	}

	for _, row := range data.Components {
		switch row := row.(type) {
		case *discord.ActionRowComponent:
			for _, component := range *row {
				add(component)
			}
		case *discord.LabelComponent:
			if row.Component != nil {
				add(row.Component)
			}
		}
	}

	return res
}

func mapEnvs[T any](values []T, env func(T) any) []any {
	res := make([]any, len(values))
	for i, v := range values {
		res[i] = env(v)
	}
	return res
}

func NewComponentEnv(component discord.Component) *ComponentEnv {
	switch c := component.(type) {
	case *discord.TextInputComponent:
		return newComponentEnv(c.CustomID, optionalValue(c.Value))
	case *discord.StringSelectComponent:
		return newComponentEnv(c.CustomID, c.Values)
	case *discord.UserSelectComponent:
		return newComponentEnv(c.CustomID, c.Values)
	case *discord.RoleSelectComponent:
		return newComponentEnv(c.CustomID, c.Values)
	case *discord.MentionableSelectComponent:
		return newComponentEnv(c.CustomID, c.Values)
	case *discord.ChannelSelectComponent:
		return newComponentEnv(c.CustomID, c.Values)
	case *discord.CheckboxGroupComponent:
		return newComponentEnv(c.CustomID, c.Values)
	case *discord.RadioGroupComponent:
		return newComponentEnv(c.CustomID, optionalValue(c.Value))
	case *discord.CheckboxComponent:
		return newComponentEnv(c.CustomID, []string{strconv.FormatBool(c.Value)})
	case *discord.FileUploadComponent:
		return newComponentEnv(c.CustomID, c.Values)
	}

	return nil
}

func newComponentEnv[T any](customID discord.ComponentID, values []T) *ComponentEnv {
	env := &ComponentEnv{
		CustomID: string(customID),
		Values:   make([]string, len(values)),
	}
	for i, v := range values {
		env.Values[i] = fmt.Sprint(v)
	}
	if len(env.Values) > 0 {
		env.Value = env.Values[0]
	}
	return env
}

func optionalValue(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
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
	Webhook  *WebhookEnv  `expr:"webhook" json:"webhook"`
	AuditLog *AuditLogEnv `expr:"audit_log" json:"audit_log"`
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

// AuditLogEnv is an entry of a server's audit log. Whoever did it is the user
// of the event.
type AuditLogEnv struct {
	ID string `expr:"id" json:"id"`
	// Action is what happened, like member_ban_add, as the Event Filter
	// matches it.
	Action string `expr:"action" json:"action"`
	Reason string `expr:"reason" json:"reason"`
	// Target is the mention of the member, role or channel the action was
	// done to, and the ID of other targets.
	Target   string `expr:"target" json:"target"`
	TargetID string `expr:"target_id" json:"target_id"`
	// Changes lists what changed, one setting per line.
	Changes      string `expr:"changes" json:"changes"`
	RolesAdded   string `expr:"roles_added" json:"roles_added"`
	RolesRemoved string `expr:"roles_removed" json:"roles_removed"`
	// TimeoutUntil is when a timeout ends as a Unix timestamp, 0 if the
	// action isn't a timeout.
	TimeoutUntil int64 `expr:"timeout_until" json:"timeout_until"`
}

func NewAuditLogEnv(e discord.AuditLogEntry) *AuditLogEnv {
	res := &AuditLogEnv{
		ID:           e.ID.String(),
		Action:       auditlog.Action(e),
		Reason:       e.Reason,
		Changes:      auditlog.Changes(e),
		RolesAdded:   auditlog.RoleNames(auditlog.RolesAdded(e)),
		RolesRemoved: auditlog.RoleNames(auditlog.RolesRemoved(e)),
	}
	if e.TargetID.IsValid() {
		res.TargetID = e.TargetID.String()
		switch auditlog.Target(e) {
		case auditlog.TargetUser:
			res.Target = discord.UserID(e.TargetID).Mention()
		case auditlog.TargetRole:
			res.Target = discord.RoleID(e.TargetID).Mention()
		case auditlog.TargetChannel:
			res.Target = discord.ChannelID(e.TargetID).Mention()
		default:
			res.Target = res.TargetID
		}
	}
	if until := auditlog.TimeoutUntil(e); !until.IsZero() {
		res.TimeoutUntil = until.Unix()
	}
	return res
}

func (e AuditLogEnv) String() string {
	return e.Action
}

type WebhookEnv struct {
	Headers map[string]string `expr:"headers" json:"headers"`
	Query   map[string]string `expr:"query" json:"query"`
	Body    string            `expr:"body" json:"body"`
	// Data is the body parsed as JSON, nil if it isn't JSON.
	Data any `expr:"data" json:"data"`
}

func NewWebhookEnv(e *webhook.Event) *WebhookEnv {
	res := &WebhookEnv{
		Headers: e.Headers,
		Query:   e.Query,
		Body:    e.Body,
	}
	// Senders don't reliably set the content type, so the body decides.
	dec := json.NewDecoder(strings.NewReader(e.Body))
	// Keeps IDs written as numbers exact.
	dec.UseNumber()
	var data any
	if err := dec.Decode(&data); err == nil && !dec.More() {
		res.Data = jsonNumbers(data)
	}
	return res
}

// jsonNumbers turns the numbers of a document decoded with UseNumber into
// int64, or float64 if they aren't integers or don't fit, so expressions can
// compare and calculate with them.
func jsonNumbers(v any) any {
	switch v := v.(type) {
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return i
		}
		f, _ := v.Float64()
		return f
	case map[string]any:
		for key, item := range v {
			v[key] = jsonNumbers(item)
		}
	case []any:
		for i, item := range v {
			v[i] = jsonNumbers(item)
		}
	}
	return v
}

func (w WebhookEnv) String() string {
	return w.Body
}

// NewEventEnv fills in what the event doesn't carry itself, like the name of
// the server, from the cache of session.
func NewEventEnv(event ws.Event, session *state.State) *EventEnv {
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
	case *guildevent.MessageDelete:
		// Discord only sends the IDs, the rest is known if the bot saw the
		// message before it was deleted.
		msg := discord.Message{ID: e.ID, ChannelID: e.ChannelID}
		if e.Message != nil {
			msg = *e.Message
			if member, ok := guild(e.GuildID).member(msg.Author.ID); ok {
				// The message doesn't keep the member of its author, but the
				// state caches it when the message is created.
				member.User = msg.Author
				env.Member = guild(e.GuildID).memberEnv(member)
				env.User = env.Member
			} else {
				env.User = NewUserEnv(msg.Author)
				env.Member = env.User
			}
		}
		env.Message = NewMessageEnv(msg)
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
		env.Message = NewMessageEnv(discord.Message{ID: e.MessageID, ChannelID: e.ChannelID})
		env.Emoji = NewEmojiEnv(e.Emoji)
	case *gateway.MessageReactionRemoveEvent:
		// MESSAGE_REACTION_REMOVE never includes member/user data, only the ID.
		env.User = NewUserIDEnv(e.UserID)
		env.Member = env.User
		env.Channel = currentChannelEnv(session, e.ChannelID, nil)
		if e.GuildID != 0 {
			env.Guild = guild(e.GuildID).guildEnv()
		}
		env.Message = NewMessageEnv(discord.Message{ID: e.MessageID, ChannelID: e.ChannelID})
		env.Emoji = NewEmojiEnv(e.Emoji)
	case *state.GuildJoinEvent:
		joined := e.Guild
		joined.ApproximateMembers = e.MemberCount
		env.Guild = guild(e.ID).newGuildEnv(joined)
	case *state.GuildLeaveEvent:
		// The server is gone from the cache by now, so only its ID is left.
		env.Guild = guild(e.ID).guildEnv()
	case *schedule.Event:
		env.Schedule = NewScheduleEnv(e)
	case *gateway.GuildAuditLogEntryCreateEvent:
		if member, ok := guild(e.GuildID).member(e.UserID); ok {
			env.Member = guild(e.GuildID).memberEnv(member)
		} else {
			env.Member = NewUserIDEnv(e.UserID)
		}
		env.User = env.Member
		env.Guild = guild(e.GuildID).guildEnv()
		if id := auditlog.ChannelID(e.AuditLogEntry); id.IsValid() {
			env.Channel = currentChannelEnv(session, id, nil)
		}
		env.AuditLog = NewAuditLogEnv(e.AuditLogEntry)
	case *webhook.Event:
		env.Webhook = NewWebhookEnv(e)
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
// only exist on one kind of interaction, so arg(), input(), inputs(), picked()
// and pickedAll() fall back to earlier ones, newest first. Without earlier
// ones, origin and previous are the current interaction or event, so they work
// before a resume point too.
func (c Context) SetResumeContext(earlier []Context) {
	if len(earlier) == 0 {
		// A copy, so the env doesn't contain itself.
		env := map[string]any(maps.Clone(c.Env))
		c.Env["origin"] = env
		c.Env["previous"] = env
		return
	}

	c.Env["origin"] = map[string]any(earlier[0].Env)
	c.Env["previous"] = map[string]any(earlier[len(earlier)-1].Env)

	for _, name := range []string{"arg", "input", "inputs", "picked", "pickedAll"} {
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
	env := NewEventEnv(event, session)
	return Context{
		Env: Env{
			"event":     env,
			"user":      env.User,
			"member":    env.Member,
			"channel":   env.Channel,
			"guild":     env.Guild,
			"server":    env.Guild,
			"message":   env.Message,
			"emoji":     env.Emoji,
			"schedule":  env.Schedule,
			"webhook":   env.Webhook,
			"audit_log": env.AuditLog,
			"app":       NewAppEnv(session),
		},
	}
}

type UserEnv struct {
	og discord.User

	ID            string `expr:"id" json:"id"`
	Username      string `expr:"username" json:"username"`
	Discriminator string `expr:"discriminator" json:"discriminator"`
	DisplayName   string `expr:"display_name" json:"display_name" desc:"The display name of the account, or the username without one"`
	Name          string `expr:"name" json:"name" desc:"Same as display_name"`
	Mention       string `expr:"mention" json:"mention"`
	AvatarURL     string `expr:"avatar_url" json:"avatar_url"`
	BannerURL     string `expr:"banner_url" json:"banner_url" desc:"Empty if the user has no banner"`
	IsBot         bool   `expr:"is_bot" json:"is_bot"`
	// CreatedAt and JoinedAt are Unix timestamps in seconds, so they can be
	// compared with now().Unix() and put into a Discord timestamp.
	CreatedAt int64 `expr:"created_at" json:"created_at" desc:"When the account was created, as a Unix timestamp in seconds"`

	// The fields below are only filled for members of a server. They are part
	// of every user so a placeholder using them is empty outside of a server
	// instead of failing the flow. The ones that need the roles of the server
	// are in lazyField. scope:"member" leaves them out of the user type of the
	// editor, as they are always empty for users from a block result.
	Nick      string   `expr:"nick" json:"nick" scope:"member" desc:"The nickname in the server, empty without one"`
	RoleIDs   []string `expr:"role_ids" json:"role_ids" scope:"member"`
	JoinedAt  int64    `expr:"joined_at" json:"joined_at" scope:"member" desc:"When the user joined the server, as a Unix timestamp in seconds"`
	RoleCount int      `expr:"role_count" json:"role_count" scope:"member"`

	IsBooster     bool  `expr:"is_booster" json:"is_booster" scope:"member"`
	BoostingSince int64 `expr:"boosting_since" json:"boosting_since" scope:"member" desc:"Unix timestamp in seconds, 0 if the user isn't boosting"`
	IsTimedOut    bool  `expr:"is_timed_out" json:"is_timed_out" scope:"member"`
	TimeoutUntil  int64 `expr:"timeout_until" json:"timeout_until" scope:"member" desc:"Unix timestamp in seconds, 0 if the user isn't timed out"`
}

func (u UserEnv) Thing() thing.Thing {
	return thing.NewDiscordUser(u.og)
}

func (u UserEnv) String() string {
	return u.ID
}

// userLazyFields are the fields of lazyField of a user and a member.
// scope:"server" marks those that fail for a member from a block result, as
// it doesn't know its server.
type userLazyFields struct {
	TopRole      *RoleEnv `expr:"top_role" scope:"server" desc:"The highest role of the user, or @everyone without roles"`
	RoleMentions string   `expr:"role_mentions" scope:"member" desc:"Mentions of all roles of the user, highest first"`
	RoleNames    string   `expr:"role_names" scope:"server" desc:"Names of all roles of the user, highest first, separated by commas"`
	Color        string   `expr:"color" scope:"server" desc:"Color of the highest role that has one, like #5865f2"`
	IsOwner      bool     `expr:"is_owner" scope:"server" desc:"Whether the user owns the server"`
	IsAdmin      bool     `expr:"is_admin" scope:"server" desc:"Whether the user owns the server or has the Administrator permission"`
	Permissions  []string `expr:"permissions" scope:"server" desc:"Lowercased permission names, like ban_members or manage_messages"`
}

func (u UserEnv) lazyFields() any {
	return userLazyFields{}
}

// lazyField has the member fields that need the roles of the server, which
// are empty for a user outside of a server.
func (u UserEnv) lazyField(name string) (any, bool, error) {
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
		RoleIDs:       []string{},
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
		RoleIDs:   []string{},
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

		roles: roles,
	}
	env.Nick = member.Nick
	env.RoleIDs = roleIDs

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
func (m MemberEnv) lazyField(name string) (any, bool, error) {
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
			if color := roleColor(role); color != "" {
				return color, true, nil
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

func (l guildLookup) member(userID discord.UserID) (discord.Member, bool) {
	if !l.cached() || !userID.IsValid() {
		return discord.Member{}, false
	}
	member, err := l.session.Cabinet.Member(l.guildID, userID)
	if err != nil {
		return discord.Member{}, false
	}
	return *member, true
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
	Type    string `expr:"type" json:"type" desc:"text, voice, category, announcement, thread, stage, forum, media, dm, group_dm, or unknown for newer types, empty if it isn't known"`
	GuildID string `expr:"guild_id" json:"guild_id" desc:"The ID of the server the channel is in, empty for direct messages and if it isn't known"`

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

	if channel.GuildID.IsValid() {
		env.GuildID = channel.GuildID.String()
	}

	// A channel that only has an ID would otherwise claim to be a text
	// channel, as that's type 0.
	if channel.Name != "" || channel.Type != discord.GuildText {
		env.Type = channelTypeName(channel.Type)
	}
	return env
}

type channelLazyFields struct {
	CategoryID   string `expr:"category_id" desc:"Empty outside of a category, for a thread the category of its channel"`
	CategoryName string `expr:"category_name"`
}

func (c ChannelEnv) lazyFields() any {
	return channelLazyFields{}
}

func (c ChannelEnv) lazyField(name string) (any, bool, error) {
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

	ID          string `expr:"id" json:"id"`
	Name        string `expr:"name" json:"name"`
	Mention     string `expr:"mention" json:"mention"`
	Color       string `expr:"color" json:"color" desc:"Like #5865f2, empty if the role has no color"`
	Hoist       bool   `expr:"hoist" json:"hoist" desc:"Whether the role is shown separately in the member list"`
	Mentionable bool   `expr:"mentionable" json:"mentionable"`
}

func NewRoleEnv(role discord.Role) *RoleEnv {
	return &RoleEnv{
		og: role,

		ID:          role.ID.String(),
		Name:        role.Name,
		Mention:     fmt.Sprintf("<@&%s>", role.ID.String()),
		Color:       roleColor(role),
		Hoist:       role.Hoist,
		Mentionable: role.Mentionable,
	}
}

func roleColor(role discord.Role) string {
	if role.Color <= 0 {
		return ""
	}
	return fmt.Sprintf("#%06x", role.Color.Uint32())
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

	ID        string `expr:"id" json:"id"`
	ChannelID string `expr:"channel_id" json:"channel_id"`
	Content   string `expr:"content" json:"content"`
	// Author is nil if only the ID of the message is known, like for a
	// reaction, but its fields are empty instead of failing, see fetchField.
	Author    *UserEnv       `expr:"author" json:"author" desc:"Empty if only the ID of the message is known, like for reactions"`
	Reactions []*ReactionEnv `expr:"reactions" json:"reactions" desc:"Empty for the message of an event, get the message with a block to count its reactions"`
	Mentions  []*UserEnv     `expr:"mentions" json:"mentions" desc:"Users the message pings, including the author of the message it replies to if the reply pings them"`
}

func NewMessageEnv(msg discord.Message) *MessageEnv {
	env := &MessageEnv{
		og: msg,

		ID:        msg.ID.String(),
		ChannelID: msg.ChannelID.String(),
		Content:   msg.Content,
		Reactions: make([]*ReactionEnv, len(msg.Reactions)),
		Mentions:  make([]*UserEnv, len(msg.Mentions)),
	}
	if msg.Author.ID.IsValid() {
		env.Author = NewUserEnv(msg.Author)
	}
	for i, r := range msg.Reactions {
		env.Reactions[i] = &ReactionEnv{Emoji: NewEmojiEnv(r.Emoji), Count: r.Count}
	}
	for i, u := range msg.Mentions {
		env.Mentions[i] = NewUserEnv(u.User)
	}
	return env
}

type ReactionEnv struct {
	Emoji *EmojiEnv `expr:"emoji" json:"emoji"`
	Count int       `expr:"count" json:"count" desc:"How many users reacted with the emoji"`
}

func (r ReactionEnv) String() string {
	return r.Emoji.String()
}

func (m MessageEnv) Thing() thing.Thing {
	return thing.NewDiscordMessage(m.og)
}

func (m MessageEnv) String() string {
	return m.ID
}

type GuildEnv struct {
	og discord.Guild

	ID          string `expr:"id" json:"id"`
	Name        string `expr:"name" json:"name"`
	IconURL     string `expr:"icon_url" json:"icon_url" desc:"Empty if the server has no icon"`
	OwnerID     string `expr:"owner_id" json:"owner_id" desc:"The ID of the user that owns the server"`
	MemberCount int    `expr:"member_count" json:"member_count" desc:"0 if the count isn't known"`
	BoostCount  int    `expr:"boost_count" json:"boost_count"`
	BoostLevel  int    `expr:"boost_level" json:"boost_level" desc:"0 to 3"`
	CreatedAt   int64  `expr:"created_at" json:"created_at" desc:"When the server was created, as a Unix timestamp in seconds"`
	BannerURL   string `expr:"banner_url" json:"banner_url" desc:"Empty if the server has no banner"`
	Description string `expr:"description" json:"description"`
	VanityURL   string `expr:"vanity_url" json:"vanity_url" desc:"The custom invite link, empty if the server has none"`

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

type guildLazyFields struct {
	RoleCount     int         `expr:"role_count" desc:"Without @everyone"`
	ChannelCount  int         `expr:"channel_count" desc:"Without categories and threads"`
	EmojiCount    int         `expr:"emoji_count"`
	RulesChannel  *ChannelEnv `expr:"rules_channel" desc:"Empty if the server has no rules channel"`
	SystemChannel *ChannelEnv `expr:"system_channel" desc:"Where Discord posts join and boost messages, empty if turned off"`
}

func (g GuildEnv) lazyFields() any {
	return guildLazyFields{}
}

func (g GuildEnv) lazyField(name string) (any, bool, error) {
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
	ID          string `expr:"id" json:"id"`
	URL         string `expr:"url" json:"url"`
	ProxyURL    string `expr:"proxy_url" json:"proxy_url"`
	Filename    string `expr:"filename" json:"filename"`
	ContentType string `expr:"content_type" json:"content_type" desc:"Like image/png"`
	Size        uint64 `expr:"size" json:"size" desc:"In bytes"`
}

func NewAttachmentEnv(attachment *discord.Attachment) *AttachmentEnv {
	return &AttachmentEnv{
		ID:          attachment.ID.String(),
		URL:         attachment.URL,
		ProxyURL:    attachment.Proxy,
		Filename:    attachment.Filename,
		ContentType: attachment.ContentType,
		Size:        attachment.Size,
	}
}

func (a AttachmentEnv) String() string {
	return a.URL
}

type EmojiEnv struct {
	og discord.Emoji

	ID       string `expr:"id" json:"id" desc:"Empty for a unicode emoji"`
	Name     string `expr:"name" json:"name"`
	Animated bool   `expr:"animated" json:"animated"`
	Mention  string `expr:"mention" json:"mention" desc:"The emoji the way Discord shows it, like 🔥 or <:name:id>"`
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

	Status     string                 `expr:"status" json:"status" desc:"Status line, like 200 OK"`
	StatusCode int                    `expr:"status_code" json:"status_code"`
	BodyFunc   func() (string, error) `expr:"body" json:"-" desc:"Body of the response as text"`
	DataFunc   func() (any, error)    `expr:"data" json:"-" desc:"Body of the response parsed as JSON"`
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
	session      *state.State
	guilds       func() appGuilds
	userInstalls func() (int, error)

	User *UserEnv `expr:"user" json:"user"`
}

func NewAppEnv(session *state.State) *AppEnv {
	return newAppEnv(session, session.Ready().User)
}

func newAppEnv(session *state.State, user discord.User) *AppEnv {
	return &AppEnv{
		session: session,
		guilds: sync.OnceValue(func() appGuilds {
			return countAppGuilds(session)
		}),

		User: NewUserEnv(user),
	}
}

// SetUserInstallCount sets where user_install_count comes from. fn is only
// called once, when an expression first reads it.
func (a *AppEnv) SetUserInstallCount(fn func() (int, error)) {
	a.userInstalls = sync.OnceValues(fn)
}

type appGuilds struct {
	count   int
	members int
}

// countAppGuilds counts the servers of the app and their members from the
// cache, which keeps the member count of each server.
func countAppGuilds(session *state.State) appGuilds {
	guilds, err := session.Cabinet.Guilds()
	if err != nil {
		// The cache returns an error when it has no servers.
		return appGuilds{}
	}

	res := appGuilds{count: len(guilds)}
	for _, guild := range guilds {
		res.members += int(guild.ApproximateMembers)
	}
	return res
}

type appLazyFields struct {
	GuildCount  int   `expr:"guild_count" desc:"The number of servers the app is in"`
	MemberCount int   `expr:"member_count" desc:"The members of all servers of the app added up, a user in several of them counted once for each"`
	Latency     int64 `expr:"latency" desc:"The number of milliseconds Discord took to answer the last heartbeat of the app"`

	UserInstallCount int `expr:"user_install_count" desc:"Discord's approximate number of users that installed the app, updated about once a day. 0 until Discord first reports it"`
}

func (a AppEnv) lazyFields() any {
	return appLazyFields{}
}

// lazyField has the stats of the app. The server and member counts go over
// every server in the cache, so they are only counted when a flow uses them.
func (a AppEnv) lazyField(name string) (any, bool, error) {
	switch name {
	case "guild_count", "member_count":
		var guilds appGuilds
		if a.guilds != nil {
			guilds = a.guilds()
		}
		if name == "guild_count" {
			return guilds.count, true, nil
		}
		return guilds.members, true, nil
	case "latency":
		return a.latency().Milliseconds(), true, nil
	case "user_install_count":
		if a.userInstalls == nil {
			return 0, true, nil
		}
		count, err := a.userInstalls()
		return count, true, err
	}
	return nil, false, nil
}

// latency is 0 between Hello and Ready. From Ready until the first heartbeat
// ack, it's how long Discord took to send Ready.
func (a *AppEnv) latency() time.Duration {
	if a.session == nil || a.session.Session == nil {
		return 0
	}
	gw := a.session.Gateway()
	if gw == nil {
		return 0
	}
	return max(gw.Latency(), 0)
}

type RobloxUserEnv struct {
	og thing.RobloxUserValue

	ID                     int64  `expr:"id" json:"id"`
	Name                   string `expr:"name" json:"name"`
	DisplayName            string `expr:"display_name" json:"display_name"`
	Description            string `expr:"description" json:"description"`
	CreatedAt              string `expr:"created_at" json:"created_at" desc:"When the account was created, as an ISO 8601 date"`
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
