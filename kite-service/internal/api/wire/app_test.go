package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppInstallationUpdateRequestValidate(t *testing.T) {
	guild := func(permissions string, scopes ...string) *AppInstallParams {
		return &AppInstallParams{Scopes: scopes, Permissions: permissions}
	}

	tests := []struct {
		name    string
		req     AppInstallationUpdateRequest
		wantErr bool
	}{
		{"server with bot", AppInstallationUpdateRequest{GuildInstall: guild("8", "applications.commands", "bot")}, false},
		{"server without bot", AppInstallationUpdateRequest{GuildInstall: guild("0", "applications.commands")}, true},
		{"account", AppInstallationUpdateRequest{UserInstall: guild("0", "applications.commands")}, false},
		{"both", AppInstallationUpdateRequest{GuildInstall: guild("8", "applications.commands", "bot"), UserInstall: guild("0", "applications.commands")}, false},
		{"server without permissions", AppInstallationUpdateRequest{GuildInstall: guild("0", "applications.commands", "bot")}, true},
		{"none", AppInstallationUpdateRequest{}, true},
		{"no commands scope", AppInstallationUpdateRequest{GuildInstall: guild("8", "bot")}, true},
		{"unknown scope", AppInstallationUpdateRequest{GuildInstall: guild("8", "applications.commands", "bot", "identify")}, true},
		{"duplicate scope", AppInstallationUpdateRequest{GuildInstall: guild("8", "applications.commands", "bot", "bot")}, true},
		{"permissions without bot", AppInstallationUpdateRequest{UserInstall: guild("8", "applications.commands")}, true},
		{"invalid permissions", AppInstallationUpdateRequest{GuildInstall: guild("-1", "applications.commands", "bot")}, true},
		{"empty permissions", AppInstallationUpdateRequest{GuildInstall: guild("", "applications.commands", "bot")}, true},
		{"bot on account", AppInstallationUpdateRequest{UserInstall: guild("0", "applications.commands", "bot")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
