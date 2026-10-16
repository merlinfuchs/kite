package access

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

func TestSupportStaff(t *testing.T) {
	tests := []struct {
		name           string
		allowedUserIDs []string
		userID         string
		wantStatus     int
	}{
		{
			name:           "listed user is allowed",
			allowedUserIDs: []string{"staff-user"},
			userID:         "staff-user",
			wantStatus:     http.StatusOK,
		},
		{
			name:           "unlisted user is forbidden",
			allowedUserIDs: []string{"staff-user"},
			userID:         "other-user",
			wantStatus:     http.StatusForbidden,
		},
		{
			name:       "empty allowlist denies everyone",
			userID:     "staff-user",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			group := handler.Group(mux, "/support",
				func(next handler.HandlerFunc) handler.HandlerFunc {
					return func(c *handler.Context) error {
						c.Session = &model.Session{UserID: tt.userID}
						return next(c)
					}
				},
				SupportStaff(tt.allowedUserIDs),
			)
			group.Get("/access", func(c *handler.Context) error {
				return c.Send(http.StatusOK, []byte("authorized"))
			})

			res := httptest.NewRecorder()
			mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/support/access", nil))
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}
		})
	}
}
