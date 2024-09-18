package message

import (
	"bytes"
	"context"
	"fmt"

	"github.com/diamondburned/arikawa/v3/utils/sendpart"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

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
