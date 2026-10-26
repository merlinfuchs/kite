package wire

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestAPITokenCreateRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     APITokenCreateRequest
		wantErr bool
	}{
		{"no expiry", APITokenCreateRequest{Name: "Claude"}, false},
		{"30 days", APITokenCreateRequest{Name: "Claude", ExpiresInDays: null.IntFrom(30)}, false},
		{"max days", APITokenCreateRequest{Name: "Claude", ExpiresInDays: null.IntFrom(MaxAPITokenDays)}, false},
		{"no name", APITokenCreateRequest{}, true},
		{"blank name", APITokenCreateRequest{Name: "   "}, true},
		{"zero days", APITokenCreateRequest{Name: "Claude", ExpiresInDays: null.IntFrom(0)}, true},
		{"negative days", APITokenCreateRequest{Name: "Claude", ExpiresInDays: null.IntFrom(-1)}, true},
		{"too many days", APITokenCreateRequest{Name: "Claude", ExpiresInDays: null.IntFrom(MaxAPITokenDays + 1)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			assert.Equal(t, tt.wantErr, err != nil, "error: %v", err)
		})
	}
}
