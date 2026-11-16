package wire

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserUpdateRequestSanitize(t *testing.T) {
	tests := []struct {
		name string
		bio  string
		want string
	}{
		{"empty", "", ""},
		{"spaces", "  hi there  ", "hi there"},
		{"windows newlines", "one\r\ntwo", "one\ntwo"},
		{"paragraphs", "one\n\n\n\ntwo", "one\n\ntwo"},
		{"trailing spaces", "one   \ntwo\t", "one\ntwo"},
		{"only newlines", "\n\n\n", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := UserUpdateRequest{Bio: tt.bio}
			req.Sanitize()
			assert.Equal(t, tt.want, req.Bio)
		})
	}
}

func TestUserUpdateRequestValidate(t *testing.T) {
	tests := []struct {
		name    string
		bio     string
		wantErr bool
	}{
		{"empty", "", false},
		{"short", "I make moderation templates.", false},
		{"max length", strings.Repeat("a", 300), false},
		// Characters count, not bytes.
		{"max length emoji", strings.Repeat("🪁", 300), false},
		{"too long", strings.Repeat("a", 301), true},
		{"max lines", "1\n2\n3\n4\n5", false},
		{"too many lines", "1\n2\n3\n4\n5\n6", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UserUpdateRequest{Bio: tt.bio}.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
