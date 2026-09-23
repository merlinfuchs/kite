package engine

import (
	"context"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

// Embedded interface is nil so unexpected calls panic.
type stubResumePointStore struct {
	store.ResumePointStore

	point *model.ResumePoint
}

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

	// MessageInstanceStore is nil, reaching it means the ownership check passed.
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

func TestResumeFlowResolvesOwnResumePoint(t *testing.T) {
	app := resumeApp(t, "app-1")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected the app's own resume point to reach the message instance store")
		}
	}()

	app.resumeFlow("rp-1", nil, nil)
}
