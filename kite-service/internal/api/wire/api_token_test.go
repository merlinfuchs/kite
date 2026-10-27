package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestAPITokenCreateRequestValidate(t *testing.T) {
	write := []string{"kite:write"}
	tests := []struct {
		name    string
		req     APITokenCreateRequest
		wantErr bool
	}{
		{"no expiry", APITokenCreateRequest{Name: "Claude", Scopes: write}, false},
		{"30 days", APITokenCreateRequest{Name: "Claude", Scopes: write, ExpiresInDays: null.IntFrom(30)}, false},
		{"max days", APITokenCreateRequest{Name: "Claude", Scopes: write, ExpiresInDays: null.IntFrom(MaxAPITokenDays)}, false},
		{"read only", APITokenCreateRequest{Name: "Claude", Scopes: []string{"kite:read"}}, false},
		{"some apps", APITokenCreateRequest{Name: "Claude", Scopes: write, AppIDs: []string{"app"}}, false},
		{"no name", APITokenCreateRequest{Scopes: write}, true},
		{"blank name", APITokenCreateRequest{Name: "   ", Scopes: write}, true},
		{"zero days", APITokenCreateRequest{Name: "Claude", Scopes: write, ExpiresInDays: null.IntFrom(0)}, true},
		{"negative days", APITokenCreateRequest{Name: "Claude", Scopes: write, ExpiresInDays: null.IntFrom(-1)}, true},
		{"too many days", APITokenCreateRequest{Name: "Claude", Scopes: write, ExpiresInDays: null.IntFrom(MaxAPITokenDays + 1)}, true},
		{"no scopes", APITokenCreateRequest{Name: "Claude"}, true},
		{"unknown scope", APITokenCreateRequest{Name: "Claude", Scopes: []string{"all"}}, true},
		// Null means all apps, so no apps must not turn into all of them.
		{"no apps", APITokenCreateRequest{Name: "Claude", Scopes: write, AppIDs: []string{}}, true},
		{"too many apps", APITokenCreateRequest{Name: "Claude", Scopes: write, AppIDs: make([]string, MaxAPITokenApps+1)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			assert.Equal(t, tt.wantErr, err != nil, "error: %v", err)
		})
	}
}
