package flow

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/stretchr/testify/assert"
)

func TestStatusActivity(t *testing.T) {
	tests := []struct {
		name         string
		activityType discord.ActivityType
		state        string
		url          string
		want         discord.Activity
	}{
		{
			// State would be rendered as a second line with the same text
			name:         "playing has no state",
			activityType: discord.GameActivity,
			want:         discord.Activity{Type: discord.GameActivity, Name: "Kite"},
		},
		{
			name:         "streaming keeps url and has no state",
			activityType: discord.StreamingActivity,
			url:          " https://twitch.tv/kite ",
			want: discord.Activity{
				Type: discord.StreamingActivity,
				Name: "Kite",
				URL:  "https://twitch.tv/kite",
			},
		},
		{
			name:         "custom shows the text as state",
			activityType: discord.CustomActivity,
			want:         discord.Activity{Type: discord.CustomActivity, Name: "Kite", State: "Kite"},
		},
		{
			name:         "playing shows state as second line",
			activityType: discord.GameActivity,
			state:        " 42 servers ",
			want:         discord.Activity{Type: discord.GameActivity, Name: "Kite", State: "42 servers"},
		},
		{
			// Older statuses stored a copy of the name as the state
			name:         "state equal to name is dropped",
			activityType: discord.ListeningActivity,
			state:        "Kite",
			want:         discord.Activity{Type: discord.ListeningActivity, Name: "Kite"},
		},
		{
			name:         "custom ignores state",
			activityType: discord.CustomActivity,
			state:        "Other",
			want:         discord.Activity{Type: discord.CustomActivity, Name: "Kite", State: "Kite"},
		},
		{
			name:         "url dropped for non streaming",
			activityType: discord.WatchingActivity,
			url:          "https://twitch.tv/kite",
			want:         discord.Activity{Type: discord.WatchingActivity, Name: "Kite"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, StatusActivity(tt.activityType, "Kite", tt.state, tt.url))
		})
	}
}

func TestIsStreamURL(t *testing.T) {
	assert.True(t, IsStreamURL("https://www.twitch.tv/kite"))
	assert.True(t, IsStreamURL("https://youtube.com/watch?v=abc"))
	assert.True(t, IsStreamURL("https://www.youtube.com/@kite/live"))
	assert.False(t, IsStreamURL("https://kite.onl"))
	assert.False(t, IsStreamURL("https://nottwitch.tv/kite"))
	assert.False(t, IsStreamURL("twitch.tv/kite"))
	assert.False(t, IsStreamURL(""))
}
