package share

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

const (
	codeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O, 1/I/L, avoids misreads
	codeLength  = 6
)

type ShareHandler struct {
	shareCodeStore store.ShareCodeStore
}

func NewShareHandler(shareCodeStore store.ShareCodeStore) *ShareHandler {
	return &ShareHandler{
		shareCodeStore: shareCodeStore,
	}
}

func (h *ShareHandler) HandleShareCodeCreate(c *handler.Context, req wire.ShareCodeCreateRequest) (*wire.ShareCodeCreateResponse, error) {
	var code string

	// Collisions are extremely unlikely at 32^6 possibilities, but retry a
	// few times rather than surfacing a 500 on the rare hit.
	for attempt := 0; attempt < 5; attempt++ {
		candidate, err := generateCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate share code: %w", err)
		}

		_, err = h.shareCodeStore.CreateShareCode(c.Context(), &model.ShareCode{
			Code:      candidate,
			Data:      req.Data,
			CreatedAt: time.Now().UTC(),
		})
		if err == nil {
			code = candidate
			break
		}
		if !isUniqueViolation(err) {
			return nil, fmt.Errorf("failed to create share code: %w", err)
		}
	}

	if code == "" {
		return nil, fmt.Errorf("failed to generate a unique share code after retries")
	}

	return &wire.ShareCodeCreateResponse{Code: code}, nil
}

func (h *ShareHandler) HandleShareCodeGet(c *handler.Context) (*wire.ShareCodeGetResponse, error) {
	code := c.Param("code")

	shareCode, err := h.shareCodeStore.ShareCode(c.Context(), code)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_share_code", "Share code not found")
		}
		return nil, fmt.Errorf("failed to get share code: %w", err)
	}

	return &wire.ShareCodeGetResponse{Data: shareCode.Data}, nil
}

func generateCode() (string, error) {
	b := make([]byte, codeLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, codeLength)
	for i, v := range b {
		out[i] = codeCharset[int(v)%len(codeCharset)]
	}
	return string(out), nil
}

// isUniqueViolation checks for a Postgres unique-constraint violation
// (SQLSTATE 23505). No existing helper for this exists elsewhere in the
// codebase to reuse, so this uses the standard pgx/v5 pattern directly.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
