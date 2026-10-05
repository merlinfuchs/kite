package message

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
	"github.com/valyala/fasttemplate"
)

// evalPlaceholders fills the placeholders of a message that is sent from the
// dashboard and returns a copy of it. There is no interaction or event here,
// so only placeholders that work without one can be filled, like now() or
// channel.id. The others are left as they were written instead of failing
// the send, which is also how they were sent before.
func evalPlaceholders(ctx context.Context, data *message.MessageData, guildID string, channelID string) (message.MessageData, error) {
	res := data.Copy()

	guild := &eval.SnowflakeEnv{ID: guildID}
	evalCtx := eval.NewContext(eval.Env{
		"guild":   guild,
		"server":  guild,
		"channel": &eval.SnowflakeEnv{ID: channelID},
	})

	err := res.EachString(func(s *string) error {
		if s == nil || *s == "" {
			return nil
		}

		filled, err := fasttemplate.ExecuteFuncStringWithErr(*s, "{{", "}}", func(w io.Writer, tag string) (int, error) {
			val, err := eval.Eval(ctx, tag, evalCtx)
			if err != nil || val.IsNil() {
				return io.WriteString(w, "{{"+tag+"}}")
			}

			return fmt.Fprintf(w, "%v", val)
		})
		if err != nil {
			return err
		}

		*s = filled
		return nil
	})
	if err != nil {
		return message.MessageData{}, fmt.Errorf("failed to fill placeholders: %w", err)
	}

	return res, nil
}

func (h *MessageHandler) attachmentsToFiles(ctx context.Context, attachments []message.MessageAttachment) ([]sendpart.File, error) {
	res := make([]sendpart.File, 0, len(attachments))

	for _, attachment := range attachments {
		asset, err := h.assetStore.AssetWithContent(ctx, attachment.AssetID)
		if err != nil {
			return nil, fmt.Errorf("failed to get asset: %w", err)
		}

		res = append(res, sendpart.File{
			Name:   asset.Name,
			Reader: bytes.NewReader(asset.Content),
		})
	}

	return res, nil
}
