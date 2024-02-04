package state

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-types/dismodel"
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

func (s *BotState) GetGuildBotMember(ctx context.Context, guildID string) (*dismodel.Member, error) {
	member, err := s.state.Member(guildID, s.state.User.ID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return modelMember(member), nil
}

func (s *BotState) GetGuildMember(ctx context.Context, guildID string, userID string) (*dismodel.Member, error) {
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

func (s *BotState) GetGuildRoles(ctx context.Context, guildID string) ([]*dismodel.Role, error) {
	guild, err := s.state.Guild(guildID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	res := make([]*dismodel.Role, len(guild.Roles))
	for i, role := range guild.Roles {
		res[i] = modelRole(role)
	}

	return res, nil
}

func modelMember(member *discordgo.Member) *dismodel.Member {
	return &dismodel.Member{
		User:     modelUser(member.User),
		Nick:     member.Nick,
		Avatar:   member.Avatar,
		Roles:    member.Roles,
		JoinedAt: member.JoinedAt,
		Deaf:     member.Deaf,
		Mute:     member.Mute,
	}
}

func modelUser(user *discordgo.User) *dismodel.User {
	return &dismodel.User{
		ID:            user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		GlobalName:    user.Username, // TODO
		Avatar:        user.Avatar,
		Banner:        user.Banner,
		AccentColor:   user.AccentColor,
		Bot:           user.Bot,
		System:        user.System,
		PublicFlags:   int(user.PublicFlags),
	}
}

func modelRole(role *discordgo.Role) *dismodel.Role {
	return &dismodel.Role{
		ID:          role.ID,
		Name:        role.Name,
		Color:       role.Color,
		Hoist:       role.Hoist,
		Position:    role.Position,
		Permissions: fmt.Sprintf("%d", role.Permissions),
		Managed:     role.Managed,
		Mentionable: role.Mentionable,
	}
}
