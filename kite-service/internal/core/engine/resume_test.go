package engine

import (
	"context"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// A resume point ID arrives in an interaction's custom_id, which is chosen by
// whoever built the message rather than by us. The message instance branch of
// resumeFlowTarget reads from the database, so without app scoping on the
// lookup an app can resume another app's flow -- including its stored
// FlowState.

// The embedded interface is left nil so that any method resumeFlow does not
// call panics rather than silently returning a zero value, the same trick
// engine_test.go uses.
type stubResumePointStore struct {
	store.ResumePointStore

	point *model.ResumePoint
}

// Stands in for the query's `AND app_id = $2`: a point owned by another app is
// indistinguishable from one that does not exist.
func (s stubResumePointStore) ResumePoint(ctx context.Context, appID string, id string) (*model.ResumePoint, error) {
	if s.point.AppID != appID {
		return nil, store.ErrNotFound
	}
	return s.point, nil
}

func messageInstanceResumePoint(appID string) *model.ResumePoint {
	return &model.ResumePoint{
		ID:                "rp-1",
		Type:              model.ResumePointTypeMessageComponents,
		AppID:             appID,
		MessageID:         null.StringFrom("msg-1"),
		MessageInstanceID: null.IntFrom(1),
		FlowSourceID:      null.StringFrom("src-1"),
		FlowNodeID:        "node-1",
	}
}

func resumeApp(t *testing.T, resumePointAppID string) *App {
	t.Helper()

	// MessageInstanceStore is deliberately nil: reaching it is what the tests
	// below distinguish on, the same trick store_asset_test.go uses.
	return NewApp("app-1", Env{
		ResumePointStore: stubResumePointStore{
			point: messageInstanceResumePoint(resumePointAppID),
		},
	})
}

func TestResumeFlowIgnoresResumePointFromAnotherApp(t *testing.T) {
	app := resumeApp(t, "app-2")

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("a foreign resume point reached the message instance store: %v", r)
		}
	}()

	app.resumeFlow("rp-1", nil, nil)
}

// The negative control: with a matching app the same call has to get past the
// ownership check and reach the store, or the test above proves nothing.
func TestResumeFlowResolvesOwnResumePoint(t *testing.T) {
	app := resumeApp(t, "app-1")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected the app's own resume point to reach the message instance store")
		}
	}()

	app.resumeFlow("rp-1", nil, nil)
}
