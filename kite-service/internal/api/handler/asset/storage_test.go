package asset

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type fakeAssetStore struct {
	// Methods the tests don't use panic.
	store.AssetStore

	used    int64
	created bool
}

func (f *fakeAssetStore) AssetStorageByApp(ctx context.Context, appID string) (int64, error) {
	return f.used, nil
}

func (f *fakeAssetStore) CreateAsset(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	f.created = true
	return asset, nil
}

// upload sends a 100 byte file to the handler and returns the status and the
// error code.
func upload(t *testing.T, h *AssetHandler) (int, string) {
	t.Helper()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", "image.png")
	if err != nil {
		t.Fatal(err)
	}
	part.Write(bytes.Repeat([]byte("a"), 100))
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	handler.APIHandler(func(c *handler.Context) error {
		c.App = &model.App{ID: "app"}
		c.Session = &model.Session{UserID: "user"}
		return handler.Typed(h.HandleAssetCreate)(c)
	}).ServeHTTP(rec, req)

	var res struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	return rec.Code, res.Error.Code
}

func TestAssetStorageLimit(t *testing.T) {
	config := AssetHandlerConfig{MaxAssetStoragePerApp: 1000}

	full := &fakeAssetStore{used: 950}
	if status, code := upload(t, NewAssetHandler(full, config)); status != http.StatusBadRequest || code != "resource_limit" {
		t.Errorf("over the limit: got %d %q, want 400 resource_limit", status, code)
	}
	if full.created {
		t.Error("an asset was created over the limit")
	}

	room := &fakeAssetStore{used: 900}
	if status, _ := upload(t, NewAssetHandler(room, config)); status != http.StatusOK || !room.created {
		t.Errorf("up to the limit: got %d, created %v, want 200 and created", status, room.created)
	}

	unlimited := &fakeAssetStore{used: 1 << 40}
	if status, _ := upload(t, NewAssetHandler(unlimited, AssetHandlerConfig{})); status != http.StatusOK || !unlimited.created {
		t.Errorf("no limit: got %d, created %v, want 200 and created", status, unlimited.created)
	}
}
