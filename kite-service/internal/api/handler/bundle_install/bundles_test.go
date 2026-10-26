package bundleinstall

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

// userBundleStore keeps the bundles users made.
type userBundleStore struct {
	store.BundleStore
	bundles map[string]*model.Bundle
	reviews []*model.BundleReview
}

func (s *userBundleStore) Bundle(ctx context.Context, id string) (*model.Bundle, error) {
	b, ok := s.bundles[id]
	if !ok {
		return nil, store.ErrNotFound
	}
	copied := *b
	return &copied, nil
}

func (s *userBundleStore) CreateBundle(ctx context.Context, b *model.Bundle, maxPerOwner int) error {
	count := 0
	for _, other := range s.bundles {
		if other.OwnerUserID == b.OwnerUserID {
			count++
		}
	}
	if count >= maxPerOwner {
		return store.ErrLimitReached
	}
	s.bundles[b.ID] = b
	return nil
}

func (s *userBundleStore) PublishBundleVersion(ctx context.Context, b *model.Bundle, fromVersion int) error {
	if s.bundles[b.ID].LatestVersion != fromVersion {
		return store.ErrOutdated
	}
	s.bundles[b.ID] = b
	return nil
}

func (s *memoryCommandStore) CommandsByApp(ctx context.Context, appID string) ([]*model.Command, error) {
	return s.commands, nil
}

func (s *userBundleStore) SubmitBundleReview(ctx context.Context, review *model.BundleReview) error {
	review.Version = s.bundles[review.BundleID].LatestVersion
	s.reviews = append(s.reviews, review)
	return nil
}

// The app stores only have commands.
type appEventListenerStore struct{ store.EventListenerStore }

func (s *appEventListenerStore) EventListenersByApp(ctx context.Context, appID string) ([]*model.EventListener, error) {
	return nil, nil
}

type appMessageStore struct{ store.MessageStore }

func (s *appMessageStore) MessagesByApp(ctx context.Context, appID string) ([]*model.Message, error) {
	return nil, nil
}

func bundlesHandler(commandCount int) (*BundleInstallHandler, *userBundleStore) {
	commands := &memoryCommandStore{}
	for i := range commandCount {
		commands.commands = append(commands.commands, &model.Command{
			ID: fmt.Sprintf("cmd%d", i),
			FlowSource: flow.FlowData{Nodes: []flow.FlowNode{
				{ID: "entry", Type: flow.FlowNodeTypeEntryCommand, Data: flow.FlowNodeData{Name: fmt.Sprintf("cmd%d", i)}},
			}},
		})
	}
	bundles := &userBundleStore{bundles: map[string]*model.Bundle{}}
	h := NewBundleInstallHandler(
		nil, bundles, commands, &appEventListenerStore{}, &appMessageStore{}, &memoryVariableStore{},
		nil, nil, nil, nil, 2,
	)
	return h, bundles
}

// callAs calls the handler as the user in the app.
func callAs(t *testing.T, h handler.HandlerFunc, userID, appID, bundleID string, req any) (int, map[string]any) {
	t.Helper()

	srv := handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: appID}
		c.Session = &model.Session{UserID: userID}
		return h(c)
	})

	body, err := json.Marshal(req)
	require.NoError(t, err)
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")
	r.SetPathValue("bundleID", bundleID)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, r)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func errorCode(res map[string]any) string {
	e, _ := res["error"].(map[string]any)
	code, _ := e["code"].(string)
	return code
}

