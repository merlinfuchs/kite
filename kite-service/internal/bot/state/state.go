package state

import (
	"context"
	"fmt"
	"sync"

	"github.com/disgoorg/disgo/rest"
	"github.com/merlinfuchs/dismod/disrest"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type BotState struct {
	sync.RWMutex
	client *disrest.Client

	botUser     distype.User
	application distype.Application
	guilds      map[distype.Snowflake]*distype.Guild
	channels    map[distype.Snowflake]*distype.Channel
	roles       map[distype.Snowflake]*distype.Role
	botMembers  map[distype.Snowflake]*distype.Member
	memberLocks sync.Map
}

func New(client *disrest.Client) *BotState {
	return &BotState{
		client:     client,
		guilds:     map[distype.Snowflake]*distype.Guild{},
		channels:   map[distype.Snowflake]*distype.Channel{},
		roles:      map[distype.Snowflake]*distype.Role{},
		botMembers: map[distype.Snowflake]*distype.Member{},
	}
}

func (s *BotState) Update(_ int, t distype.EventType, e any) {
	fmt.Printf("%T\n", e)

	switch t {
	case distype.EventTypeReady:
		s.Lock()
		e := e.(*distype.ReadyEvent)
		s.botUser = e.User
		s.application = e.Application
		s.Unlock()
	case distype.EventTypeGuildCreate:
		s.Lock()
		e := e.(*distype.GuildCreateEvent)
		s.guilds[e.ID] = e
		for _, member := range e.Members {
			s.botMembers[member.User.ID] = &member
		}
		e.Members = nil
		s.Unlock()
	}
}

func (s *BotState) GetGuildBotMember(ctx context.Context, guildID distype.Snowflake) (*distype.Member, error) {
	s.RLock()
	defer s.RUnlock()

	fmt.Println("get bot mmeber")

	member, exists := s.botMembers[guildID]
	if exists {
		return member, nil
	}
	return nil, store.ErrNotFound
}

func (s *BotState) GetGuildMember(ctx context.Context, guildID distype.Snowflake, userID distype.Snowflake) (*distype.Member, error) {
	lockKey := memberLockKey{
		guildID: guildID,
		userID:  userID,
	}

	lock, _ := s.memberLocks.LoadOrStore(lockKey, &sync.Mutex{})

	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()
	defer s.memberLocks.Delete(lockKey)

	var member *distype.Member
	err := s.client.Request(rest.GetMember.Compile(nil, guildID, userID), nil, &member, rest.WithCtx(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to request member: %w", err)
	}

	// TODO?: cache member temporarily

	return member, nil
}

func (s *BotState) GetGuildOwnerID(ctx context.Context, guildID distype.Snowflake) (distype.Snowflake, error) {
	guild, exists := s.guilds[guildID]
	if !exists {
		return "", store.ErrNotFound
	}

	return guild.OwnerID, nil
}

func (s *BotState) GetGuildRoles(ctx context.Context, guildID distype.Snowflake) ([]distype.Role, error) {
	guild, exists := s.guilds[guildID]
	if !exists {
		return nil, store.ErrNotFound
	}

	return guild.Roles, nil
}
