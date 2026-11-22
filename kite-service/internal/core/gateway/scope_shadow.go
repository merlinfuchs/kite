package gateway

import (
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/diamondburned/arikawa/v3/utils/httputil/httpdriver"
	"github.com/kitecloud/kite/kite-service/internal/metrics"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

// shadowScope counts how often the requests of an own app name a channel that
// isn't in the cache. Shared apps can only reach channels of their server, see
// the sharedbot package, and ask Discord for the server of such a channel
// first. Own apps aren't checked, so this only shows what that would cost.
func shadowScope(session *state.State) {
	hc := session.Client.Client
	hc.Client = shadowDriver{Client: hc.Client, channels: session.Cabinet.ChannelStore}
}

type shadowDriver struct {
	httpdriver.Client
	channels store.ChannelStore
}

func (d shadowDriver) Do(r httpdriver.Request) (httpdriver.Response, error) {
	segments := util.DiscordAPISegments(r.GetPath())
	if len(segments) >= 2 && segments[0] == "channels" {
		if id, err := discord.ParseSnowflake(segments[1]); err == nil {
			if _, err := d.channels.Channel(discord.ChannelID(id)); err == nil {
				metrics.SharedBotShadow.Add("channel_cached", 1)
			} else {
				metrics.SharedBotShadow.Add("channel_uncached", 1)
			}
		}
	}
	return d.Client.Do(r)
}
