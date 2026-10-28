package plan

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type creatorStore struct {
	store.BundleStore
	ids []string
}

func (s creatorStore) ListedBundleOwnerDiscordIDs(context.Context) ([]string, error) {
	return s.ids, nil
}

type brokenSubscriptionStore struct {
	store.SubscriptionStore
}

func (brokenSubscriptionStore) AllSubscriptions(context.Context) ([]*model.Subscription, error) {
	return nil, errors.New("broken")
}

func TestHandOutCreatorRoles(t *testing.T) {
	var mu sync.Mutex
	var granted []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		granted = append(granted, r.Method+" "+r.URL.Path)
		mu.Unlock()
		// The second creator isn't in the server.
		if strings.Contains(r.URL.Path, "/members/222/") {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"code": 10007, "message": "Unknown Member"}`))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	endpoint := api.EndpointGuilds
	api.EndpointGuilds = srv.URL + "/guilds/"
	defer func() { api.EndpointGuilds = endpoint }()

	m := NewPlanManager(nil, brokenSubscriptionStore{}, nil, creatorStore{ids: []string{"111", "222"}}, nil, PlanManagerConfig{
		DiscordBotToken:      "token",
		DiscordGuildID:       "1",
		DiscordCreatorRoleID: "9",
	})
	err := m.handOutRoles(context.Background())

	// The plan roles failing doesn't keep the creators from getting theirs.
	if err == nil || !strings.Contains(err.Error(), "subscriptions") {
		t.Fatalf("expected the subscriptions error, got %v", err)
	}
	want := []string{"PUT /guilds/1/members/111/roles/9", "PUT /guilds/1/members/222/roles/9"}
	if strings.Join(granted, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", granted, want)
	}
}
