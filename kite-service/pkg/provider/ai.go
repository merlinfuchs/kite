package provider

import (
	"context"
)

// AIProvider provides access to AI services.
type AIProvider interface {
	CreateResponse(ctx context.Context, opts CreateResponseOpts) (string, error)
}

type CreateResponseOpts struct {
	Model           string
	ReasoningEffort string
	SystemPrompt    string
	Prompt          string
	Tools           []AIToolType
	MaxToolCalls    int
	// MaxOutputTokens includes reasoning tokens.
	MaxOutputTokens int
}

type AIToolType string

const (
	AIToolTypeWebSearch AIToolType = "web_search"
)

type MockAIProvider struct{}

func (m *MockAIProvider) CreateResponse(ctx context.Context, opts CreateResponseOpts) (string, error) {
	return "", nil
}

// DecisionProvider asks a decision model, which answers a question with
// probabilities instead of text.
type DecisionProvider interface {
	Decide(ctx context.Context, opts DecideOpts) (Decision, error)
}

type DecisionType string

const (
	DecisionTypeChoice DecisionType = "choice"
	// A yes or no question, answered with the probability of yes.
	DecisionTypeNoul  DecisionType = "noul"
	DecisionTypeScore DecisionType = "score"
)

type DecideOpts struct {
	Model    string
	Type     DecisionType
	Question string
	// What the question is about, like a message.
	Input string
	// Guidance by option name for choices, or by "true" and "false" for yes
	// or no questions. Guidance can be empty.
	Options map[string]string
	// The levels of a score, lowest first.
	Levels []string
}

type Decision struct {
	// The picked option of a choice.
	Choice string
	// The probability of yes.
	Noul float64
	// The level of a score, from 0 for the first level. Can be in between
	// levels.
	Score      float64
	Confidence float64
}

type MockDecisionProvider struct{}

func (m *MockDecisionProvider) Decide(ctx context.Context, opts DecideOpts) (Decision, error) {
	return Decision{}, nil
}
