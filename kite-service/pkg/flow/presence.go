package flow

import (
	"net/url"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

// StatusActivity builds the activity shown on the bot's profile.
//
// Custom statuses only show State, so the text goes there. Every other type
// shows the name with State as an optional second line under it. A state equal
// to the name is dropped: older statuses stored a copy of the name there,
// which showed the text twice.
//
// Discord only uses the URL for Streaming activities (it backs the "Watch"
// button) and only accepts Twitch and YouTube links there, so it is dropped
// for every other type.
func StatusActivity(activityType discord.ActivityType, text string, state string, activityURL string) discord.Activity {
	activity := discord.Activity{
		Type: activityType,
		Name: text,
	}

	if activityType == discord.CustomActivity {
		activity.State = text
	} else if state = strings.TrimSpace(state); state != text {
		activity.State = state
	}

	if activityType == discord.StreamingActivity {
		activity.URL = discord.URL(strings.TrimSpace(activityURL))
	}

	return activity
}

// IsStreamURL reports whether Discord accepts the URL for a Streaming activity.
func IsStreamURL(raw string) bool {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return false
	}

	host := strings.ToLower(u.Hostname())
	for _, allowed := range []string{"twitch.tv", "youtube.com"} {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}
