package sharecode

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gopkg.in/guregu/null.v4"
)

const (
	// No 0/O or 1/I so codes survive being retyped.
	codeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	codeLength  = 8

	// Limits last_used_at writes for popular codes.
	touchInterval = 24 * time.Hour
)

type ShareCodeHandler struct {
	shareCodeStore store.ShareCodeStore
}

func NewShareCodeHandler(shareCodeStore store.ShareCodeStore) *ShareCodeHandler {
	return &ShareCodeHandler{
		shareCodeStore: shareCodeStore,
	}
}

func (h *ShareCodeHandler) HandleShareCodeCreate(c *handler.Context, req wire.ShareCodeCreateRequest) (*wire.ShareCodeCreateResponse, error) {
	code, err := gonanoid.Generate(codeCharset, codeLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate share code: %w", err)
	}

	now := time.Now().UTC()
	code, err = h.shareCodeStore.CreateShareCode(c.Context(), &model.ShareCode{
		Code:          code,
		Type:          model.ShareCodeType(req.Type),
		Data:          req.Data,
		CreatorUserID: c.Session.UserID,
		AppID:         null.StringFrom(c.App.ID),
		CreatedAt:     now,
		LastUsedAt:    now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create share code: %w", err)
	}

	return &wire.ShareCodeCreateResponse{Code: code}, nil
}

func (h *ShareCodeHandler) HandleShareCodeGet(c *handler.Context) (*wire.ShareCodeGetResponse, error) {
	code := strings.ToUpper(c.Param("code"))

	shareCode, err := h.shareCodeStore.ShareCode(c.Context(), code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_share_code", "Share code not found")
		}
		return nil, fmt.Errorf("failed to get share code: %w", err)
	}

	now := time.Now().UTC()
	if now.Sub(shareCode.LastUsedAt) >= touchInterval {
		if err := h.shareCodeStore.TouchShareCode(c.Context(), code, now); err != nil {
			return nil, fmt.Errorf("failed to touch share code: %w", err)
		}
	}

	return wire.ShareCodeToWire(shareCode), nil
}
