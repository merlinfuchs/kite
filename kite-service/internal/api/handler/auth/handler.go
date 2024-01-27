package auth

import "github.com/merlinfuchs/kite/kite-service/internal/api/session"

type AuthHandler struct {
	sessionManager *session.SessionManager
}

func New(sessionManager *session.SessionManager) *AuthHandler {
	return &AuthHandler{
		sessionManager: sessionManager,
	}
}
