package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
)

// syncCreators gives the creator role to every user with a template in the
// marketplace who is in the server, every hour. Users not in it yet get it
// once they joined. Taking it away is left to the staff, like when a
// template is taken down.
func (b *Bot) syncCreators(ctx context.Context) {
	if b.creatorGuild == 0 || b.creatorRole == 0 {
		return
	}

	// The ones that have it since the bot started, so it doesn't ask
	// Discord again every hour.
	granted := map[discord.UserID]bool{}
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		creators, err := fetchCreators(ctx, b.cfg.Creators.BundlesURL)
		if err != nil {
			slog.With("err", err).Warn("fetch creators failed")
		}
		for _, userID := range creators {
			if granted[userID] {
				continue
			}
			err := b.state.AddRole(b.creatorGuild, userID, b.creatorRole, api.AddRoleData{
				AuditLogReason: "kite-support: has a template in the marketplace",
			})
			var httpErr *httputil.HTTPError
			switch {
			case err == nil:
				granted[userID] = true
			case errors.As(err, &httpErr) && httpErr.Status == http.StatusNotFound:
				// Not in the server.
			default:
				slog.With("err", err).Warn("grant creator role failed", "user_id", userID)
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// fetchCreators are the Discord IDs of the users with a template in the
// marketplace.
func fetchCreators(ctx context.Context, url string) ([]discord.UserID, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch bundles: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch bundles: status %d", resp.StatusCode)
	}

	var body struct {
		Success bool `json:"success"`
		Data    []struct {
			Owner *struct {
				DiscordID string `json:"discord_id"`
			} `json:"owner"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode bundles: %w", err)
	}
	if !body.Success {
		return nil, fmt.Errorf("fetch bundles: request failed")
	}

	seen := map[discord.UserID]bool{}
	var creators []discord.UserID
	for _, bundle := range body.Data {
		// Kite's own templates have no owner.
		if bundle.Owner == nil {
			continue
		}
		id, err := discord.ParseSnowflake(bundle.Owner.DiscordID)
		if err != nil || seen[discord.UserID(id)] {
			continue
		}
		seen[discord.UserID(id)] = true
		creators = append(creators, discord.UserID(id))
	}
	return creators, nil
}
