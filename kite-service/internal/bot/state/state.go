package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type BotState struct {
	state       *discordgo.State
	session     *discordgo.Session
	memberLocks sync.Map
}

func New(state *discordgo.State, session *discordgo.Session) *BotState {
	return &BotState{
		state:   state,
		session: session,
	}
}

func (s *BotState) GetGuildBotMember(ctx context.Context, guildID string) (*distype.Member, error) {
	member, err := s.state.Member(guildID, s.state.User.ID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return modelMember(member), nil
}

func (s *BotState) GetGuildMember(ctx context.Context, guildID string, userID string) (*distype.Member, error) {
	lockKey := memberLockKey{
		guildID: guildID,
		userID:  userID,
	}

	lock, _ := s.memberLocks.LoadOrStore(lockKey, &sync.Mutex{})

	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()
	defer s.memberLocks.Delete(lockKey)

	member, err := s.state.Member(guildID, userID)
	if err == nil {
		return modelMember(member), nil
	} else if err != discordgo.ErrStateNotFound {
		return nil, err
	}

	member, err = s.session.GuildMember(guildID, userID, discordgo.WithContext(ctx))
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			var derr *discordgo.RESTError
			if errors.As(err, &derr) {
				if derr.Message != nil && derr.Message.Code == discordgo.ErrCodeUnknownMember {
					return nil, store.ErrNotFound
				}
			}
		}

		return nil, err
	}

	err = s.state.MemberAdd(member)
	if err != nil {
		slog.With(logattr.Error(err)).Error("failed to add member to state")
	}

	return modelMember(member), nil
}

func (s *BotState) GetGuildOwnerID(ctx context.Context, guildID string) (string, error) {
	guild, err := s.state.Guild(guildID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return "", store.ErrNotFound
		}
		return "", err
	}

	return guild.OwnerID, nil
}

func (s *BotState) GetGuildRoles(ctx context.Context, guildID string) ([]*distype.Role, error) {
	guild, err := s.state.Guild(guildID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	res := make([]*distype.Role, len(guild.Roles))
	for i, role := range guild.Roles {
		res[i] = modelRole(role)
	}

	return res, nil
}

func modelMember(member *discordgo.Member) *distype.Member {
	roles := make([]distype.Snowflake, len(member.Roles))
	for i, role := range member.Roles {
		roles[i] = distype.Snowflake(role)
	}

	return &distype.Member{
		User:     modelUser(member.User),
		Nick:     distype.OptionalNullString(member.Nick, member.Nick != ""),
		Avatar:   distype.OptionalNullString(member.Avatar, member.Avatar != ""),
		Roles:    roles,
		JoinedAt: member.JoinedAt,
		Deaf:     member.Deaf,
		Mute:     member.Mute,
	}
}

func modelUser(user *discordgo.User) *distype.User {
	var publicFlags distype.Optional[distype.UserFlags]
	if user.PublicFlags != 0 {
		v := distype.UserFlags(user.PublicFlags)
		publicFlags = &v
	}

	return &distype.User{
		ID:            distype.Snowflake(user.ID),
		Username:      user.Username,
		Discriminator: user.Discriminator,
		GlobalName:    distype.NullString("", false), // TODO
		Avatar:        distype.NullString(user.Avatar, user.Avatar != ""),
		Banner:        distype.OptionalNullString(user.Banner, user.Banner != ""),
		AccentColor:   &distype.Nullable[int]{Value: int(user.AccentColor), Valid: user.AccentColor != 0},
		Bot:           &user.Bot,
		System:        &user.System,
		PublicFlags:   publicFlags,
	}
}

func modelRole(role *discordgo.Role) *distype.Role {
	return &distype.Role{
		ID:          distype.Snowflake(role.ID),
		Name:        role.Name,
		Color:       role.Color,
		Hoist:       role.Hoist,
		Position:    role.Position,
		Permissions: distype.Permissions(fmt.Sprintf("%d", role.Permissions)),
		Managed:     role.Managed,
		Mentionable: role.Mentionable,
	}
}
