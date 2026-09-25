package model

import "time"

// FlowAIPrompt is a prompt sent to the flow AI. Repairs of the edits it
// produced are added as rounds instead of counting as new prompts.
type FlowAIPrompt struct {
	ID        string
	AppID     string
	UserID    string
	Model     string
	Rounds    int
	Usage     FlowAIUsage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FlowAIUsage is the tokens used by one or more model calls.
type FlowAIUsage struct {
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
}
