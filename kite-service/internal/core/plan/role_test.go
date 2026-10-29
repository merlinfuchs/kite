package plan

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

type creatorStore struct {
	store.BundleStore
	ids []string
}

func (s creatorStore) CreatorDiscordIDs(context.Context, int) ([]string, error) {
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

type subscriptionList struct {
	store.SubscriptionStore
	subs []*model.Subscription
}

func (s subscriptionList) AllSubscriptions(context.Context) ([]*model.Subscription, error) {
	return s.subs, nil
}

type usersByID struct {
	store.UserStore
}

// User has the Discord ID of the user ID, so the tests can name both the same.
func (usersByID) User(_ context.Context, id string) (*model.User, error) {
	return &model.User{ID: id, DiscordID: id}, nil
}

func TestSyncPlanRoles(t *testing.T) {
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	sub := func(userID, productID, status string, endsAt time.Time) *model.Subscription {
		s := &model.Subscription{
			UserID:                userID,
			Status:                status,
			LemonsqueezyProductID: null.NewString(productID, true),
		}
		if !endsAt.IsZero() {
			s.EndsAt = null.NewTime(endsAt, true)
		}
		return s
	}
	unpaid := sub("107", "premium", model.SubscriptionStatusUnpaid, time.Time{})
	unpaid.UpdatedAt = now.Add(-24 * time.Hour)
	longUnpaid := sub("108", "premium", model.SubscriptionStatusUnpaid, time.Time{})
	longUnpaid.UpdatedAt = now.Add(-5 * 24 * time.Hour)
	subs := []*model.Subscription{
		// Pays for Premium and has the role.
		sub("101", "premium", model.SubscriptionStatusActive, time.Time{}),
		// Cancelled, but paid until next week.
		sub("102", "premium", model.SubscriptionStatusCancelled, now.Add(7*24*time.Hour)),
		// Ended yesterday, still within the grace period.
		sub("103", "premium", model.SubscriptionStatusExpired, now.Add(-24*time.Hour)),
		// Ended five days ago.
		sub("104", "premium", model.SubscriptionStatusExpired, now.Add(-5*24*time.Hour)),
		// Switched from Premium to Plus.
		sub("105", "plus", model.SubscriptionStatusActive, time.Time{}),
		// Not in the server.
		sub("106", "premium", model.SubscriptionStatusActive, time.Time{}),
		// Unpaid since yesterday, which has no end.
		unpaid,
		// Unpaid for five days.
		longUnpaid,
	}
	// Every member has the Premium role and a role that isn't a plan's.
	var mu sync.Mutex
	var changed []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			if strings.HasSuffix(r.URL.Path, "/106") {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"code": 10007, "message": "Unknown Member"}`))
				return
			}
			w.Write([]byte(`{"user": {"id": "1"}, "roles": ["7", "99"]}`))
			return
		}
		mu.Lock()
		changed = append(changed, r.Method+" "+r.URL.Path)
		mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()
	endpoint := api.EndpointGuilds
	api.EndpointGuilds = srv.URL + "/guilds/"
	defer func() { api.EndpointGuilds = endpoint }()

	plans := []model.Plan{
		{ID: "basic"},
		{ID: "plus", LemonSqueezyProductID: "plus", DiscordRoleID: "8"},
		{ID: "premium", LemonSqueezyProductID: "premium", DiscordRoleID: "7"},
	}
	m := NewPlanManager(nil, subscriptionList{subs: subs}, usersByID{}, nil, plans, PlanManagerConfig{})
	r := roleGranter{client: api.NewClient("Bot token"), guildID: 1}
	if err := m.syncPlanRoles(context.Background(), r, now); err != nil {
		t.Fatal(err)
	}

	sort.Strings(changed)
	want := []string{
		"DELETE /guilds/1/members/104/roles/7",
		"DELETE /guilds/1/members/105/roles/7",
		"DELETE /guilds/1/members/108/roles/7",
		"PUT /guilds/1/members/105/roles/8",
	}
	if strings.Join(changed, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", changed, want)
	}
}
