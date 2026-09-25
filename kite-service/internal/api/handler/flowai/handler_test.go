package flowai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/core/flowai"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePromptStore struct {
	prompts map[string]*model.FlowAIPrompt
}

func (s *fakePromptStore) CreateFlowAIPrompt(ctx context.Context, prompt *model.FlowAIPrompt) error {
	s.prompts[prompt.ID] = prompt
	return nil
}

func (s *fakePromptStore) FlowAIPrompt(ctx context.Context, appID string, id string) (*model.FlowAIPrompt, error) {
	prompt, ok := s.prompts[id]
	if !ok || prompt.AppID != appID {
		return nil, store.ErrNotFound
	}
	return prompt, nil
}

func (s *fakePromptStore) DeleteFlowAIPrompt(ctx context.Context, appID string, id string) error {
	delete(s.prompts, id)
	return nil
}

func (s *fakePromptStore) StartFlowAIPromptRound(ctx context.Context, appID string, id string, maxRounds int, updatedAt time.Time) (bool, error) {
	prompt, err := s.FlowAIPrompt(ctx, appID, id)
	if err != nil || prompt.Rounds >= maxRounds {
		return false, err
	}
	prompt.Rounds++
	return true, nil
}

func (s *fakePromptStore) AddFlowAIPromptUsage(ctx context.Context, appID string, id string, usage model.FlowAIUsage, edited bool, updatedAt time.Time) error {
	prompt, err := s.FlowAIPrompt(ctx, appID, id)
	if err != nil {
		return err
	}
	prompt.Usage.InputTokens += usage.InputTokens
	prompt.Edited = prompt.Edited && edited
	return nil
}

func (s *fakePromptStore) CountFlowAIPromptsBetween(ctx context.Context, appID string, start time.Time, end time.Time) (model.FlowAIPromptCount, error) {
	var count model.FlowAIPromptCount
	for _, prompt := range s.prompts {
		if prompt.AppID == appID && !prompt.CreatedAt.Before(start) && !prompt.CreatedAt.After(end) {
			count.Total++
			if prompt.Edited {
				count.Edited++
			}
		}
	}
	return count, nil
}

type fakeAssistant struct {
	err error
	// answered makes the error come with usage, like when the model answered.
	answered bool
	// noEdits makes the answer a question without edits.
	noEdits bool
	calls   int
}

func (a *fakeAssistant) Model() string { return "gpt-5-mini" }

func (a *fakeAssistant) Respond(ctx context.Context, req flowai.Request) (*flowai.Response, error) {
	a.calls++
	if a.err != nil {
		if a.answered {
			return &flowai.Response{Usage: model.FlowAIUsage{InputTokens: 100}}, a.err
		}
		return nil, a.err
	}
	if a.noEdits {
		return &flowai.Response{Message: "Which channel?"}, nil
	}
	return &flowai.Response{
		Message: "Done.",
		Edits:   []map[string]any{{"op": "remove_node", "id": "a"}},
		Usage:   model.FlowAIUsage{InputTokens: 100},
	}, nil
}

type testSetup struct {
	store     *fakePromptStore
	assistant *fakeAssistant
	handler   *FlowAIHandler
}

func setup(assistant *fakeAssistant) *testSetup {
	s := &testSetup{
		store:     &fakePromptStore{prompts: map[string]*model.FlowAIPrompt{}},
		assistant: assistant,
	}
	s.handler = &FlowAIHandler{promptStore: s.store, maxRepairs: 2}
	if assistant != nil {
		s.handler.assistant = assistant
	}
	return s
}

// chat sends a chat request for app "app" with the given prompt limit and
// returns the status code and response body.
func (s *testSetup) chat(t *testing.T, limit int, body string) (int, map[string]any) {
	t.Helper()

	h := handler.APIHandler(func(c *handler.Context) error {
		c.Session = &model.Session{UserID: "user"}
		c.App = &model.App{ID: "app"}
		c.Features = model.Features{MaxAIPromptsPerMonth: limit}
		return handler.TypedWithBody(s.handler.HandleFlowAIChat)(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var res map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &res))
	return rec.Code, res
}

func errCode(res map[string]any) any {
	return res["error"].(map[string]any)["code"]
}

const prompt = `{"flow": "Blocks:", "messages": [{"role": "user", "content": "Remove a"}]}`

func repair(promptID string) string {
	return `{"flow": "Blocks:", "repair_prompt_id": "` + promptID + `",
		"messages": [{"role": "user", "content": "Remove a"}, {"role": "assistant", "content": "Done."}],
		"issues": ["'Log Message' never runs"]}`
}

func TestChatCountsPromptsAgainstTheLimit(t *testing.T) {
	s := setup(&fakeAssistant{})

	code, res := s.chat(t, 2, prompt)
	require.Equal(t, http.StatusOK, code, res)
	data := res["data"].(map[string]any)
	assert.Equal(t, "Done.", data["message"])
	usage := data["usage"].(map[string]any)
	assert.Equal(t, float64(1), usage["prompts_used"])
	assert.Equal(t, float64(2), usage["prompts_limit"])

	code, _ = s.chat(t, 2, prompt)
	require.Equal(t, http.StatusOK, code)

	code, res = s.chat(t, 2, prompt)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "resource_limit", errCode(res))
	assert.Equal(t, 2, s.assistant.calls)
}

