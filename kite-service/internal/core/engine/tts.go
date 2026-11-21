package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kitecloud/kite/kite-service/internal/core/sound"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

const (
	// openAISpeechURL is the endpoint that turns text into speech.
	openAISpeechURL = "https://api.openai.com/v1/audio/speech"
	// ttsModel is pinned, so a new version can't change how flows sound or
	// what they cost.
	ttsModel = "tts-1"
	// ttsMaxAudioSize bounds the audio read from the response, the longest
	// text of Speak Text is about 1 MB of MP3.
	ttsMaxAudioSize = 10 << 20
)

// TTSProvider speaks text with the OpenAI API, as the sounds the voice player
// plays.
type TTSProvider struct {
	appID    string
	features FeatureProvider
	client   *http.Client
	apiKey   string
	// url is the speech endpoint, replaced in tests.
	url string
	// convert turns the speech into a sound, replaced in tests which can't
	// rely on ffmpeg.
	convert func(ctx context.Context, audio io.Reader) ([]byte, error)
}

func NewTTSProvider(appID string, features FeatureProvider, client *http.Client, apiKey string) *TTSProvider {
	return &TTSProvider{
		appID:    appID,
		features: features,
		client:   client,
		apiKey:   apiKey,
		url:      openAISpeechURL,
		convert:  sound.Convert,
	}
}

type speechRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed,omitempty"`
}

func (p *TTSProvider) Speak(ctx context.Context, opts provider.SpeakOpts) (provider.TTSAudio, error) {
	// Speech is played in voice channels, which is a paid feature. Checked
	// first, as every request costs money.
	features, err := p.features.AppFeatures(ctx, p.appID)
	if err != nil {
		return provider.TTSAudio{}, fmt.Errorf("failed to get app features: %w", err)
	}
	if !features.VoiceAudio {
		return provider.TTSAudio{}, errors.New("speaking text requires premium")
	}

	body, err := json.Marshal(speechRequest{
		Model:          ttsModel,
		Input:          opts.Text,
		Voice:          opts.Voice,
		ResponseFormat: "mp3",
		Speed:          opts.Speed,
	})
	if err != nil {
		return provider.TTSAudio{}, fmt.Errorf("failed to encode speech request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.url, bytes.NewReader(body))
	if err != nil {
		return provider.TTSAudio{}, fmt.Errorf("failed to create speech request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := p.client.Do(req)
	if err != nil {
		return provider.TTSAudio{}, fmt.Errorf("failed to speak text: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return provider.TTSAudio{}, errors.New(speechError(res))
	}

	content, err := io.ReadAll(io.LimitReader(res.Body, ttsMaxAudioSize+1))
	if err != nil {
		return provider.TTSAudio{}, fmt.Errorf("failed to read speech: %w", err)
	}
	if len(content) > ttsMaxAudioSize {
		return provider.TTSAudio{}, fmt.Errorf("the speech is too long, shorten the text")
	}

	// Converted like uploaded sounds, so it plays the same way: Opus frames of
	// 20ms at an even volume.
	ogg, err := p.convert(ctx, bytes.NewReader(content))
	if err != nil {
		return provider.TTSAudio{}, fmt.Errorf("failed to convert the speech: %w", err)
	}

	return provider.TTSAudio{Content: ogg}, nil
}

// speechError describes a failed response by its status. The API's message is
// only logged, as it can tell about Kite's account, like its quota.
func speechError(res *http.Response) string {
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if err := json.Unmarshal(raw, &parsed); err == nil && strings.TrimSpace(parsed.Error.Message) != "" {
		slog.Warn("Speech request failed",
			slog.Int("status", res.StatusCode),
			slog.String("message", strings.TrimSpace(parsed.Error.Message)),
		)
	}

	switch {
	case res.StatusCode == http.StatusTooManyRequests:
		return "the speech service is busy, try again later"
	case res.StatusCode == http.StatusBadRequest:
		return "the speech service rejected the text"
	default:
		return fmt.Sprintf("the speech service failed with status %d", res.StatusCode)
	}
}
