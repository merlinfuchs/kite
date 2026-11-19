package auth

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// fakeInstallStore keeps the shared apps of servers and their collaborators.
// Other methods aren't used by sharedApp.
type fakeInstallStore struct {
	store.AppStore
	store.UserStore

	apps          map[string]*model.App
	collaborators map[string]bool
}

func (s *fakeInstallStore) UserBan(ctx context.Context, userID string, feature model.UserBanFeature) (*model.UserBan, error) {
	return nil, store.ErrNotFound
}

func (s *fakeInstallStore) SharedAppByGuild(ctx context.Context, guildID string) (*model.App, error) {
	if app, ok := s.apps[guildID]; ok {
		return app, nil
	}
	return nil, store.ErrNotFound
}

func (s *fakeInstallStore) CountAppsByUser(ctx context.Context, userID string) (int, error) {
	return 0, nil
}

func (s *fakeInstallStore) CreateSharedApp(ctx context.Context, opts store.SharedAppCreateOpts) (*model.App, error) {
	app := &model.App{ID: opts.ID, Name: opts.Name, OwnerUserID: opts.OwnerUserID, BotType: model.AppBotTypeShared, DiscordGuildID: null.StringFrom(opts.DiscordGuildID), Ref: opts.Ref}
	s.apps[opts.DiscordGuildID] = app
	return app, nil
}

func (s *fakeInstallStore) EnableSharedAppDisabledWithReason(ctx context.Context, guildID string, reason string, now time.Time) (bool, error) {
	return false, nil
}

func (s *fakeInstallStore) Collaborator(ctx context.Context, appID string, userID string) (*model.AppCollaborator, error) {
	if s.collaborators[appID+"/"+userID] {
		return &model.AppCollaborator{AppID: appID, UserID: userID}, nil
	}
	return nil, store.ErrNotFound
}

func (s *fakeInstallStore) CreateCollaborator(ctx context.Context, c *model.AppCollaborator) (*model.AppCollaborator, error) {
	s.collaborators[c.AppID+"/"+c.UserID] = true
	return c, nil
}

// The first user to add the bot to a server gets its app with the ref of
// their link, and the next one gets the same app.
func TestSharedApp(t *testing.T) {
	s := &fakeInstallStore{apps: map[string]*model.App{}, collaborators: map[string]bool{}}
	h := &AuthHandler{appStore: s, userStore: s, install: InstallConfig{MaxAppsPerUser: 10}}

	first, created, code, err := h.sharedApp(context.Background(), &model.User{ID: "u1"}, "100", "Server", null.StringFrom("xenon"))
	if err != nil || code != "" || first == nil {
		t.Fatalf("first install: app %v, code %q, err %v", first, code, err)
	}
	if !created {
		t.Error("first install didn't say it created the app")
	}
	if first.OwnerUserID != "u1" {
		t.Errorf("owner = %s, want u1", first.OwnerUserID)
	}
	if first.Ref != null.StringFrom("xenon") {
		t.Errorf("ref = %v, want xenon", first.Ref)
	}

	second, created2, code, err := h.sharedApp(context.Background(), &model.User{ID: "u2"}, "100", "Server", null.StringFrom("other"))
	if err != nil || code != "" || second == nil || second.ID != first.ID {
		t.Fatalf("second install: app %v, code %q, err %v", second, code, err)
	}
	if created2 {
		t.Error("second install said it created the app")
	}
	if !s.collaborators[first.ID+"/u2"] {
		t.Error("the second user wasn't added to the app's list")
	}
	if second.Ref != null.StringFrom("xenon") {
		t.Errorf("ref = %v, want the first one, xenon", second.Ref)
	}
}

func TestRefPattern(t *testing.T) {
	for ref, want := range map[string]bool{
		"xenon":                             true,
		"x-1_2":                             true,
		"":                                  false,
		"Xenon":                             false,
		"a b":                               false,
		"xenon.gg":                          false,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": false,
	} {
		if got := refPattern.MatchString(ref); got != want {
			t.Errorf("refPattern.MatchString(%q) = %v, want %v", ref, got, want)
		}
	}
}
