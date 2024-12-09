package placeholder

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/message"
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

func (e *Engine) FillMessage(ctx context.Context, msg *message.MessageData) error {
	var err error

	msg.Content, err = e.Fill(ctx, msg.Content)
	if err != nil {
		return err
	}

	for i := range msg.Embeds {
		embed := &msg.Embeds[i]

		embed.Title, err = e.Fill(ctx, embed.Title)
		if err != nil {
			return err
		}

		embed.Description, err = e.Fill(ctx, embed.Description)
		if err != nil {
			return err
		}

		embed.URL, err = e.Fill(ctx, embed.URL)
		if err != nil {
			return err
		}

		if embed.Author != nil {
			embed.Author.Name, err = e.Fill(ctx, embed.Author.Name)
			if err != nil {
				return err
			}

			embed.Author.IconURL, err = e.Fill(ctx, embed.Author.IconURL)
			if err != nil {
				return err
			}

			embed.Author.URL, err = e.Fill(ctx, embed.Author.URL)
			if err != nil {
				return err
			}
		}

		if embed.Footer != nil {
			embed.Footer.Text, err = e.Fill(ctx, embed.Footer.Text)
			if err != nil {
				return err
			}

			embed.Footer.IconURL, err = e.Fill(ctx, embed.Footer.IconURL)
			if err != nil {
				return err
			}
		}

		for _, field := range embed.Fields {
			field.Name, err = e.Fill(ctx, field.Name)
			if err != nil {
				return err
			}

			field.Value, err = e.Fill(ctx, field.Value)
			if err != nil {
				return err
			}
		}

		if embed.Image != nil {
			embed.Image.URL, err = e.Fill(ctx, embed.Image.URL)
			if err != nil {
				return err
			}
		}

		if embed.Thumbnail != nil {
			embed.Thumbnail.URL, err = e.Fill(ctx, embed.Thumbnail.URL)
			if err != nil {
				return err
			}
		}
	}

	for i := range msg.Components {
		row := &msg.Components[i]

		for j := range row.Components {
			component := &row.Components[j]

			component.Label, err = e.Fill(ctx, component.Label)
			if err != nil {
				return err
			}

			component.URL, err = e.Fill(ctx, component.URL)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
