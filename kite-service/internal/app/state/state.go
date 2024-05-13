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

var _ store.DiscordStateStore = (*AppState)(nil)

type memberKey struct {
	guildID distype.Snowflake
	userID  distype.Snowflake
}

type AppState struct {
	sync.RWMutex
	client *disrest.Client

	botUser       distype.User
	application   distype.Application
	guilds        map[distype.Snowflake]distype.Guild
	guildRoles    map[distype.Snowflake][]distype.Snowflake
	guildChannels map[distype.Snowflake][]distype.Snowflake
	channels      map[distype.Snowflake]distype.Channel
	roles         map[distype.Snowflake]distype.Role
	botMembers    map[distype.Snowflake]distype.Member
	members       *expirable.LRU[memberKey, distype.Member]
	memberLocks   sync.Map
}

func New(client *disrest.Client) *AppState {
	return &AppState{
		client:        client,
		guilds:        map[distype.Snowflake]distype.Guild{},
		guildRoles:    map[distype.Snowflake][]distype.Snowflake{},
		guildChannels: map[distype.Snowflake][]distype.Snowflake{},
		channels:      map[distype.Snowflake]distype.Channel{},
		roles:         map[distype.Snowflake]distype.Role{},
		botMembers:    map[distype.Snowflake]distype.Member{},
		members:       expirable.NewLRU[memberKey, distype.Member](1000, nil, time.Minute),
	}
}

func (s *AppState) Update(_ int, t distype.EventType, e any) {
	switch t {
	case distype.EventTypeReady:
		s.Lock()
		e := e.(*distype.ReadyEvent)
		s.botUser = e.User
		s.application = e.Application
		s.Unlock()
	case distype.EventTypeGuildCreate:
		s.Lock()
		e := *e.(*distype.GuildCreateEvent)

		guildRoles := make([]distype.Snowflake, len(e.Roles))
		guildChannels := make([]distype.Snowflake, len(e.Channels))

		for i, channel := range e.Channels {
			channel.GuildID = &e.ID
			s.channels[channel.ID] = channel
			guildChannels[i] = channel.ID
		}
		for i, role := range e.Roles {
			s.roles[role.ID] = role
			guildRoles[i] = role.ID
		}
		s.guildRoles[e.ID] = guildRoles
		s.guildChannels[e.ID] = guildChannels

		for _, member := range e.Members {
			if member.User.ID == s.botUser.ID {
				s.botMembers[e.ID] = member
			}
		}

		e.Roles = nil
		e.Channels = nil
		e.Members = nil
		s.guilds[e.ID] = e
		s.Unlock()
	case distype.EventTypeGuildUpdate:
		s.Lock()
		e := e.(*distype.GuildUpdateEvent)
		// TODO: retain info from previous state
		s.guilds[e.ID] = *e
		s.Unlock()
	case distype.EventTypeGuildDelete:
		s.Lock()
		e := e.(*distype.GuildDeleteEvent)
		if !e.Unavailable {
			roles := s.guildRoles[e.ID]
			for _, roleID := range roles {
				delete(s.roles, roleID)
			}
			channels := s.guildChannels[e.ID]
			for _, channelID := range channels {
				delete(s.channels, channelID)
			}

			delete(s.guilds, e.ID)
			delete(s.botMembers, e.ID)
			delete(s.guildRoles, e.ID)
			delete(s.guildChannels, e.ID)
		}
		s.Unlock()
	case distype.EventTypeChannelCreate:
		s.Lock()
		e := e.(*distype.ChannelCreateEvent)
		s.channels[e.ID] = *e
		if e.GuildID != nil {
			s.guildChannels[*e.GuildID] = append(s.guildChannels[*e.GuildID], e.ID)
		}
		s.Unlock()
	case distype.EventTypeChannelUpdate:
		s.Lock()
		e := e.(*distype.ChannelUpdateEvent)
		s.channels[e.ID] = *e
		s.Unlock()
	case distype.EventTypeChannelDelete:
		s.Lock()
		e := e.(*distype.ChannelDeleteEvent)
		if e.GuildID != nil {
			channels := s.guildChannels[*e.GuildID]
			for i, channelID := range channels {
				if channelID == e.ID {
					s.guildChannels[*e.GuildID] = append(channels[:i], channels[i+1:]...)
					break
				}
			}
		}
		delete(s.channels, e.ID)
		s.Unlock()
	case distype.EventTypeGuildRoleCreate:
		s.Lock()
		e := e.(*distype.RoleCreateEvent)
		s.guildRoles[e.GuildID] = append(s.guildRoles[e.GuildID], e.Role.ID)
		s.Unlock()
	case distype.EventTypeGuildRoleUpdate:
		s.Lock()
		e := e.(*distype.RoleUpdateEvent)
		s.roles[e.Role.ID] = e.Role
		s.Unlock()
	case distype.EventTypeGuildRoleDelete:
		s.Lock()
		e := e.(*distype.RoleDeleteEvent)
		roles := s.guildRoles[e.GuildID]
		for i, roleID := range roles {
			if roleID == e.RoleID {
				s.guildRoles[e.GuildID] = append(roles[:i], roles[i+1:]...)
				break
			}
		}
		delete(s.roles, e.RoleID)
		s.Unlock()
	case distype.EventTypeGuildMemberUpdate:
		s.Lock()
		e := e.(*distype.MemberUpdateEvent)
		if e.User.ID == s.botUser.ID {
			s.botMembers[e.GuildID] = e.Member
		}
		s.Unlock()
	case distype.EventTypeGuildMemberRemove:
		s.Lock()
		e := e.(*distype.MemberRemoveEvent)
		if e.User.ID == s.botUser.ID {
			delete(s.botMembers, e.GuildID)
		}
		s.Unlock()
	}
}

