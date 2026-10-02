package model

import "time"

// AssistantPrompt is a prompt sent to the dashboard's AI assistant, like the
// one that edits flows. Repairs of the edits it produced are added as rounds
// instead of counting as new prompts.
type AssistantPrompt struct {
	ID     string
	AppID  string
	UserID string
	Model  string
	// Prompt is the user's message.
	Prompt string
	Rounds int
	// Edited is whether the AI made changes. Only prompts that did count
	// towards the plan's limit.
	Edited    bool
	Usage     AssistantUsage
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AssistantPromptCount is how many prompts an app sent in some time.
type AssistantPromptCount struct {
	Edited int
	Total  int
}

// AssistantUsage is the tokens used by one or more model calls.
type AssistantUsage struct {
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
}
