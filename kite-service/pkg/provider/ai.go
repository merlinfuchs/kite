package provider

import "context"

// AIProvider provides access to AI services.
type AIProvider interface {
	CreateChatCompletion(ctx context.Context, opts CreateChatCompletionOpts) (string, error)
}

type CreateChatCompletionOpts struct {
	Model               string
	SystemPrompt        string
	Prompt              string
	MaxCompletionTokens int
}

type MockAIProvider struct{}

func (m *MockAIProvider) CreateChatCompletion(ctx context.Context, opts CreateChatCompletionOpts) (string, error) {
	return "", nil
}
