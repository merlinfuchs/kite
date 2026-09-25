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
	// ReasoningTokens is added to MaxOutputTokens, which only caps the answer.
	ReasoningTokens int
	SystemPrompt    string
	Prompt          string
	Tools           []AIToolType
	MaxToolCalls    int
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
