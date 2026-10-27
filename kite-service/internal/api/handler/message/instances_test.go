package message

import (
	"errors"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func TestTokenCanSend(t *testing.T) {
	write := &model.APIToken{Scopes: []string{model.APITokenScopeKiteWrite}}
	read := &model.APIToken{Scopes: []string{model.APITokenScopeKiteRead}}
	fromTemplate := &model.Message{BundleInstallID: null.StringFrom("install")}
	own := &model.Message{}

	tests := []struct {
		name     string
		token    *model.APIToken
		msg      *model.Message
		sent     bool
		wantCode string
	}{
		{"template message", write, fromTemplate, false, ""},
		{"read only token", read, fromTemplate, false, "missing_api_token_scope"},
		{"own message", write, own, false, "api_token_not_allowed"},
		{"sent before", write, fromTemplate, true, "message_already_sent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tokenCanSend(tt.token, tt.msg, tt.sent)
			var herr *handler.Error
			code := ""
			if errors.As(err, &herr) {
				code = herr.Code
			} else if err != nil {
				t.Fatalf("got %v, want a handler error", err)
			}
			if code != tt.wantCode {
				t.Errorf("got %q, want %q", code, tt.wantCode)
			}
		})
	}
}
