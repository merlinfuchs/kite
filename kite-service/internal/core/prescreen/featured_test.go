package prescreen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type featuredStore struct {
	store.BundleStore
	bundles []*model.Bundle
	// checked is the result by bundle ID and version.
	checked map[string]bool
}

func (s *featuredStore) UncheckedFeaturableBundles(context.Context, int) ([]*model.Bundle, error) {
	return s.bundles, nil
}

func (s *featuredStore) SetBundleFeaturable(_ context.Context, id string, version int, featurable bool) error {
	s.checked[fmt.Sprintf("%s@%d", id, version)] = featurable
	return nil
}

func listedBundle(id string, content string) *model.Bundle {
	return &model.Bundle{
		ID:              id,
		ApprovedVersion: 2,
		Listing:         &model.BundleListing{Name: "Tickets"},
		Content:         json.RawMessage(content),
	}
}

func TestChecksFeatured(t *testing.T) {
	client, inputs := fakeOpenAI(t, `{"reason": "A full ticket system.", "featurable": true}`)
	s := &featuredStore{bundles: []*model.Bundle{listedBundle("b1", `{}`)}, checked: map[string]bool{}}
	p := New(client, "gpt-6-luna", s, nil)

	p.checkFeatured(context.Background())
	if featurable, ok := s.checked["b1@2"]; !ok || !featurable {
		t.Fatalf("expected b1 at version 2 to be featurable, got %v", s.checked)
	}
	if len(*inputs) != 1 || !strings.Contains((*inputs)[0], `Tickets`) {
		t.Fatalf("unexpected inputs %v", *inputs)
	}
}

func TestDoesntFeatureTooBigOnes(t *testing.T) {
	client, inputs := fakeOpenAI(t, `{}`)
	big := fmt.Sprintf(`{"messages": [{"data": {"content": %q}}]}`, strings.Repeat("a", maxContentSize))
	s := &featuredStore{bundles: []*model.Bundle{listedBundle("b1", big)}, checked: map[string]bool{}}
	p := New(client, "gpt-6-luna", s, nil)

	p.checkFeatured(context.Background())
	if featurable, ok := s.checked["b1@2"]; !ok || featurable || len(*inputs) != 0 {
		t.Fatalf("expected b1 not to be featurable without asking the model, got %v", s.checked)
	}
}

func TestChecksFeaturedAgainAfterFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": {"message": "overloaded"}}`, http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client := openai.NewClient(option.WithAPIKey("test"), option.WithBaseURL(server.URL), option.WithMaxRetries(0))
	s := &featuredStore{bundles: []*model.Bundle{listedBundle("b1", `{}`), listedBundle("b2", `{}`)}, checked: map[string]bool{}}
	p := New(&client, "gpt-6-luna", s, nil)

	p.checkFeatured(context.Background())
	if len(s.checked) != 0 {
		t.Fatalf("expected nothing to be stored, got %v", s.checked)
	}
}
