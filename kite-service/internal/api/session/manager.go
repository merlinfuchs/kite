package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

type Session struct {
	UserID      string
	GuildIDs    []string
	AccessToken string
	CreatedAt   time.Time
	ExpiresAt   time.Time
}

type SessionManager struct {
	store store.SessionStore
}

func New(store store.SessionStore) *SessionManager {
	return &SessionManager{
		store: store,
	}
}

func (s *SessionManager) GetSession(c *fiber.Ctx) (*Session, error) {
	token := c.Get("Authorization", c.Cookies("session_token"))
	if token == "" {
		return nil, nil
	}

	tokenHash, err := hashSessionToken(token)
	if err != nil {
		return nil, err
	}

	model, err := s.store.GetSession(c.Context(), tokenHash)
	if err != nil {
		if err == store.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &Session{
		UserID:      model.UserID,
		GuildIDs:    model.GuildIds,
		AccessToken: model.AccessToken,
		CreatedAt:   model.CreatedAt,
		ExpiresAt:   model.ExpiresAt,
	}, nil
}

func (s *SessionManager) CreateSession(ctx context.Context, userID string, guildIDs []string, accessToken string) (string, error) {
	token, err := generateSessionToken()
	if err != nil {
		return "", err
	}

	tokenHash, err := hashSessionToken(token)
	if err != nil {
		return "", err
	}

	err = s.store.CreateSession(ctx, &model.Session{
		TokenHash:   tokenHash,
		UserID:      userID,
		GuildIds:    guildIDs,
		AccessToken: accessToken,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *SessionManager) CreateSessionCookie(c *fiber.Ctx, userID string, guildIDs []string, accessToken string) error {
	token, err := s.CreateSession(c.Context(), userID, guildIDs, accessToken)
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    token,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "strict",
		Expires:  time.Now().UTC().Add(30 * 24 * time.Hour),
	})

	return nil
}

func (s *SessionManager) DeleteSession(c *fiber.Ctx) error {
	token := c.Cookies("session_token")
	if token == "" {
		return nil
	}

	c.ClearCookie("session_token")

	tokenHash, err := hashSessionToken(token)
	if err != nil {
		return err
	}

	err = s.store.DeleteSession(c.Context(), tokenHash)
	if err != nil {
		if err == store.ErrNotFound {
			return nil
		}
		return err
	}

	return nil
}

func generateSessionToken() (string, error) {
	b := make([]byte, 35)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session token: %v", err)
	}

	token := base32.HexEncoding.EncodeToString(b)
	return token, nil
}

func hashSessionToken(token string) (string, error) {
	b, err := base32.HexEncoding.DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("failed to decode token: %v", err)
	}
	tokenHashBytes := sha256.Sum256(b)
	tokenHash := base32.HexEncoding.EncodeToString(tokenHashBytes[:])

	return tokenHash, nil
}
