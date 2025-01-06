package eval

import (
	"fmt"
	"io"
	"net/http"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
)

type Env map[string]any

type InteractionEnv struct {
	interaction *discord.InteractionEvent

	ID      string      `expr:"id"`
	Channel *ChannelEnv `expr:"channel"`
	Guild   *GuildEnv   `expr:"guild"`
	User    any         `expr:"user"`
	Command *CommandEnv `expr:"command"`
}

func NewInteractionEnv(i *discord.InteractionEvent) *InteractionEnv {
	e := &InteractionEnv{
		interaction: i,

		ID:      i.ID.String(),
		Channel: NewChannelEnv(i.ChannelID),
	}

	if i.User != nil {
		e.User = NewUserEnv(i.User)
	} else {
		e.User = NewMemberEnv(i.Member)
	}

	if i.GuildID != 0 {
		e.Guild = NewGuildEnv(i.GuildID)
	}

	if i.Data.InteractionType() == discord.CommandInteractionType {
		e.Command = NewCommandEnv(i)
	}

	return e
}

func NewEnvWithInteraction(i *discord.InteractionEvent) Env {
	interactionEnv := NewInteractionEnv(i)
	env := Env{
		"interaction": interactionEnv,
	}

	return env
}

func (e InteractionEnv) String() string {
	return e.interaction.ID.String()
}

type CommandEnv struct {
	interaction *discord.InteractionEvent
	cmd         *discord.CommandInteraction

	ID   string                       `expr:"id"`
	Args map[string]*CommandOptionEnv `expr:"args"`
}

func NewCommandEnv(i *discord.InteractionEvent) *CommandEnv {
	data, _ := i.Data.(*discord.CommandInteraction)

	args := make(map[string]*CommandOptionEnv)
	for _, option := range data.Options {
		args[option.Name] = NewCommandOptionEnv(i, data, &option)
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

type CommandOptionEnv struct {
	interaction *discord.InteractionEvent
	cmd         *discord.CommandInteraction
	option      *discord.CommandInteractionOption
}

func NewCommandOptionEnv(
	i *discord.InteractionEvent,
	cmd *discord.CommandInteraction,
	option *discord.CommandInteractionOption,
) *CommandOptionEnv {
	return &CommandOptionEnv{
		interaction: i,
		cmd:         cmd,
		option:      option,
	}
}

func (o CommandOptionEnv) String() string {
	return o.option.String()
}

type EventEnv struct {
	event ws.Event

	User    any         `expr:"user"`
	Member  any         `expr:"member"`
	Channel *ChannelEnv `expr:"channel"`
	Message *MessageEnv `expr:"message"`
	Guild   *GuildEnv   `expr:"guild"`
}

func NewEventEnv(event ws.Event) *EventEnv {
	env := &EventEnv{
		event: event,
	}

	switch e := event.(type) {
	case *gateway.MessageCreateEvent:
		if e.Member != nil {
			env.Member = NewMemberEnv(e.Member)
			env.User = env.Member
		} else {
			env.User = NewUserEnv(&e.Author)
			env.Member = env.User
		}
		env.Channel = NewChannelEnv(e.ChannelID)
		if e.GuildID != 0 {
			env.Guild = NewGuildEnv(e.GuildID)
		}
		env.Message = NewMessageEnv(&e.Message)
	case *gateway.MessageUpdateEvent:
		if e.Member != nil {
			env.Member = NewMemberEnv(e.Member)
			env.User = env.Member
		} else {
			env.User = NewUserEnv(&e.Author)
			env.Member = env.User
		}
		env.Channel = NewChannelEnv(e.ChannelID)
		if e.GuildID != 0 {
			env.Guild = NewGuildEnv(e.GuildID)
		}
		env.Message = NewMessageEnv(&e.Message)
	case *gateway.MessageDeleteEvent:
		env.Message = NewMessageEnv(&discord.Message{
			ID: e.ID,
		})
		env.Channel = NewChannelEnv(e.ChannelID)
		if e.GuildID != 0 {
			env.Guild = NewGuildEnv(e.GuildID)
		}
	case *gateway.GuildMemberAddEvent:
		env.Member = NewMemberEnv(&e.Member)
		env.User = env.Member
		env.Guild = NewGuildEnv(e.GuildID)
	case *gateway.GuildMemberRemoveEvent:
		env.User = NewUserEnv(&e.User)
		env.Member = env.User
		env.Guild = NewGuildEnv(e.GuildID)
	}

	return env
}

func NewEnvWithEvent(event ws.Event) Env {
	return Env{
		"event": NewEventEnv(event),
	}
}

type UserEnv struct {
	ID            string `expr:"id"`
	Username      string `expr:"username"`
	Discriminator string `expr:"discriminator"`
	DisplayName   string `expr:"display_name"`
	Mention       string `expr:"mention"`
	AvatarURL     string `expr:"avatar_url"`
	BannerURL     string `expr:"banner_url"`
}

func (u UserEnv) String() string {
	return u.ID
}

func NewUserEnv(user *discord.User) *UserEnv {
	return &UserEnv{
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
	UserEnv

	Nick    string   `expr:"nick"`
	RoleIDs []string `expr:"role_ids"`
}

func (m MemberEnv) String() string {
	return m.UserEnv.String()
}

func NewMemberEnv(member *discord.Member) *MemberEnv {
	roleIDs := make([]string, len(member.RoleIDs))
	for i, role := range member.RoleIDs {
		roleIDs[i] = role.String()
	}

	return &MemberEnv{
		UserEnv: *NewUserEnv(&member.User),

		Nick:    member.Nick,
		RoleIDs: roleIDs,
	}
}

type ChannelEnv struct {
	ID string `expr:"id"`
}

func NewChannelEnv(channelID discord.ChannelID) *ChannelEnv {
	return &ChannelEnv{
		ID: channelID.String(),
	}
}

func (c ChannelEnv) String() string {
	return c.ID
}

type MessageEnv struct {
	ID      string `expr:"id"`
	Content string `expr:"content"`
}

func NewMessageEnv(msg *discord.Message) *MessageEnv {
	return &MessageEnv{
		ID:      msg.ID.String(),
		Content: msg.Content,
	}
}

func (m MessageEnv) String() string {
	return m.ID
}

type GuildEnv struct {
	ID string `expr:"id"`
}

func NewGuildEnv(guildID discord.GuildID) *GuildEnv {
	return &GuildEnv{
		ID: guildID.String(),
	}
}

func (g GuildEnv) String() string {
	return g.ID
}

type HTTPResponseEnv struct {
	Status     string                 `expr:"status"`
	StatusCode int                    `expr:"status_code"`
	BodyFunc   func() (string, error) `expr:"body"`
}

func NewHTTPResponseEnv(resp *http.Response) *HTTPResponseEnv {
	return &HTTPResponseEnv{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		BodyFunc: func() (string, error) {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			return string(body), nil
		},
	}
}

func (h HTTPResponseEnv) String() string {
	return h.Status
}
