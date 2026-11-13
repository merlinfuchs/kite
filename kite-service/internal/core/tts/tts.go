// Package tts turns text into speech with Google Translate's text to speech
// endpoint, which needs no API key. Bots play the result in voice channels.
package tts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// MaxTextLength is how many characters can be spoken at once.
	MaxTextLength = 500
	// DefaultLanguage is used if no language is set.
	DefaultLanguage = "en"

	endpoint = "https://translate.google.com/translate_tts"
	// chunkLength is how much text the endpoint accepts in one request.
	chunkLength = 200
	// requestTimeout is how long fetching the speech of the whole text may take.
	requestTimeout = 15 * time.Second
	// maxResponseSize is how much audio one request may return.
	maxResponseSize = 2 << 20
)

var (
	ErrEmptyText = errors.New("there is no text to speak")
	ErrTooLong   = fmt.Errorf("the text can be at most %d characters long", MaxTextLength)
	ErrLanguage  = errors.New("the language must be a code like en, de or pt-BR")

	languagePattern = regexp.MustCompile(`^[a-zA-Z]{2,3}(-[a-zA-Z]{2,4})?$`)
	client          = &http.Client{Timeout: requestTimeout}
)

// Speak returns the speech of the text as MP3 audio. The language is a code
// like en, de or pt-BR, and defaults to English.
func Speak(ctx context.Context, text string, language string) ([]byte, error) {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return nil, ErrEmptyText
	}
	if len([]rune(text)) > MaxTextLength {
		return nil, ErrTooLong
	}

	language = strings.TrimSpace(language)
	if language == "" {
		language = DefaultLanguage
	}
	if !languagePattern.MatchString(language) {
		return nil, ErrLanguage
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// MP3 frames can simply be appended to each other.
	var audio bytes.Buffer
	for _, chunk := range Split(text, chunkLength) {
		part, err := fetch(ctx, chunk, language)
		if err != nil {
			return nil, err
		}
		audio.Write(part)
	}

	return audio.Bytes(), nil
}

func fetch(ctx context.Context, text string, language string) ([]byte, error) {
	query := url.Values{}
	query.Set("ie", "UTF-8")
	query.Set("client", "tw-ob")
	query.Set("tl", language)
	query.Set("q", text)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create text to speech request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Kite)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach the text to speech service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, errors.New("the text to speech service is rate limited right now, try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the text to speech service returned status %d", resp.StatusCode)
	}

	audio, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read the text to speech response: %w", err)
	}
	if len(audio) > maxResponseSize {
		return nil, errors.New("the text to speech response was too large")
	}
	if len(audio) == 0 {
		return nil, errors.New("the text to speech service returned no audio")
	}

	return audio, nil
}

// Split cuts text into chunks of at most size characters, at spaces where it
// can. Text without spaces is cut at size.
func Split(text string, size int) []string {
	var chunks []string

	runes := []rune(text)
	for len(runes) > size {
		cut := size
		for i := size; i > size/2; i-- {
			if runes[i] == ' ' {
				cut = i
				break
			}
		}

		if chunk := strings.TrimSpace(string(runes[:cut])); chunk != "" {
			chunks = append(chunks, chunk)
		}
		runes = runes[cut:]
	}

	if chunk := strings.TrimSpace(string(runes)); chunk != "" {
		chunks = append(chunks, chunk)
	}
	return chunks
}
