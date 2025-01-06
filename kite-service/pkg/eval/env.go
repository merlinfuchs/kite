package eval

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/ws"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type Env map[string]any

type InteractionEnv struct {
	interaction *discord.InteractionEvent

	ID      string      `expr:"id" json:"id"`
	Channel *ChannelEnv `expr:"channel" json:"channel"`
	Guild   *GuildEnv   `expr:"guild" json:"guild"`
	User    any         `expr:"user" json:"user"`
	Command *CommandEnv `expr:"command" json:"command"`
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

	ID   string         `expr:"id" json:"id"`
	Args map[string]any `expr:"args" json:"args"`
}

func NewCommandEnv(i *discord.InteractionEvent) *CommandEnv {
	data, _ := i.Data.(*discord.CommandInteraction)

	args := make(map[string]any)
	for _, option := range data.Options {
		var value any
		_ = json.Unmarshal(option.Value, &value)
		args[option.Name] = value
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

type EventEnv struct {
	event ws.Event

	User    any         `expr:"user" json:"user"`
	Member  any         `expr:"member" json:"member"`
	Channel *ChannelEnv `expr:"channel" json:"channel"`
	Message *MessageEnv `expr:"message" json:"message"`
	Guild   *GuildEnv   `expr:"guild" json:"guild"`
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
	ID            string `expr:"id" json:"id"`
	Username      string `expr:"username" json:"username"`
	Discriminator string `expr:"discriminator" json:"discriminator"`
	DisplayName   string `expr:"display_name" json:"display_name"`
	Mention       string `expr:"mention" json:"mention"`
	AvatarURL     string `expr:"avatar_url" json:"avatar_url"`
	BannerURL     string `expr:"banner_url" json:"banner_url"`
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

	Nick    string   `expr:"nick" json:"nick"`
	RoleIDs []string `expr:"role_ids" json:"role_ids"`
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
	ID string `expr:"id" json:"id"`
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
	ID      string `expr:"id" json:"id"`
	Content string `expr:"content" json:"content"`
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
	ID string `expr:"id" json:"id"`
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
	resp *http.Response

	Status     string                 `expr:"status" json:"status"`
	StatusCode int                    `expr:"status_code" json:"status_code"`
	BodyFunc   func() (string, error) `expr:"body" json:"-"`
}

func NewHTTPResponseEnv(resp *http.Response) *HTTPResponseEnv {
	return &HTTPResponseEnv{
		resp: resp,

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

func NewAnyEnv(v any) any {
	switch v := v.(type) {
	case thing.Any:
		return NewAnyEnv(v.Inner)
	case *discord.Message:
		return NewMessageEnv(v)
	case *http.Response:
		return NewHTTPResponseEnv(v)
	case *discord.User:
		return NewUserEnv(v)
	case *discord.Member:
		return NewMemberEnv(v)
	case *discord.Channel:
		return NewChannelEnv(v.ID)
	case *discord.Guild:
		return NewGuildEnv(v.ID)
	case *discord.InteractionEvent:
		return NewInteractionEnv(v)
	default:
		return v
	}
}