func TestChatNeedsThePlanToIncludeIt(t *testing.T) {
	s := setup(&fakeAssistant{})

	code, res := s.chat(t, 0, prompt)
	assert.Equal(t, http.StatusForbidden, code)
	assert.Equal(t, "feature_unavailable", errCode(res))
}

func TestRepairsDontCountAndAreLimited(t *testing.T) {
	s := setup(&fakeAssistant{})

	_, res := s.chat(t, 1, prompt)
	promptID := res["data"].(map[string]any)["prompt_id"].(string)

	// The limit is used up, but repairs of the prompt still work.
	for range 2 {
		code, res := s.chat(t, 1, repair(promptID))
		require.Equal(t, http.StatusOK, code, res)
		assert.Equal(t, float64(1), res["data"].(map[string]any)["usage"].(map[string]any)["prompts_used"])
	}

	code, res := s.chat(t, 1, repair(promptID))
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "repair_limit", errCode(res))
	assert.Equal(t, 3, s.store.prompts[promptID].Rounds)
}

func TestRepairOfUnknownPrompt(t *testing.T) {
	s := setup(&fakeAssistant{})

	code, _ := s.chat(t, 1, repair("missing"))
	assert.Equal(t, http.StatusNotFound, code)
}

func TestUnusableAnswersCount(t *testing.T) {
	s := setup(&fakeAssistant{err: &flowai.ErrResponse{Message: "The AI's answer was cut off."}, answered: true})

	code, res := s.chat(t, 1, prompt)
	assert.Equal(t, http.StatusServiceUnavailable, code)
	assert.Equal(t, "The AI's answer was cut off.", res["error"].(map[string]any)["message"])
	require.Len(t, s.store.prompts, 1)
	for _, p := range s.store.prompts {
		assert.Equal(t, 100, p.Usage.InputTokens)
	}
}

func TestUnansweredPromptsDontCount(t *testing.T) {
	s := setup(&fakeAssistant{err: errors.New("connection refused")})

	code, res := s.chat(t, 1, prompt)
	assert.Equal(t, http.StatusServiceUnavailable, code)
	assert.Equal(t, "flow_ai_unavailable", errCode(res))
	assert.Empty(t, s.store.prompts)
}

func TestRepairsExpire(t *testing.T) {
	s := setup(&fakeAssistant{})

	_, res := s.chat(t, 1, prompt)
	promptID := res["data"].(map[string]any)["prompt_id"].(string)
	s.store.prompts[promptID].CreatedAt = time.Now().UTC().Add(-2 * time.Hour)

	code, res := s.chat(t, 1, repair(promptID))
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "repair_expired", errCode(res))
}

func TestChatWithoutOpenAI(t *testing.T) {
	s := setup(nil)

	code, res := s.chat(t, 1, prompt)
	assert.Equal(t, http.StatusServiceUnavailable, code)
	assert.Equal(t, "flow_ai_unavailable", errCode(res))
}

func TestChatValidatesTheRequest(t *testing.T) {
	s := setup(&fakeAssistant{})

	for _, body := range []string{
		`{"flow": "Blocks:", "messages": [{"role": "user", "content": ""}]}`,
		`{"flow": "Blocks:", "messages": [{"role": "assistant", "content": "Hi"}]}`,
		`{"flow": "Blocks:", "messages": [{"role": "user", "content": "Hi"}], "repair_prompt_id": "x"}`,
		`{"flow": "Blocks:", "messages": [{"role": "user", "content": "Hi"}], "issues": ["x"]}`,
		`{"flow": "Blocks:", "messages": [{"role": "user", "content": "Hi"}], "repair_prompt_id": "x", "issues": ["x"]}`,
	} {
		code, _ := s.chat(t, 1, body)
		assert.Equal(t, http.StatusBadRequest, code, body)
	}
	assert.Zero(t, s.assistant.calls)
}

func TestAnswersWithoutEditsDontCount(t *testing.T) {
	s := setup(&fakeAssistant{noEdits: true})

	code, res := s.chat(t, 1, prompt)
	require.Equal(t, http.StatusOK, code, res)
	assert.Equal(t, float64(0), res["data"].(map[string]any)["usage"].(map[string]any)["prompts_used"])
	for _, p := range s.store.prompts {
		assert.False(t, p.Edited)
		assert.Equal(t, "Remove a", p.Prompt)
	}

	// They are limited to 3 times the plan's prompts.
	for range 2 {
		code, _ = s.chat(t, 1, prompt)
		require.Equal(t, http.StatusOK, code)
	}
	code, res = s.chat(t, 1, prompt)
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "resource_limit", errCode(res))
}

func TestPromptsWithoutEditsCantBeRepaired(t *testing.T) {
	s := setup(&fakeAssistant{noEdits: true})

	_, res := s.chat(t, 1, prompt)
	promptID := res["data"].(map[string]any)["prompt_id"].(string)

	code, res := s.chat(t, 1, repair(promptID))
	assert.Equal(t, http.StatusBadRequest, code)
	assert.Equal(t, "nothing_to_repair", errCode(res))
}
