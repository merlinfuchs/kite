package appstate

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	arikawastore "github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
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

func (h *AppStateHandler) HandleStateUserGet(c *handler.Context) (*wire.StateUserGetResponse, error) {
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if errors.Is(err, store.ErrNotFound) {
		return &wire.StateUserGetResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	user, err := state.AppUser(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get app user: %w", err)
	}

	return wire.StateUserToWire(user), nil
}

func (h *AppStateHandler) HandleStateGuildProfileGet(c *handler.Context) (*wire.StateGuildProfileGetResponse, error) {
	guildID, err := discord.ParseSnowflake(c.Param("guildID"))
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_guild_id", "Invalid guild ID")
	}

	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("app_not_running", "The app isn't running.")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	member, err := state.AppGuildBotMember(c.Context(), c.Param("guildID"))
	if err != nil {
		return nil, err
	}

	return &wire.StateGuildProfileGetResponse{
		Nick:      member.Nick,
		AvatarURL: member.AvatarURL(discord.GuildID(guildID)),
		BannerURL: member.BannerURL(discord.GuildID(guildID)),
	}, nil
}

func (h *AppStateHandler) HandleStateUserUpdate(c *handler.Context, req wire.StateUserUpdateRequest) (*wire.StateUserUpdateResponse, error) {
	for _, image := range []*string{req.Avatar, req.Banner} {
		if image != nil && util.IsAnimatedImageDataURI(*image) && !c.Features.AnimatedProfile {
			return nil, handler.ErrForbidden("feature_unavailable", "Your plan doesn't include animated avatars and banners.")
		}
	}

	client, err := h.appStateManager.AppClient(c.Context(), c.App.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	user, err := util.EditBotProfile(c.Context(), client, util.BotProfileData{
		Avatar: profileValue(req.Avatar),
		Banner: profileValue(req.Banner),
	})
	if err != nil {
		return nil, err
	}

	return wire.StateUserToWire(user), nil
}

func (h *AppStateHandler) HandleStateGuildList(c *handler.Context) (*wire.StateGuildListResponse, error) {
	// An app that isn't connected, or is in no servers, has no guilds to list.
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if errors.Is(err, store.ErrNotFound) {
		return &wire.StateGuildListResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	guilds, err := state.AppGuilds(c.Context())
	if errors.Is(err, arikawastore.ErrNotFound) {
		return &wire.StateGuildListResponse{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get guilds: %w", err)
	}

	res := make([]*wire.Guild, len(guilds))
	for i, guild := range guilds {
		w := wire.GuildToWire(&guild)
		if perms, err := state.AppGuildBotPermissions(c.Context(), guild.ID.String()); err == nil {
			w.BotPermissions = null.StringFrom(strconv.FormatUint(uint64(perms), 10))
		}
		res[i] = w
	}

	return &res, nil
}

func (h *AppStateHandler) HandleStateGuildChannelList(c *handler.Context) (*wire.StateGuildChannelListResponse, error) {
	guildID := c.Param("guildID")
	if _, err := discord.ParseSnowflake(guildID); err != nil {
		return nil, handler.ErrBadRequest("invalid_guild_id", "Invalid guild ID")
	}

	// An app that isn't connected has no channels to list, like guilds.
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if errors.Is(err, store.ErrNotFound) {
		return &wire.StateGuildChannelListResponse{}, nil
	}
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
	if _, err := discord.ParseSnowflake(guildID); err != nil {
		return nil, handler.ErrBadRequest("invalid_guild_id", "Invalid guild ID")
	}

	// An app that isn't connected has no roles to list, like guilds.
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if errors.Is(err, store.ErrNotFound) {
		return &wire.StateGuildRoleListResponse{}, nil
	}
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

// maxStateEmojis caps the emojis HandleStateEmojiList returns, so the picker
// stays usable for apps in thousands of guilds.
const maxStateEmojis = 1000

// HandleStateEmojiList returns the custom emojis of the guilds the app is in,
// grouped by guild, so they can be picked next to the app's own emojis. Guilds
// are added by name until maxStateEmojis is reached.
func (h *AppStateHandler) HandleStateEmojiList(c *handler.Context) (*wire.StateEmojiListResponse, error) {
	state, err := h.appStateManager.AppState(c.Context(), c.App.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// App is not running, so it has no guilds to take emojis from
			res := wire.StateEmojiListResponse{}
			return &res, nil
		}

		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	guilds, err := state.AppGuilds(c.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get guilds: %w", err)
	}

	sort.Slice(guilds, func(i, j int) bool {
		return strings.ToLower(guilds[i].Name) < strings.ToLower(guilds[j].Name)
	})

	res := wire.StateEmojiListResponse{}
	total := 0
	for _, guild := range guilds {
		if total >= maxStateEmojis {
			break
		}

		emojis, err := state.AppGuildEmojis(c.Context(), guild.ID.String())
		if err != nil {
			return nil, fmt.Errorf("failed to get guild emojis: %w", err)
		}

		group := &wire.StateGuildEmojis{
			GuildID:   guild.ID.String(),
			GuildName: guild.Name,
			Emojis:    make([]*wire.AppEmoji, 0, len(emojis)),
		}

		for _, emoji := range emojis {
			// Emojis can become unavailable when a guild loses boosts
			if !emoji.Available {
				continue
			}

			group.Emojis = append(group.Emojis, &wire.AppEmoji{
				ID:        emoji.ID.String(),
				Name:      emoji.Name,
				Animated:  emoji.Animated,
				Available: emoji.Available,
			})
		}

		if len(group.Emojis) > 0 {
			res = append(res, group)
			total += len(group.Emojis)
		}
	}

	return &res, nil
}

func (h *AppStateHandler) HandleStateGuildLeave(c *handler.Context) (*wire.StateGuildLeaveResponse, error) {
	guildID, err := strconv.ParseInt(c.Param("guildID"), 10, 64)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_guild_id", "Invalid guild ID")
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

func (h *AppStateHandler) HandleStateGuildProfileUpdate(c *handler.Context, req wire.StateGuildProfileUpdateRequest) (*wire.StateGuildProfileUpdateResponse, error) {
	if !c.Features.ServerProfiles {
		return nil, handler.ErrForbidden("feature_unavailable", "Your plan doesn't include server profiles.")
	}

	guildID, err := strconv.ParseInt(c.Param("guildID"), 10, 64)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_guild_id", "Invalid guild ID")
	}

	client, err := h.appStateManager.AppClient(c.Context(), c.App.ID)
	if errors.Is(err, store.ErrNotFound) {
		return nil, handler.ErrNotFound("app_not_running", "The app isn't running.")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app state: %w", err)
	}

	err = util.EditServerProfile(c.Context(), client, discord.GuildID(guildID), provider.ServerProfileData{
		Nick:   profileValue(req.Nick),
		Bio:    profileValue(req.Bio),
		Avatar: profileValue(req.Avatar),
		Banner: profileValue(req.Banner),
	})
	if err != nil {
		return nil, err
	}

	return &wire.StateGuildProfileUpdateResponse{}, nil
}

func profileValue(v *string) option.NullableString {
	if v == nil {
		return nil
	}
	if *v == "" {
		return option.NullString
	}
	return option.NewNullableString(*v)
}
