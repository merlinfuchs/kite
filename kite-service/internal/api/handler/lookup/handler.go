package lookup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"gopkg.in/guregu/null.v4"
)

const cacheDuration = 10 * time.Minute

var snowflakeRegex = regexp.MustCompile(`^[0-9]{17,20}$`)

// LookupHandler looks up Discord users for the lookup tool on the website,
// with the bot of the Kite server. Discord only shows users to bots.
type LookupHandler struct {
	botToken string
	baseURL  string
	client   *http.Client
	// Users by ID, nil for IDs that aren't a user.
	cache *ristretto.Cache
}

func NewLookupHandler(botToken string) (*LookupHandler, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 100_000,
		MaxCost:     10_000,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &LookupHandler{
		botToken: botToken,
		baseURL:  "https://discord.com/api/v10",
		client:   &http.Client{Timeout: 10 * time.Second},
		cache:    cache,
	}, nil
}

type discordUser struct {
	ID                   string      `json:"id"`
	Username             string      `json:"username"`
	GlobalName           null.String `json:"global_name"`
	Discriminator        string      `json:"discriminator"`
	Avatar               null.String `json:"avatar"`
	Banner               null.String `json:"banner"`
	AccentColor          null.Int    `json:"accent_color"`
	Bot                  bool        `json:"bot"`
	System               bool        `json:"system"`
	PublicFlags          int         `json:"public_flags"`
	AvatarDecorationData *struct {
		Asset string `json:"asset"`
	} `json:"avatar_decoration_data"`
	PrimaryGuild *struct {
		IdentityGuildID null.String `json:"identity_guild_id"`
		IdentityEnabled bool        `json:"identity_enabled"`
		Tag             null.String `json:"tag"`
		Badge           null.String `json:"badge"`
	} `json:"primary_guild"`
}

func (h *LookupHandler) HandleUserGet(c *handler.Context) (*wire.LookupUserGetResponse, error) {
	userID := c.Param("userID")
	if !snowflakeRegex.MatchString(userID) {
		return nil, handler.ErrBadRequest("invalid_user_id", "User IDs are numbers with 17 to 20 digits.")
	}

	if h.botToken == "" {
		return nil, handler.ErrServiceUnavailable("lookup_unavailable", "User lookup isn't set up on this server.")
	}

	if cached, ok := h.cache.Get(userID); ok {
		user, _ := cached.(*wire.LookupUserGetResponse)
		if user == nil {
			return nil, errUnknownUser
		}
		return user, nil
	}

	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, h.baseURL+"/users/"+userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bot "+h.botToken)

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		h.cache.SetWithTTL(userID, (*wire.LookupUserGetResponse)(nil), 1, cacheDuration)
		return nil, errUnknownUser
	case http.StatusTooManyRequests:
		return nil, handler.ErrRateLimit("Too many lookups right now. Please try again in a minute.")
	default:
		return nil, fmt.Errorf("failed to get user: discord returned status %d", resp.StatusCode)
	}

	var du discordUser
	if err := json.NewDecoder(resp.Body).Decode(&du); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	user := &wire.LookupUserGetResponse{
		ID:            du.ID,
		Username:      du.Username,
		GlobalName:    du.GlobalName,
		Discriminator: du.Discriminator,
		Avatar:        du.Avatar,
		Banner:        du.Banner,
		AccentColor:   du.AccentColor,
		Bot:           du.Bot,
		System:        du.System,
		PublicFlags:   du.PublicFlags,
	}
	if du.AvatarDecorationData != nil && du.AvatarDecorationData.Asset != "" {
		user.AvatarDecoration = null.StringFrom(du.AvatarDecorationData.Asset)
	}
	if pg := du.PrimaryGuild; pg != nil && pg.IdentityEnabled && pg.Tag.Valid {
		user.PrimaryGuild = &wire.LookupUserPrimaryGuild{
			GuildID: pg.IdentityGuildID.String,
			Tag:     pg.Tag.String,
			Badge:   pg.Badge.String,
		}
	}

	h.cache.SetWithTTL(userID, user, 1, cacheDuration)
	return user, nil
}

var errUnknownUser = handler.ErrNotFound("unknown_user", "There is no Discord user with this ID.")
