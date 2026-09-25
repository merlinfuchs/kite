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
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/flowai"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
)

func main() {
	addr := flag.String("addr", "localhost:4455", "address to listen on")
	model := flag.String("model", "gpt-5-mini", "model for building")
	effort := flag.String("effort", "low", "reasoning effort for building")
	checkModel := flag.String("check-model", "gpt-5-nano", "model for checking prompts")
	checkEffort := flag.String("check-effort", "low", "reasoning effort for checking prompts")
	flag.Parse()

	opts := []option.RequestOption{option.WithAPIKey(os.Getenv("OPENAI_API_KEY"))}
	prefix := ""
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
		opts = []option.RequestOption{option.WithAPIKey(key), option.WithBaseURL("https://openrouter.ai/api/v1")}
		prefix = "openai/"
	}
	client := openai.NewClient(opts...)

	assistant := flowai.NewAssistant(&client, flowai.Config{
		ModelConfig: flowai.ModelConfig{Model: prefix + *model, ReasoningEffort: *effort, MaxOutputTokens: 16000},
		Check:       flowai.ModelConfig{Model: prefix + *checkModel, ReasoningEffort: *checkEffort, MaxOutputTokens: 2000},
	})

	// The editor's wire types, plus the tokens used and how long it took.
	http.HandleFunc("POST /chat", func(w http.ResponseWriter, r *http.Request) {
		var req wire.FlowAIChatRequest
		if !decode(w, r, &req) {
			return
		}

		messages := make([]flowai.Message, len(req.Messages))
		for i, m := range req.Messages {
			messages[i] = flowai.Message{Role: m.Role, Content: m.Content}
		}
		start := time.Now()
		res, err := assistant.Respond(r.Context(), flowai.Request{
			Flow:     req.Flow,
			Messages: messages,
			Issues:   req.Issues,
			UserID:   "eval",
		})
		if err != nil {
			fail(w, err)
			return
		}
		respond(w, map[string]any{
			"prompt_id": "eval",
			"message":   res.Message,
			"edits":     res.Edits,
			"issues":    res.Issues,
			"usage":     wire.FlowAIUsage{},
			"eval":      map[string]any{"model": *model, "tokens": res.Usage, "ms": time.Since(start).Milliseconds()},
		})
	})

	http.HandleFunc("POST /check", func(w http.ResponseWriter, r *http.Request) {
		var req wire.FlowAICheckRequest
		if !decode(w, r, &req) {
			return
		}

		start := time.Now()
		res, err := assistant.Check(r.Context(), flowai.CheckRequest{Flow: req.Flow, Prompt: req.Prompt, UserID: "eval"})
		if err != nil {
			fail(w, err)
			return
		}
		respond(w, map[string]any{
			"verdict":          res.Verdict,
			"message":          res.Message,
			"suggested_prompt": res.SuggestedPrompt,
			"fields":           res.Fields,
			"eval":             map[string]any{"model": *checkModel, "tokens": res.Usage, "ms": time.Since(start).Milliseconds()},
		})
	})

	log.Printf("Serving the flow AI with %s and %s on %s", strings.TrimPrefix(*model, prefix), *checkModel, *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		fail(w, err)
		return false
	}
	return true
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
