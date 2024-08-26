package placeholder

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/valyala/fasttemplate"
)

const startTag = "{{"
const endTag = "}}"

type Engine struct {
	providers map[string]Provider
}

func NewEngine() *Engine {
	return &Engine{
		providers: make(map[string]Provider),
	}
}

func (e *Engine) AddProvider(key string, provider Provider) {
	e.providers[key] = provider
}

func (e *Engine) Fill(ctx context.Context, input string) (string, error) {
	res, err := fasttemplate.ExecuteFuncStringWithErr(input, startTag, endTag, func(w io.Writer, tag string) (int, error) {
		keys := strings.Split(strings.TrimSpace(tag), ".")

		var provider Provider = e

		for _, key := range keys {
			var err error
			provider, err = provider.GetPlaceholder(ctx, key)
			if err != nil {
				if err == ErrNotFound {
					return 0, nil
				}
				return 0, fmt.Errorf("failed to get placeholder: %w", err)
			}
		}

		value, err := provider.ResolvePlaceholder(ctx)
		if err != nil {
			return 0, fmt.Errorf("failed to resolve placeholder value: %w", err)
		}

		return w.Write([]byte(value))
	})
	if err != nil {
		return "", fmt.Errorf("failed to fill placeholders: %w", err)
	}

	return res, nil
}

func (s Engine) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	provider, ok := s.providers[key]
	if !ok {
		return nil, ErrNotFound
	}
	return provider, nil
}

func (s Engine) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}
