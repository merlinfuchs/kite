package eval

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/expr-lang/expr/ast"
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
	Channel    *SnowflakeEnv            `expr:"channel" json:"channel"`
	Guild      *SnowflakeEnv            `expr:"guild" json:"guild"`
	User       any                      `expr:"user" json:"user"`
	Member     any                      `expr:"member" json:"member"`
	Command    *CommandEnv              `expr:"command" json:"command"`
	Components map[string]*ComponentEnv `expr:"components" json:"components"`
}

func NewInteractionEnv(i *discord.InteractionEvent) *InteractionEnv {
	e := &InteractionEnv{
		interaction: i,

		ID:         i.ID.String(),
		Channel:    NewSnowflakeEnv(i.ChannelID),
		Components: NewComponentsEnv(i),
	}

	if i.Member != nil {
		e.Member = NewMemberEnv(*i.Member)
		e.User = e.Member
	} else {
		e.User = NewUserEnv(*i.User)
		e.Member = e.User
	}

	if i.GuildID != 0 {
		e.Guild = NewSnowflakeEnv(i.GuildID)
	}

	if i.Data.InteractionType() == discord.CommandInteractionType {
		e.Command = NewCommandEnv(i)
	}

	return e
}

func NewContextFromInteraction(i *discord.InteractionEvent, session *state.State) Context {
	interactionEnv := NewInteractionEnv(i)

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
	data, _ := i.Data.(*discord.CommandInteraction)

	args := make(map[string]any)
	for _, option := range data.Options {
		var value any
		_ = json.Unmarshal(option.Value, &value)

		switch option.Type {
		case discord.UserOptionType:
			userID, _ := strconv.ParseInt(value.(string), 10, 64)
			user := data.Resolved.Users[discord.UserID(userID)]

			if member, ok := data.Resolved.Members[discord.UserID(userID)]; ok {
				member.User = user
				args[option.Name] = NewMemberEnv(member)
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
			args[option.Name] = NewChannelEnv(channel)
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
		default:
			args[option.Name] = value
		}
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

	User    any           `expr:"user" json:"user"`
	Member  any           `expr:"member" json:"member"`
	Channel *SnowflakeEnv `expr:"channel" json:"channel"`
	Message *MessageEnv   `expr:"message" json:"message"`
	Guild   *SnowflakeEnv `expr:"guild" json:"guild"`
}

func NewEventEnv(event ws.Event) *EventEnv {
	env := &EventEnv{
		event: event,
	}

	switch e := event.(type) {
	case *gateway.MessageCreateEvent:
		if e.Member != nil {
			env.Member = NewMemberEnv(*e.Member)
			env.User = env.Member
		} else {
			env.User = NewUserEnv(e.Author)
			env.Member = env.User
		}
		env.Channel = NewSnowflakeEnv(e.ChannelID)
		if e.GuildID != 0 {
			env.Guild = NewSnowflakeEnv(e.GuildID)
		}
		env.Message = NewMessageEnv(e.Message)
	case *gateway.MessageUpdateEvent:
		if e.Member != nil {
			env.Member = NewMemberEnv(*e.Member)
			env.User = env.Member
		} else {
			env.User = NewUserEnv(e.Author)
			env.Member = env.User
		}
		env.Channel = NewSnowflakeEnv(e.ChannelID)
		if e.GuildID != 0 {
			env.Guild = NewSnowflakeEnv(e.GuildID)
		}
		env.Message = NewMessageEnv(e.Message)
	case *gateway.MessageDeleteEvent:
		env.Message = NewMessageEnv(discord.Message{
			ID: e.ID,
		})
		env.Channel = NewSnowflakeEnv(e.ChannelID)
		if e.GuildID != 0 {
			env.Guild = NewSnowflakeEnv(e.GuildID)
		}
	case *gateway.GuildMemberAddEvent:
		env.Member = NewMemberEnv(e.Member)
		env.User = env.Member
		env.Guild = NewSnowflakeEnv(e.GuildID)
	case *gateway.GuildMemberRemoveEvent:
		env.User = NewUserEnv(e.User)
		env.Member = env.User
		env.Guild = NewSnowflakeEnv(e.GuildID)
	}

	return env
}

func NewContext(env Env) Context {
	return Context{
		Env: env,
	}
}

func NewContextFromEvent(event ws.Event, session *state.State) Context {
	return Context{
		Env: Env{
			"event":   NewEventEnv(event),
			"user":    NewEventEnv(event).User,
			"member":  NewEventEnv(event).Member,
			"channel": NewEventEnv(event).Channel,
			"guild":   NewEventEnv(event).Guild,
			"server":  NewEventEnv(event).Guild,
			"message": NewEventEnv(event).Message,
			"app":     NewAppEnv(session),
		},
	}
}

type UserEnv struct {
	og discord.User

	ID            string `expr:"id" json:"id"`
	Username      string `expr:"username" json:"username"`
	Discriminator string `expr:"discriminator" json:"discriminator"`
	DisplayName   string `expr:"display_name" json:"display_name"`
	Mention       string `expr:"mention" json:"mention"`
	AvatarURL     string `expr:"avatar_url" json:"avatar_url"`
	BannerURL     string `expr:"banner_url" json:"banner_url"`
}

func (u UserEnv) Thing() thing.Thing {
	return thing.NewDiscordUser(u.og)
}

func (u UserEnv) String() string {
	return u.ID
}

func NewUserEnv(user discord.User) *UserEnv {
	return &UserEnv{
		og: user,

		ID:            user.ID.String(),
		Username:      user.Username,
		Discriminator: user.Discriminator,
		DisplayName:   user.DisplayName,

		Mention:   fmt.Sprintf("<@%s>", user.ID.String()),
		AvatarURL: user.AvatarURL(),
		BannerURL: user.BannerURL(),
	}
}

type MemberEnv struct {
	og discord.Member

	UserEnv

	Nick    string   `expr:"nick" json:"nick"`
	RoleIDs []string `expr:"role_ids" json:"role_ids"`
}

func (m MemberEnv) String() string {
	return m.UserEnv.String()
}

func NewMemberEnv(member discord.Member) *MemberEnv {
	roleIDs := make([]string, len(member.RoleIDs))
	for i, role := range member.RoleIDs {
		roleIDs[i] = role.String()
	}

	return &MemberEnv{
		og: member,

		UserEnv: *NewUserEnv(member.User),

		Nick:    member.Nick,
		RoleIDs: roleIDs,
	}
}

func (m MemberEnv) Thing() thing.Thing {
	return thing.NewDiscordMember(m.og)
}

type ChannelEnv struct {
	og discord.Channel

	ID      string `expr:"id" json:"id"`
	Name    string `expr:"name" json:"name"`
	Mention string `expr:"mention" json:"mention"`
}

func NewChannelEnv(channel discord.Channel) *ChannelEnv {
	return &ChannelEnv{
		og: channel,

		ID:      channel.ID.String(),
		Name:    channel.Name,
		Mention: fmt.Sprintf("<#%s>", channel.ID.String()),
	}
}

func (c ChannelEnv) Thing() thing.Thing {
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

	ID   string `expr:"id" json:"id"`
	Name string `expr:"name" json:"name"`
}

func NewGuildEnv(guild discord.Guild) *GuildEnv {
	return &GuildEnv{
		og: guild,

		ID:   guild.ID.String(),
		Name: guild.Name,
	}
}

func (g GuildEnv) Thing() thing.Thing {
	return thing.NewDiscordGuild(g.og)
}

func (g GuildEnv) String() string {
	return g.ID
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
}

func NewHTTPResponseEnv(resp thing.HTTPResponseValue) *HTTPResponseEnv {
	res := &HTTPResponseEnv{
		og: resp,

		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		BodyFunc: func() (string, error) {
			return string(resp.Body), nil
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
