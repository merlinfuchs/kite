package session

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"gopkg.in/guregu/null.v4"
)

func (s *SessionManager) CreatePendingSession(ctx context.Context) (string, error) {
	code, err := generateLoginCode()
	if err != nil {
		return "", err
	}

	err = s.store.CreatePendingSession(ctx, &model.PendingSession{
		Code:      code,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Minute * 3),
	})
	if err != nil {
		return "", err
	}

	return code, nil
}

func (s *SessionManager) UpdatePendingSession(ctx context.Context, code string, token string) error {
	_, err := s.store.UpdatePendingSession(ctx, &model.PendingSession{
		Code:      code,
		Token:     null.NewString(token, true),
		ExpiresAt: time.Now().UTC().Add(time.Minute * 1),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *SessionManager) GetPendingSession(ctx context.Context, code string) (*model.PendingSession, error) {
	pendingSession, err := s.store.GetPendingSession(ctx, code)
	if err != nil {
		return nil, err
	}

	return pendingSession, nil
}

func generateLoginCode() (string, error) {
	b := make([]byte, 35)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate login code: %v", err)
	}

	token := base32.HexEncoding.EncodeToString(b)
	return token, nil
}
