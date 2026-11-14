package auth

import "testing"

func TestIsDiscordUserAllowed(t *testing.T) {
	tests := []struct {
		name      string
		allowed   []string
		discordID string
		want      bool
	}{
		{name: "allowlist disabled", discordID: "123", want: true},
		{name: "listed user", allowed: []string{"123", "456"}, discordID: "456", want: true},
		{name: "unlisted user", allowed: []string{"123"}, discordID: "456", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := &AuthHandler{config: AuthHandlerConfig{AllowedDiscordIDs: test.allowed}}
			if got := handler.isDiscordUserAllowed(test.discordID); got != test.want {
				t.Errorf("isDiscordUserAllowed(%q) = %v, want %v", test.discordID, got, test.want)
			}
		})
	}
}
