package eval

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
)

func TestInviteCreateEventEnv(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	event := &gateway.InviteCreateEvent{
		Code:      "abc123",
		CreatedAt: discord.NewTimestamp(createdAt),
		ChannelID: 2,
		GuildID:   3,
		Inviter:   &discord.User{ID: 1, Username: "inviter"},
		InviteMetadata: discord.InviteMetadata{
			MaxAge:  7 * 24 * 60 * 60,
			MaxUses: 5,
		},
	}

	tests := map[string]string{
		"{{user.id}}":           "1",
		"{{user.mention}}":      "<@1>",
		"{{channel.id}}":        "2",
		"{{guild.id}}":          "3",
		"{{invite}}":            "abc123",
		"{{invite.code}}":       "abc123",
		"{{invite.url}}":        "https://discord.gg/abc123",
		"{{invite.max_age}}":    "604800",
		"{{invite.duration}}":   "7 days",
		"{{invite.created_at}}": "1767268800",
		"{{invite.expires_at}}": "1767873600",
		"{{invite.expires}}":    "<t:1767873600:R>",
		"{{invite.max_uses}}":   "5",
		"{{invite.temporary}}":  "false",
	}

	env := NewEventEnv(event)
	ctx := Context{Env: Env{
		"user":    env.User,
		"channel": env.Channel,
		"guild":   env.Guild,
		"invite":  env.Invite,
	}}
	for template, want := range tests {
		got, err := EvalTemplate(context.Background(), template, ctx)
		if err != nil {
			t.Errorf("%s: %v", template, err)
			continue
		}
		if got.String() != want {
			t.Errorf("%s = %q, want %q", template, got.String(), want)
		}
	}
}

// An invite that never expires has nothing to point a timestamp at.
func TestInviteEnvNeverExpires(t *testing.T) {
	env := NewInviteEnv(&gateway.InviteCreateEvent{
		Code:      "abc123",
		CreatedAt: discord.NewTimestamp(time.Now()),
	})

	if env.Duration != "never" || env.Expires != "never" || env.ExpiresAt != 0 {
		t.Errorf("got duration %q, expires %q, expires_at %d", env.Duration, env.Expires, env.ExpiresAt)
	}
}

func TestFormatInviteDuration(t *testing.T) {
	tests := map[time.Duration]string{
		30 * time.Minute:   "30 minutes",
		time.Hour:          "1 hour",
		90 * time.Minute:   "1 hour 30 minutes",
		24 * time.Hour:     "1 day",
		7 * 24 * time.Hour: "7 days",
		45 * time.Second:   "45 seconds",
		25*time.Hour + 5*time.Minute + 3*time.Second: "1 day 1 hour",
	}

	for d, want := range tests {
		if got := formatInviteDuration(d); got != want {
			t.Errorf("formatInviteDuration(%s) = %q, want %q", d, got, want)
		}
	}
}
