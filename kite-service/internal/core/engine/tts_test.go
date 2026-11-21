package engine

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ttsTestFeatures struct {
	FeatureProvider
	features model.Features
}

func (f ttsTestFeatures) AppFeatures(ctx context.Context, appID string) (model.Features, error) {
	return f.features, nil
}

func newTestTTSProvider(t *testing.T, voiceAudio bool, handler http.HandlerFunc) *TTSProvider {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	p := NewTTSProvider("app", ttsTestFeatures{features: model.Features{VoiceAudio: voiceAudio}}, server.Client(), "sk-test")
	p.url = server.URL
	p.convert = func(ctx context.Context, audio io.Reader) ([]byte, error) {
		content, err := io.ReadAll(audio)
		return append([]byte("ogg of "), content...), err
	}
	return p
}

func TestTTSProviderSpeak(t *testing.T) {
	var got speechRequest
	var auth string
	p := newTestTTSProvider(t, true, func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		assert.Equal(t, http.MethodPost, r.Method)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("mp3 bytes"))
	})

	audio, err := p.Speak(context.Background(), provider.SpeakOpts{Text: "Hello", Voice: "nova", Speed: 1.25})
	require.NoError(t, err)

	// The speech is requested as MP3, which is converted to the sound.
	assert.Equal(t, "ogg of mp3 bytes", string(audio.Content))
	assert.Equal(t, "Bearer sk-test", auth)
	assert.Equal(t, speechRequest{Model: ttsModel, Input: "Hello", Voice: "nova", ResponseFormat: "mp3", Speed: 1.25}, got)
}

func TestTTSProviderSpeakNeedsPremium(t *testing.T) {
	called := false
	p := newTestTTSProvider(t, false, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	_, err := p.Speak(context.Background(), provider.SpeakOpts{Text: "Hello", Voice: "alloy"})
	assert.ErrorContains(t, err, "requires premium")
	assert.False(t, called, "no request is made, as it would cost money")
}

func TestTTSProviderSpeakErrors(t *testing.T) {
	tests := []struct {
		status int
		body   string
		want   string
	}{
		{http.StatusTooManyRequests, `{"error":{"message":"You exceeded your current quota"}}`, "busy"},
		{http.StatusBadRequest, `{"error":{"message":"bad input"}}`, "rejected the text"},
		{http.StatusInternalServerError, `oops`, "status 500"},
	}

	for _, tt := range tests {
		p := newTestTTSProvider(t, true, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.status)
			_, _ = w.Write([]byte(tt.body))
		})

		_, err := p.Speak(context.Background(), provider.SpeakOpts{Text: "Hello", Voice: "alloy"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), tt.want)
		// The account's quota isn't for the app owners to see.
		assert.NotContains(t, err.Error(), "quota")
	}
}

func TestTTSProviderSpeakTooLarge(t *testing.T) {
	p := newTestTTSProvider(t, true, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", ttsMaxAudioSize+1)))
	})

	_, err := p.Speak(context.Background(), provider.SpeakOpts{Text: "Hello", Voice: "alloy"})
	assert.ErrorContains(t, err, "too long")
}

func TestTTSProviderSpeakConvertFails(t *testing.T) {
	p := newTestTTSProvider(t, true, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("mp3 bytes"))
	})
	p.convert = func(ctx context.Context, audio io.Reader) ([]byte, error) {
		return nil, errors.New("ffmpeg isn't installed")
	}

	_, err := p.Speak(context.Background(), provider.SpeakOpts{Text: "Hello", Voice: "alloy"})
	assert.ErrorContains(t, err, "failed to convert the speech")
	assert.ErrorContains(t, err, "ffmpeg isn't installed")
}
