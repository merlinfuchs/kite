package prescreen

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

type fakeStore struct {
	store.BundleStore
	review    *model.PendingBundleReview
	prescreen *model.BundleReviewPrescreen
	decided   *model.BundleReview
}

func (s *fakeStore) PendingBundleReview(context.Context, string) (*model.PendingBundleReview, error) {
	return s.review, nil
}

func (s *fakeStore) SetBundleReviewPrescreen(_ context.Context, _ string, p model.BundleReviewPrescreen) error {
	s.prescreen = &p
	return nil
}

func (s *fakeStore) DecideBundleReview(_ context.Context, d *model.BundleReview) (*model.BundleReview, error) {
	s.decided = d
	return d, nil
}

// fakeOpenAI answers every request with output and counts them.
func fakeOpenAI(t *testing.T, output string) (*openai.Client, *[]string) {
	var inputs []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		inputs = append(inputs, string(body))
		text, _ := json.Marshal(output)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"id": "resp_1", "object": "response", "created_at": 0, "status": "completed", "model": "gpt-5-mini",
			"output": [{"type": "message", "id": "msg_1", "status": "completed", "role": "assistant",
				"content": [{"type": "output_text", "text": %s, "annotations": []}]}],
			"usage": {"input_tokens": 1000, "output_tokens": 200, "total_tokens": 1200}
		}`, text)
	}))
	t.Cleanup(server.Close)
	client := openai.NewClient(option.WithAPIKey("test"), option.WithBaseURL(server.URL))
	return &client, &inputs
}

func pendingReview(content string) *model.PendingBundleReview {
	return &model.PendingBundleReview{
		Review: &model.BundleReview{ID: "r1", BundleID: "b1", Version: 1, Listing: model.BundleListing{Name: "test"}},
		Bundle: &model.Bundle{ID: "b1", Name: "test", Content: json.RawMessage(content), Owner: &model.User{DiscordID: "1"}},
	}
}

func TestRejectsObviousOnes(t *testing.T) {
	client, inputs := fakeOpenAI(t, `{"summary": "A test.", "concerns": [], "rating": 5,
		"reject": true, "reject_reason": "test", "creator_message": "This looks like a test."}`)
	s := &fakeStore{review: pendingReview(`{"commands": [{"flow_source": {"nodes": [{"id": "a", "type": "entry_command", "position": {"x": 1, "y": 2}, "data": {}}]}}]}`)}
	p := New(client, "gpt-5-mini", s, nil)

	if err := p.screen(context.Background(), "r1"); err != nil {
		t.Fatal(err)
	}
	if s.prescreen.Verdict != model.BundleReviewPrescreenVerdictReject || s.prescreen.Rating.Int64 != 5 {
		t.Fatalf("unexpected pre-screen %+v", s.prescreen)
	}
	if s.decided == nil || s.decided.Status != model.BundleReviewStatusRejected || s.decided.ReviewerUserID.Valid {
		t.Fatalf("expected a rejection without reviewer, got %+v", s.decided)
	}
	if s.decided.Reason != "This looks like a test." {
		t.Fatalf("unexpected reason %q", s.decided.Reason)
	}
	// The model doesn't get where blocks are.
	if strings.Contains((*inputs)[0], "position") {
		t.Fatal("the input has the positions of blocks")
	}
}

func TestLeavesOtherRejectionsToAdmins(t *testing.T) {
	client, _ := fakeOpenAI(t, `{"summary": "Low effort.", "concerns": [{"severity": "medium", "text": "No replies."}],
		"rating": 20, "reject": true, "reject_reason": "none", "creator_message": "Add replies."}`)
	s := &fakeStore{review: pendingReview(`{}`)}
	p := New(client, "gpt-5-mini", s, nil)

	if err := p.screen(context.Background(), "r1"); err != nil {
		t.Fatal(err)
	}
	if s.prescreen.Verdict != model.BundleReviewPrescreenVerdictReview || len(s.prescreen.Concerns) != 1 {
		t.Fatalf("unexpected pre-screen %+v", s.prescreen)
	}
	if s.decided != nil {
		t.Fatal("expected no decision")
	}
}

func TestSkipsTooBigOnes(t *testing.T) {
	client, inputs := fakeOpenAI(t, `{}`)
	big := fmt.Sprintf(`{"messages": [{"data": {"content": %q}}]}`, strings.Repeat("a", maxContentSize))
	s := &fakeStore{review: pendingReview(big)}
	p := New(client, "gpt-5-mini", s, nil)

	if err := p.screen(context.Background(), "r1"); err != nil {
		t.Fatal(err)
	}
	if s.prescreen.Verdict != model.BundleReviewPrescreenVerdictSkipped || len(*inputs) != 0 {
		t.Fatalf("expected a skip without asking the model, got %+v", s.prescreen)
	}
}
