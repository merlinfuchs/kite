package appstate

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

type AppStateHandler struct {
	appStateManager store.AppStateManager
}

func NewAppStateHandler(appStateManager store.AppStateManager) *AppStateHandler {
	return &AppStateHandler{appStateManager: appStateManager}
}

func (h *AppStateHandler) HandleStateStatusGet(c *handler.Context) (*wire.StateStatusGetResponse, error) {
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// App state not found, app is not running
			return &wire.AppStateStatus{}, nil
		}

		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	status, err := state.AppStatus(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get app status: %w", err)
	}

	return &wire.AppStateStatus{
		Online: status.Online,
	}, nil
}

func (h *AppStateHandler) HandleStateGuildList(c *handler.Context) (*wire.StateGuildListResponse, error) {
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	guilds, err := state.AppGuilds(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get guilds: %w", err)
	}

	// Member counts aren't cached and cost a request to Discord, so they are
	// only loaded when asked for.
	memberCounts := map[discord.GuildID]uint64{}
	if c.Query("with_counts") == "true" {
		client, err := h.appStateManager.AppClient(c.Context(), c.App.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get app state: %w", err)
		}

		counted, err := client.WithContext(c.Context()).GuildsWithCounts(0)
		if err == nil {
			for _, guild := range counted {
				memberCounts[guild.ID] = guild.ApproximateMembers
			}
		}
	}

	res := make([]*wire.Guild, len(guilds))
	for i, guild := range guilds {
		res[i] = wire.GuildToWire(&guild)

		member, err := state.AppGuildMember(c.Context(), guild.ID.String())
		if err == nil && member.Joined.IsValid() {
			res[i].JoinedAt = null.TimeFrom(member.Joined.Time())
		}

		if count, ok := memberCounts[guild.ID]; ok {
			res[i].MemberCount = null.IntFrom(int64(count))
		}
	}

	return &res, nil
}

func (h *AppStateHandler) HandleStateGuildGet(c *handler.Context) (*wire.StateGuildGetResponse, error) {
	guildID := c.Param("guildID")

	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	guilds, err := state.AppGuilds(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get guilds: %w", err)
	}

	guildIndex := slices.IndexFunc(guilds, func(guild discord.Guild) bool {
		return guild.ID.String() == guildID
	})
	if guildIndex == -1 {
		return nil, handler.ErrNotFound("unknown_guild", "The app isn't in this server")
	}
	guild := guilds[guildIndex]

	roles, err := state.AppGuildRoles(c.Context(), guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild roles: %w", err)
	}

	client, err := h.appStateManager.AppClient(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}
	client = client.WithContext(c.Context())

	member, err := state.AppGuildMember(c.Context(), guildID)
	if err != nil {
		// The app's member isn't cached, so it's fetched from Discord.
		me, err := client.Me()
		if err != nil {
			return nil, fmt.Errorf("failed to get app user: %w", err)
		}

		member, err = client.Member(guild.ID, me.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get app member: %w", err)
		}
	}

	var permissions discord.Permissions
	for _, role := range roles {
		// The @everyone role has the ID of the guild and applies to everyone.
		if role.ID == discord.RoleID(guild.ID) || slices.Contains(member.RoleIDs, role.ID) {
			permissions |= role.Permissions
		}
	}

	res := &wire.GuildDetails{
		Owner: wire.GuildOwner{
			ID: guild.OwnerID.String(),
		},
		Permissions: strconv.FormatUint(uint64(permissions), 10),
	}

	// The owner is only known by ID, the rest is best effort.
	owner, err := client.User(guild.OwnerID)
	if err == nil {
		avatarURL := owner.AvatarURL()

		res.Owner.Username = owner.Username
		res.Owner.DisplayName = owner.DisplayName
		res.Owner.AvatarURL = null.NewString(avatarURL, avatarURL != "")
	}

	return res, nil
}

func (h *AppStateHandler) HandleStateGuildChannelList(c *handler.Context) (*wire.StateGuildChannelListResponse, error) {
	guildID := c.Param("guildID")

	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	channels, err := state.AppGuildChannels(c.Context(), guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild channels: %w", err)
	}

	res := make([]*wire.Channel, len(channels))
	for i, channel := range channels {
		res[i] = wire.ChannelToWire(&channel)
	}

	return &res, nil
}

func (h *AppStateHandler) HandleStateGuildRoleList(c *handler.Context) (*wire.StateGuildRoleListResponse, error) {
	guildID := c.Param("guildID")

	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	roles, err := state.AppGuildRoles(c.Context(), guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild roles: %w", err)
	}

	res := make([]*wire.Role, len(roles))
	for i, role := range roles {
		res[i] = wire.RoleToWire(&role)
	}

	return &res, nil
}

func (h *AppStateHandler) HandleStateGuildLeave(c *handler.Context) (*wire.StateGuildLeaveResponse, error) {
	guildID, err := strconv.ParseInt(c.Param("guildID"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid guild ID: %w", err)
	}

	client, err := h.appStateManager.AppClient(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	err = client.WithContext(c.Context()).LeaveGuild(discord.GuildID(guildID))
	if err != nil {
		return nil, fmt.Errorf("failed to leave guild: %w", err)
	}

	return &wire.StateGuildLeaveResponse{}, nil
}
