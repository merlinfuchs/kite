package appstate

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
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

	res := make([]*wire.Guild, len(guilds))
	for i, guild := range guilds {
		res[i] = wire.GuildToWire(&guild)
	}

	return &res, nil
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

// maxInviteChannelAttempts limits how many channels are tried when creating
// an invite, so a server where the app can't invite anywhere fails quickly.
const maxInviteChannelAttempts = 10

func (h *AppStateHandler) HandleStateGuildInviteCreate(c *handler.Context) (*wire.StateGuildInviteCreateResponse, error) {
	guildID := c.Param("guildID")

	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	channels, err := state.AppGuildChannels(c.Context(), guildID)
	if err != nil {
		return nil, fmt.Errorf("failed to get guild channels: %w", err)
	}

	// Invites to text channels land the user somewhere readable, so only
	// those are tried, from the top of the channel list down.
	candidates := make([]discord.Channel, 0, len(channels))
	for _, channel := range channels {
		if channel.Type == discord.GuildText || channel.Type == discord.GuildAnnouncement {
			candidates = append(candidates, channel)
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Position < candidates[j].Position
	})
	if len(candidates) > maxInviteChannelAttempts {
		candidates = candidates[:maxInviteChannelAttempts]
	}

	if len(candidates) == 0 {
		return nil, handler.ErrBadRequest("no_invite_channel", "The server has no text channel to create an invite in")
	}

	client, err := h.appStateManager.AppClient(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	// The app may lack the Create Invite permission in some channels, so the
	// next channel is tried until one works.
	for _, channel := range candidates {
		invite, err := client.WithContext(c.Context()).CreateInvite(channel.ID, api.CreateInviteData{
			MaxAge:         option.NewUint(600),
			MaxUses:        1,
			Unique:         true,
			AuditLogReason: "Requested from the Kite dashboard",
		})
		if err != nil {
			continue
		}

		return &wire.StateGuildInviteCreateResponse{
			URL: "https://discord.gg/" + invite.Code,
		}, nil
	}

	return nil, handler.ErrBadRequest("invite_create_failed", "The app doesn't have permission to create an invite in this server")
}
