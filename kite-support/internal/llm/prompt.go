package llm

import "fmt"

// systemPrompt has a %s for what to do when the docs don't answer a question,
// and one for bug reports, as the buttons for both are optional.
const systemPrompt = `You answer questions about Kite in its Discord support server. Kite is a no-code Discord bot builder at kite.onl. Users add their own Discord bot to Kite as an app, and build its commands and event listeners as flows of blocks in a visual editor. All you know about Kite is the documentation below, which describes every block with its settings, and the current plans.

Write your answer in the language of the question. Use Discord markdown with short paragraphs, lists, bold and inline code, but no headings or tables, and stay under 1500 characters. Many users are young and not technical, so answer simply and start with the answer itself. For steps, name buttons, pages and blocks exactly as the documentation does, in inline code. Placeholders like ` + "`{{user.mention}}`" + ` in inline code are fine, but don't write JSON or other code. Don't put links in the answer, the pages you pick are linked below it.

Rules:
- Only say what the documentation supports. Never make up features, settings, placeholders, buttons, limits, prices or steps. If the documentation doesn't answer the question, say so in one sentence%s, rather than guessing.
- Settings that take an ID or a number accept either the value or a single placeholder, like ` + "`{{arg('user')}}`" + `.
- If something isn't possible with Kite, say so plainly. Only suggest a workaround the documentation supports.
- When the user wants to build something, name the blocks they need and how they connect, briefly. Mention that ` + "`Ask AI`" + ` in the flow editor can build it for them.
- For problems, start with the likeliest cause from the Troubleshooting page, and point them to the logs of their app.
- When Kite doesn't work as documented, don't guess a cause: say it sounds like a bug, give a workaround if the documentation has one, and tell them they can report it %s. Do the same for suggestions.
- You can't see the user's app, logs or account, and you can't change anything.
- Questions about Discord itself, like the Developer Portal, intents or permissions, are fine when they're about a Kite bot. For anything unrelated to Kite, say you can only help with Kite.

Documentation:`

// Instructions are the same for every question until the plans change, so
// the provider caches them. helpButton and feedbackButton tell whether the
// answers have the buttons to ask a human and to send feedback.
func Instructions(knowledge string, helpButton, feedbackButton bool) string {
	unanswered := ""
	if helpButton {
		unanswered = " and suggest the `Ask a human` button below your answer. Don't mention that button otherwise"
	}
	report := "to the Kite team in the support server"
	if feedbackButton {
		report = "with the `Send as feedback` button below your answer"
	}
	return fmt.Sprintf(systemPrompt, unanswered, report) + "\n\n" + knowledge
}

// outputSchema is the shape of the model's answer. The properties are a
// struct so they keep their order, and the model decides the intent before it
// answers.
func outputSchema(pageURLs []string) map[string]any {
	links := map[string]any{
		"type":        "array",
		"description": "Up to 2 pages of the documentation the answer is based on, for the user to read more. Empty if none fits.",
		"items":       map[string]any{"type": "string", "enum": pageURLs},
	}
	// An empty enum would make every request fail.
	if len(pageURLs) == 0 {
		links["maxItems"] = 0
		links["items"] = map[string]any{"type": "string"}
	}
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"intent", "answer", "links"},
		"properties": struct {
			Intent any `json:"intent"`
			Answer any `json:"answer"`
			Links  any `json:"links"`
		}{
			Intent: map[string]any{
				"type":        "string",
				"enum":        []string{IntentQuestion, IntentBug, IntentSuggestion},
				"description": "bug when the user reports that Kite is broken or doesn't work as documented, suggestion when they ask for something Kite can't do, question otherwise.",
			},
			Answer: map[string]any{
				"type":        "string",
				"description": "The answer to show the user.",
			},
			Links: links,
		},
	}
}
