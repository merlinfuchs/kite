package model

import "time"

// FlowAIPrompt is a prompt sent to the flow AI. Repairs of the edits it
// produced are added as rounds instead of counting as new prompts.
type FlowAIPrompt struct {
	ID     string
	AppID  string
	UserID string
	Model  string
	// Prompt is the user's message.
	Prompt string
	Rounds int
	// Edited is whether the AI changed the flow. Only prompts that did count
	// towards the plan's limit.
	Edited    bool
	Usage     FlowAIUsage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FlowAIPromptCount is how many prompts an app sent in some time.
type FlowAIPromptCount struct {
	Edited int
	Total  int
}

// FlowAIUsage is the tokens used by one or more model calls.
type FlowAIUsage struct {
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
}
