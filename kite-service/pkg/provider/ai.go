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
	SystemPrompt    string
	Prompt          string
	Tools           []AIToolType
	MaxOutputTokens int
}

type AIToolType string

const (
	AIToolTypeWebSearchPreview AIToolType = "web_search_preview"
)

type MockAIProvider struct{}

func (m *MockAIProvider) CreateResponse(ctx context.Context, opts CreateResponseOpts) (string, error) {
	return "", nil
}
