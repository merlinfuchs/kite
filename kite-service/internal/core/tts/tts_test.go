package tts

import (
	"context"
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	text := strings.Repeat("hello world ", 40)
	text = strings.TrimSpace(text)

	chunks := Split(text, 50)
	if len(chunks) < 2 {
		t.Fatalf("expected several chunks, got %d", len(chunks))
	}
	for _, c := range chunks {
		if len([]rune(c)) > 50 {
			t.Errorf("chunk too long: %d", len([]rune(c)))
		}
	}
	if got := strings.Join(chunks, " "); got != text {
		t.Errorf("chunks don't add up to the text: %q", got)
	}

	// Text without spaces is cut at the size.
	chunks = Split(strings.Repeat("a", 120), 50)
	if len(chunks) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunks))
	}

	if chunks := Split("short", 50); len(chunks) != 1 || chunks[0] != "short" {
		t.Errorf("unexpected chunks: %v", chunks)
	}
}

func TestSpeakValidation(t *testing.T) {
	ctx := context.Background()

	if _, err := Speak(ctx, "   ", "en"); err != ErrEmptyText {
		t.Errorf("expected ErrEmptyText, got %v", err)
	}
	if _, err := Speak(ctx, strings.Repeat("a", MaxTextLength+1), "en"); err != ErrTooLong {
		t.Errorf("expected ErrTooLong, got %v", err)
	}
	if _, err := Speak(ctx, "hi", "en/../x"); err != ErrLanguage {
		t.Errorf("expected ErrLanguage, got %v", err)
	}
}
