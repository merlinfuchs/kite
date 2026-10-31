package message

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

// evalPlaceholders fills the placeholders of a message that is sent from the
// dashboard and returns a copy of it. There is no interaction or event here,
// so only placeholders that work without one can be filled, like now() or
// channel.id. The others are left as they were written instead of failing
// the send, which is also how they were sent before. now() is sentAt, so
// editing a message keeps the times it was sent with.
func evalPlaceholders(ctx context.Context, data *message.MessageData, guildID string, channelID string, sentAt time.Time) (message.MessageData, error) {
	res := data.Copy()

	guild := &eval.SnowflakeEnv{ID: guildID}
	evalCtx := eval.NewContext(eval.Env{
		"guild":   guild,
		"server":  guild,
		"channel": &eval.SnowflakeEnv{ID: channelID},
		// Overrides the builtin, in UTC like it is with expr.Timezone.
		"now": func() time.Time { return sentAt.UTC() },
	})
	evalCtx.KeepUnresolved = true

	err := res.EachString(func(s *string) error {
		filled, err := eval.EvalTemplateToString(ctx, *s, evalCtx)
		if err != nil {
			return err
		}

		*s = filled
		return nil
	})
	if err != nil {
		// Placeholders that fail are kept, so this is a template that's too long.
		return message.MessageData{}, handler.ErrBadRequest("template_too_long", err.Error())
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
