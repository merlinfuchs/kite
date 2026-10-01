// Command evalserver serves the flow AI without a database or limits, so the
// editor's eval in kite-web/src/lib/flow/ai.eval.test.ts can run the real
// prompts against the models. It uses OpenRouter if OPENROUTER_API_KEY is set,
// and OpenAI with OPENAI_API_KEY otherwise.
package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/flowai"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func main() {
	addr := flag.String("addr", "localhost:4455", "address to listen on")
	modelName := flag.String("model", "gpt-5-mini", "model")
	effort := flag.String("effort", "low", "reasoning effort")
	flag.Parse()

	opts := []option.RequestOption{option.WithAPIKey(os.Getenv("OPENAI_API_KEY"))}
	prefix := ""
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		opts = []option.RequestOption{option.WithAPIKey(key), option.WithBaseURL("https://openrouter.ai/api/v1")}
		prefix = "openai/"
	}
	client := openai.NewClient(opts...)

	assistant := flowai.NewAssistant(&client, flowai.Config{
		Model:           prefix + *modelName,
		ReasoningEffort: *effort,
		MaxOutputTokens: 16000,
	})

	// The editor's wire types, plus the tokens used and how long it took.
	http.HandleFunc("POST /chat", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			wire.FlowAIChatRequest
			// The service loads them from the app, which the eval doesn't have.
			Variables []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Scoped bool   `json:"scoped"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			fail(w, err)
			return
		}

		variables := make([]*model.Variable, len(req.Variables))
		for i, v := range req.Variables {
			variables[i] = &model.Variable{ID: v.ID, Name: v.Name, Scoped: v.Scoped}
		}
		start := time.Now()
		res, err := assistant.Respond(r.Context(), flowai.Request{
			Flow:      req.Flow,
			Messages:  req.AssistantMessages(),
			Issues:    req.Issues,
			Variables: variables,
			UserID:    "eval",
		})
		if err != nil {
			fail(w, err)
			return
		}
		respond(w, struct {
			wire.FlowAIChatResponse
			Eval evalInfo `json:"eval"`
		}{
			FlowAIChatResponse: wire.FlowAIChatResponse{
				PromptID:    "eval",
				Message:     res.Message,
				BuildPrompt: res.BuildPrompt,
				Fields:      wire.FlowAIFieldsToWire(res.Fields),
				Edits:       res.Edits,
			},
			Eval: evalInfo{Model: *modelName, Tokens: res.Usage, MS: time.Since(start).Milliseconds()},
		})
	})

	log.Printf("Serving the flow AI with %s on %s", *modelName, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// evalInfo is added to responses for the eval's report.
type evalInfo struct {
	Model  string               `json:"model"`
	Tokens model.AssistantUsage `json:"tokens"`
	MS     int64                `json:"ms"`
}

func respond(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true, "data": data})
}

func fail(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"error":   map[string]any{"code": "eval_failed", "message": err.Error()},
	})
}
