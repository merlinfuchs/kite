package state

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/bwmarrin/discordgo"
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

func (s *BotState) GetGuildBotMember(ctx context.Context, guildID string) (*discordgo.Member, error) {
	member, err := s.state.Member(guildID, s.state.User.ID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return nil, store.ErrNotFound
		}
		return nil, err
	}
	return member, nil
}

func (s *BotState) GetGuildMember(ctx context.Context, guildID string, userID string) (*discordgo.Member, error) {
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
		return member, nil
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

	return member, nil
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

func (s *BotState) GetGuildRoles(ctx context.Context, guildID string) ([]*discordgo.Role, error) {
	guild, err := s.state.Guild(guildID)
	if err != nil {
		if err == discordgo.ErrStateNotFound {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return guild.Roles, nil
}
