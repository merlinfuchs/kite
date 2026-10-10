package appstate

import (
"errors"
"fmt"
"sort"
"strconv"
"strings"

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

// HandleStateEmojiList returns the custom emojis of every guild the app is in,
// grouped by guild, so they can be picked next to the app's own emojis.
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

res := make([]*wire.StateGuildEmojis, 0, len(guilds))
for _, guild := range guilds {
emojis, err := state.AppGuildEmojis(c.Context(), guild.ID.String())
if err != nil {
return nil, fmt.Errorf("failed to get guild emojis: %w", err)
}

group := &wire.StateGuildEmojis{
GuildID:   guild.ID.String(),
GuildName: guild.Name,
Emojis:    make([]*wire.AppEmoji, 0, len(emojis)),
}

iconURL := guild.IconURL()
group.GuildIconURL = null.NewString(iconURL, iconURL != "")

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
}
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
