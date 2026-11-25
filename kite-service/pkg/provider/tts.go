package provider

import (
	"context"
	"errors"
)

// ErrTTSUnavailable is returned by providers that can't speak text, like on
// a Kite service without an OpenAI API key.
var ErrTTSUnavailable = errors.New("text to speech isn't available")

// TTSProvider turns text into spoken audio.
type TTSProvider interface {
	Speak(ctx context.Context, opts SpeakOpts) (TTSAudio, error)
}

type SpeakOpts struct {
	Text string
	// Voice is one of the voices of the Speak Text block, like alloy.
	Voice string
	// Speed is how fast the text is spoken, 1 is normal speed. 0 means normal.
	Speed float64
}

type TTSAudio struct {
	// Content is the sound the voice player plays, an Ogg Opus file with 20ms
	// frames like the ones of uploaded sounds.
	Content []byte
}

// MockTTSProvider returns fixed bytes instead of a sound, for tests.
type MockTTSProvider struct{}

func (p *MockTTSProvider) Speak(ctx context.Context, opts SpeakOpts) (TTSAudio, error) {
	return TTSAudio{Content: []byte("audio")}, nil
}

// NoTTSProvider fails every request, for services that can't speak text.
type NoTTSProvider struct{}

func (p NoTTSProvider) Speak(ctx context.Context, opts SpeakOpts) (TTSAudio, error) {
	return TTSAudio{}, ErrTTSUnavailable
}
