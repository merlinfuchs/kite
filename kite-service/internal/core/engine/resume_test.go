package engine

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// A resume point ID arrives in an interaction's custom_id, which is chosen by
// whoever built the message rather than by us. The message instance branch of
// resumeFlowTarget reads from the database, so without an ownership check an
// app can resume another app's flow -- including its stored FlowState.

type stubResumePointStore struct {
	point *model.ResumePoint
}

func (s stubResumePointStore) ResumePoint(ctx context.Context, id string) (*model.ResumePoint, error) {
	return s.point, nil
}

func (s stubResumePointStore) CreateResumePoint(ctx context.Context, resumePoint *model.ResumePoint) error {
	return nil
}

func (s stubResumePointStore) DeleteResumePoint(ctx context.Context, id string) error {
	return nil
}

func (s stubResumePointStore) DeleteExpiredResumePoints(ctx context.Context, timestamp time.Time) error {
	return nil
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

var _ store.ResumePointStore = stubResumePointStore{}
