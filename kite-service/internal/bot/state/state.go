package state

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/disgoorg/disgo/rest"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/merlinfuchs/dismod/disrest"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type memberKey struct {
	guildID distype.Snowflake
	userID  distype.Snowflake
}

type BotState struct {
	sync.RWMutex
	client *disrest.Client

	botUser     distype.User
	application distype.Application
	guilds      map[distype.Snowflake]*distype.Guild
	channels    map[distype.Snowflake]*distype.Channel
	roles       map[distype.Snowflake]*distype.Role
	botMembers  map[distype.Snowflake]*distype.Member
	members     *expirable.LRU[memberKey, *distype.Member]
	memberLocks sync.Map
}

func New(client *disrest.Client) *BotState {
	return &BotState{
		client:     client,
		guilds:     map[distype.Snowflake]*distype.Guild{},
		channels:   map[distype.Snowflake]*distype.Channel{},
		roles:      map[distype.Snowflake]*distype.Role{},
		botMembers: map[distype.Snowflake]*distype.Member{},
		members:    expirable.NewLRU[memberKey, *distype.Member](1000, nil, time.Minute),
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
			if member.User.ID == s.botUser.ID {
				s.botMembers[e.ID] = &member
			}
		}
		e.Members = nil
		s.Unlock()
	}
}

func (s *BotState) GetGuildBotMember(ctx context.Context, guildID distype.Snowflake) (*distype.Member, error) {
	s.RLock()
	defer s.RUnlock()

	member, exists := s.botMembers[guildID]
	if exists {
		return member, nil
	}
	return nil, store.ErrNotFound
}

func (s *BotState) GetGuildMember(ctx context.Context, guildID distype.Snowflake, userID distype.Snowflake) (*distype.Member, error) {
	key := memberKey{
		guildID: guildID,
		userID:  userID,
	}

	lock, _ := s.memberLocks.LoadOrStore(key, &sync.Mutex{})

	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()
	defer s.memberLocks.Delete(key)

	if member, exists := s.members.Get(key); exists {
		return member, nil
	}

	var member *distype.Member
	err := s.client.Request(rest.GetMember.Compile(nil, guildID, userID), nil, &member, rest.WithCtx(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to request member: %w", err)
	}

	s.members.Add(key, member)

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
