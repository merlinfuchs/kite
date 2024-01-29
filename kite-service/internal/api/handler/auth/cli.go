package auth

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

func (h *AuthHandler) HandleAuthCLIStart(c *fiber.Ctx) error {
	loginCode, err := h.sessionManager.CreatePendingSession(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(wire.AuthCLIStartResponse{
		Success: true,
		Data: wire.AuthCLIStartResponseData{
			Code: loginCode,
		},
	})
}

func (h *AuthHandler) HandleAuthCLIRedirect(c *fiber.Ctx) error {
	loginCode := c.Query("code")
	if loginCode == "" {
		return helpers.BadRequest("missing_code", "Missing code query parameter")
	}

	setLoginCodeCookie(c, loginCode)

	state := setOauthStateCookie(c)
	return c.Redirect(h.cliOauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleAuthCLICallback(c *fiber.Ctx) error {
	state := getOauthStateCookie(c)
	if state == "" || c.Query("state") != state {
		return helpers.BadRequest("invalid_state", "Invalid oauth2 state, try again")
	}

	loginCode := getLoginCodeCookie(c)
	if loginCode == "" {
		return helpers.BadRequest("missing_login_code", "Missing login code cookie, try again")
	}

	accessToken, userID, guildIDs, err := h.ExchangeAccessToken(c.Context(), h.cliOauth2Config, c.Query("code"))
	if err != nil {
		return fmt.Errorf("Failed to exchange access token: %w", err)
	}

	token, err := h.sessionManager.CreateSession(c.Context(), model.SessionTypeCLI, userID, guildIDs, accessToken)
	if err != nil {
		return err
	}

	err = h.sessionManager.UpdatePendingSession(c.Context(), loginCode, token)
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_pending_session", "Pending session doesn't exist or has expired")
		}
		return err
	}

	return c.JSON(wire.AuthCLICallbackResponse{
		Success: true,
		Data: wire.AuthCLICallbackResponseData{
			Message: "Successfully logged in, continue in the CLI!",
		},
	})
}

func (h *AuthHandler) HandleAuthCLICheck(c *fiber.Ctx) error {
	loginCode := c.Query("code")
	if loginCode == "" {
		return helpers.BadRequest("missing_code", "Missing code query parameter")
	}

	pendingSession, err := h.sessionManager.GetPendingSession(c.Context(), loginCode)
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_pending_session", "Pending session doesn't exist or has expired")
		}
		return err
	}

	return c.JSON(wire.AuthCLICheckResponse{
		Success: true,
		Data: wire.AuthCLICheckResponseData{
			Pending: !pendingSession.Token.Valid,
			Token:   pendingSession.Token.String,
		},
	})
}

func getLoginCodeCookie(c *fiber.Ctx) string {
	state := c.Cookies("kite_login_code")
	c.ClearCookie("kite_login_code")
	return state
}

func setLoginCodeCookie(c *fiber.Ctx, code string) {
	c.Cookie(&fiber.Cookie{
		Name:     "kite_login_code",
		Value:    code,
		HTTPOnly: true,
		Secure:   true,
	})
}