func TestBundleCreateAndPublish(t *testing.T) {
	h, bundles := bundlesHandler(3)
	create := handler.TypedWithBody(h.HandleBundleCreate)
	publish := handler.TypedWithBody(h.HandleBundleVersionCreate)

	code, res := callAs(t, create, "user", "app", "", wire.BundleCreateRequest{Name: "Mod", BundleExportRequest: wire.BundleExportRequest{CommandIDs: []string{"cmd0"}}})
	require.Equal(t, http.StatusOK, code, res)
	id := res["data"].(map[string]any)["id"].(string)

	b := bundles.bundles[id]
	assert.Equal(t, "user", b.OwnerUserID.String)
	assert.Equal(t, "app", b.SourceAppID.String)
	content, err := bundle.Parse(b.Content)
	require.NoError(t, err)
	// Installs use the ID and version of the content.
	assert.Equal(t, id, content.ID)
	assert.Equal(t, 1, content.Version)

	req := wire.BundleVersionCreateRequest{
		BundleCreateRequest: wire.BundleCreateRequest{Name: "Moderation", BundleExportRequest: wire.BundleExportRequest{CommandIDs: []string{"cmd1", "cmd2"}}},
		Notes:               " Added /kick. ",
	}
	code, res = callAs(t, publish, "user", "other_app", id, req)
	assert.Equal(t, "wrong_app", errorCode(res), code)
	code, res = callAs(t, publish, "other_user", "app", id, req)
	assert.Equal(t, "unknown_bundle", errorCode(res), code)

	code, res = callAs(t, publish, "user", "app", id, req)
	require.Equal(t, http.StatusOK, code, res)
	b = bundles.bundles[id]
	assert.Equal(t, 2, b.Version)
	assert.Equal(t, "Moderation", b.Name)
	assert.Equal(t, "Added /kick.", b.Notes)
	content, err = bundle.Parse(b.Content)
	require.NoError(t, err)
	assert.Equal(t, 2, content.Version)
	// Items can change between versions.
	assert.Len(t, content.Commands, 2)
	assert.Equal(t, "cmd1", content.Commands[0].ID)
}

func TestBundleCreateLimits(t *testing.T) {
	h, _ := bundlesHandler(maxBundleItems + 1)
	create := handler.TypedWithBody(h.HandleBundleCreate)

	var all []string
	for i := range maxBundleItems + 1 {
		all = append(all, fmt.Sprintf("cmd%d", i))
	}
	code, res := callAs(t, create, "user", "app", "", wire.BundleCreateRequest{Name: "All", BundleExportRequest: wire.BundleExportRequest{CommandIDs: all}})
	assert.Equal(t, "too_many_items", errorCode(res), code)

	// The handler allows 2 bundles per user.
	for range 2 {
		code, res = callAs(t, create, "user", "app", "", wire.BundleCreateRequest{Name: "One", BundleExportRequest: wire.BundleExportRequest{CommandIDs: []string{"cmd0"}}})
		require.Equal(t, http.StatusOK, code, res)
	}
	code, res = callAs(t, create, "user", "app", "", wire.BundleCreateRequest{Name: "One", BundleExportRequest: wire.BundleExportRequest{CommandIDs: []string{"cmd0"}}})
	assert.Equal(t, "too_many_bundles", errorCode(res), code)
}

// Anyone with its link can install a user's bundle, until Kite takes it down.
func TestCurrentBundleOfUser(t *testing.T) {
	h, bundles := bundlesHandler(1)
	create := handler.TypedWithBody(h.HandleBundleCreate)
	code, res := callAs(t, create, "user", "app", "", wire.BundleCreateRequest{Name: "Mod", BundleExportRequest: wire.BundleExportRequest{CommandIDs: []string{"cmd0"}}})
	require.Equal(t, http.StatusOK, code, res)
	id := res["data"].(map[string]any)["id"].(string)

	get := func(userID string) error {
		var err error
		srv := handler.APIHandler(func(c *handler.Context) error {
			c.App = &model.App{ID: "other_app"}
			c.Session = &model.Session{UserID: userID}
			_, err = h.currentBundle(c, id)
			return nil
		})
		srv.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		return err
	}

	assert.NoError(t, get("user"))
	assert.NoError(t, get("other_user"))

	// The app it's made from has its items already.
	install := handler.TypedWithBody(h.HandleBundleInstallCreate)
	code, res = callAs(t, install, "user", "app", "", wire.BundleInstallCreateRequest{BundleID: id, BundleVersion: 1, Items: []string{"cmd0"}})
	assert.Equal(t, "bundle_source_app", errorCode(res), code)

	bundles.bundles[id].RemovedAt = null.TimeFrom(time.Now())
	assert.NoError(t, get("user"))
	assert.Error(t, get("other_user"))
}

