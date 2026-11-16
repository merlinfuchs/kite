package flow

// TestNoteCode tells the notes of test runs apart, so the editor can word
// them for users and offer to set a block's test result. The text of a note
// is written for MCP agents.
type TestNoteCode string

const (
	// The block calls a service other than Discord and wasn't run.
	TestNoteNotRun TestNoteCode = "not_run"
	// The block was mocked.
	TestNoteMocked TestNoteCode = "mocked"
	// AI Choice and AI Check, which aren't asked AI in test runs.
	TestNoteAIChoiceNotAsked TestNoteCode = "ai_choice_not_asked"
	TestNoteAICheckNotAsked  TestNoteCode = "ai_check_not_asked"
	// Verify Webhook, which takes the valid branch.
	TestNoteSignatureNotChecked TestNoteCode = "signature_not_checked"
	TestNoteWaitSkipped         TestNoteCode = "wait_skipped"
)

// TestNote is how a block ran differently than it would outside of a test.
type TestNote struct {
	Code TestNoteCode `json:"code"`
	Text string       `json:"text"`
}
