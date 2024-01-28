package auth

import "github.com/gofiber/fiber/v2"

func (h *AuthHandler) HandleAuthCLIStart(c *fiber.Ctx) error {
	// generates a random token, put into database and return it to the cli
	return nil
}

func (h *AuthHandler) HandleAuthCLIRedirect(c *fiber.Ctx) error {
	// token is passed as query string, sets cookie and redirects to discord with random state
	return nil
}

func (h *AuthHandler) HandleAuthCLICallback(c *fiber.Ctx) error {
	// state is validated, token is read from cookie, session token is stored associated with the token
	return nil
}

func (h *AuthHandler) HandleAuthCLIGetResult(c *fiber.Ctx) error {
	// retrieves session tokeb by token, if not found cli refetches periodically
	return nil
}