func TestBundleListingSubmit(t *testing.T) {
	h, bundles := bundlesHandler(1)
	create := handler.TypedWithBody(h.HandleBundleCreate)
	submit := handler.TypedWithBody(h.HandleBundleListingSubmit)

	code, res := callAs(t, create, "user", "app", "", wire.BundleCreateRequest{Name: "Mod", BundleExportRequest: wire.BundleExportRequest{CommandIDs: []string{"cmd0"}}})
	require.Equal(t, http.StatusOK, code, res)
	id := res["data"].(map[string]any)["id"].(string)

	listing := wire.BundleListing{Icon: "gavel", Category: "moderation", About: []string{"Bans members."}}
	bundles.bundles[id].LatestVersion = 3
	code, res = callAs(t, submit, "other_user", "app", id, listing)
	assert.Equal(t, "unknown_bundle", errorCode(res), code)
	code, res = callAs(t, submit, "user", "app", id, wire.BundleListing{Icon: "Gavel!", Category: "nope"})
	assert.Equal(t, http.StatusBadRequest, code, res)

	code, res = callAs(t, submit, "user", "app", id, listing)
	require.Equal(t, http.StatusOK, code, res)
	require.Len(t, bundles.reviews, 1)
	review := bundles.reviews[0]
	// The newest version is reviewed, also when an older one is listed.
	assert.Equal(t, 3, review.Version)
	assert.Equal(t, model.BundleReviewStatusPending, review.Status)
	assert.Equal(t, []string{"Bans members."}, review.Listing.About)
}

func TestRemovedBundleCantBeListedOrDeleted(t *testing.T) {
	h, bundles := bundlesHandler(1)
	create := handler.TypedWithBody(h.HandleBundleCreate)
	code, res := callAs(t, create, "user", "app", "", wire.BundleCreateRequest{Name: "Mod", BundleExportRequest: wire.BundleExportRequest{CommandIDs: []string{"cmd0"}}})
	require.Equal(t, http.StatusOK, code, res)
	id := res["data"].(map[string]any)["id"].(string)
	bundles.bundles[id].RemovedAt = null.TimeFrom(time.Now())

	code, res = callAs(t, handler.TypedWithBody(h.HandleBundleListingSubmit), "user", "app", id, wire.BundleListing{Icon: "gavel", Category: "moderation", About: []string{"Bans."}})
	assert.Equal(t, "bundle_removed", errorCode(res), code)

	// Deleting it would also delete the takedown and its reports.
	code, res = callAs(t, handler.Typed(h.HandleBundleDelete), "user", "app", id, nil)
	assert.Equal(t, "bundle_removed", errorCode(res), code)
}

func TestSetInstallSource(t *testing.T) {
	owner := &model.User{ID: "owner", DisplayName: "Jack", DiscordUsername: "jack"}
	tests := map[string]struct {
		bundle    *model.Bundle
		official  bool
		yours     bool
		removed   bool
		available null.Int
	}{
		"deleted":  {bundle: nil},
		"official": {bundle: &model.Bundle{LatestVersion: 3}, official: true, available: null.IntFrom(3)},
		// Apps get the approved version while a newer one is in review, also
		// after the bundle left the marketplace.
		"listed": {
			bundle:    &model.Bundle{OwnerUserID: null.StringFrom("owner"), Owner: owner, LatestVersion: 5, ApprovedVersion: 4, Listed: true},
			available: null.IntFrom(4),
		},
		"unlisted": {
			bundle:    &model.Bundle{OwnerUserID: null.StringFrom("owner"), Owner: owner, LatestVersion: 5, ApprovedVersion: 4},
			available: null.IntFrom(4),
		},
		"of another user": {bundle: &model.Bundle{OwnerUserID: null.StringFrom("owner"), Owner: owner, LatestVersion: 2}, available: null.IntFrom(2)},
		"yours":           {bundle: &model.Bundle{OwnerUserID: null.StringFrom("user"), LatestVersion: 2}, yours: true, available: null.IntFrom(2)},
		"taken down": {
			bundle:  &model.Bundle{OwnerUserID: null.StringFrom("owner"), Owner: owner, LatestVersion: 2, RemovedAt: null.TimeFrom(time.Now())},
			removed: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			res := &wire.BundleInstall{}
			setInstallSource(res, tt.bundle, "user")
			assert.Equal(t, tt.official, res.Official)
			assert.Equal(t, tt.yours, res.Yours)
			assert.Equal(t, tt.removed, res.Removed)
			assert.Equal(t, tt.available, res.AvailableVersion)
			if tt.bundle != nil && tt.bundle.Owner != nil {
				assert.Equal(t, "jack", res.Owner.DiscordUsername)
			}
		})
	}
}
