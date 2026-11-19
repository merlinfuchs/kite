package access

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
)

// Apps on the Kite bot can only change things with the routes that allow
// them, and read with all.
func TestCustomBotOnly(t *testing.T) {
	m := &AccessManager{}

	for _, botType := range []model.AppBotType{model.AppBotTypeShared, model.AppBotTypeOwn} {
		withApp := func(next handler.HandlerFunc) handler.HandlerFunc {
			return func(c *handler.Context) error {
				c.App = &model.App{ID: "app", BotType: botType}
				return next(c)
			}
		}
		ok := func(c *handler.Context) error {
			return c.Send(http.StatusNoContent, nil)
		}

		mux := http.NewServeMux()
		group := handler.Group(mux, "/app", withApp, m.CustomBotOnly)
		group.Post("/build", ok)
		group.Get("/read", ok)
		group.AllowSharedApp().Post("/templates", ok)
		group.AllowSharedApp().Group("/nested").Post("/x", ok)

		tests := []struct {
			method, path  string
			sharedAllowed bool
		}{
			{"POST", "/app/build", false},
			{"GET", "/app/read", true},
			{"POST", "/app/templates", true},
			{"POST", "/app/nested/x", true},
		}
		for _, tt := range tests {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			allowed := rec.Code == http.StatusNoContent
			want := botType == model.AppBotTypeOwn || tt.sharedAllowed
			if allowed != want {
				t.Errorf("%s app, %s %s: allowed = %v, want %v (status %d)", botType, tt.method, tt.path, allowed, want, rec.Code)
			}
		}
	}
}