func (s *AppState) GetGuild(ctx context.Context, guildID distype.Snowflake) (*distype.Guild, error) {
	s.RLock()
	defer s.RUnlock()

	guild, exists := s.guilds[guildID]
	if exists {
		return &guild, nil
	}
	return nil, store.ErrNotFound
}

func (s *AppState) GetGuildChannels(ctx context.Context, guildID distype.Snowflake) ([]distype.Channel, error) {
	s.RLock()
	defer s.RUnlock()

	channels, exists := s.guildChannels[guildID]
	if exists {
		res := make([]distype.Channel, len(channels))
		for i, channelID := range channels {
			res[i] = s.channels[channelID]
		}
		return res, nil
	}
	return nil, store.ErrNotFound
}

func (s *AppState) GetRole(ctx context.Context, guildID distype.Snowflake, roleID distype.Snowflake) (*distype.Role, error) {
	s.RLock()
	defer s.RUnlock()

	role, exists := s.roles[roleID]
	if exists {
		return &role, nil
	}
	return nil, store.ErrNotFound
}

func (s *AppState) GetChannel(ctx context.Context, channelID distype.Snowflake) (*distype.Channel, error) {
	s.RLock()
	defer s.RUnlock()

	channel, exists := s.channels[channelID]
	if exists {
		return &channel, nil
	}
	return nil, store.ErrNotFound

}

func (s *AppState) GetGuildBotMember(ctx context.Context, guildID distype.Snowflake) (*distype.Member, error) {
	s.RLock()
	defer s.RUnlock()

	member, exists := s.botMembers[guildID]
	if exists {
		return &member, nil
	}
	return nil, store.ErrNotFound
}

func (s *AppState) GetGuildMember(ctx context.Context, guildID distype.Snowflake, userID distype.Snowflake) (*distype.Member, error) {
	key := memberKey{
		guildID: guildID,
		userID:  userID,
	}

	lock, _ := s.memberLocks.LoadOrStore(key, &sync.Mutex{})

	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()
	defer s.memberLocks.Delete(key)

	if member, exists := s.members.Get(key); exists {
		return &member, nil
	}

	var member distype.Member
	err := s.client.Request(rest.GetMember.Compile(nil, guildID, userID), nil, &member, rest.WithCtx(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to request member: %w", err)
	}

	s.members.Add(key, member)

	return &member, nil
}

func (s *AppState) GetGuildOwnerID(ctx context.Context, guildID distype.Snowflake) (distype.Snowflake, error) {
	guild, exists := s.guilds[guildID]
	if !exists {
		return "", store.ErrNotFound
	}

	return guild.OwnerID, nil
}

func (s *AppState) GetGuildRoles(ctx context.Context, guildID distype.Snowflake) ([]distype.Role, error) {
	s.RLock()
	defer s.RUnlock()

	roleIDs, exists := s.guildRoles[guildID]
	if exists {
		res := make([]distype.Role, len(roleIDs))
		for i, roleID := range roleIDs {
			res[i] = s.roles[roleID]
		}
		return res, nil
	}
	return nil, store.ErrNotFound
}
