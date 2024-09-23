package handler

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/sethvargo/go-limiter/memorystore"
)

func TypedWithBody[REQ any, RESP any](next func(c *Context, r REQ) (*RESP, error)) HandlerFunc {
	return func(c *Context) error {
		var v REQ
		if err := c.ParseBody(&v); err != nil {
			return err
		}

		if sanitizable, ok := interface{}(&v).(BodySanitize); ok {
			sanitizable.Sanitize()
		}

		if validatable, ok := interface{}(&v).(BodyValidate); ok {
			if err := validatable.Validate(); err != nil {
				return err
			}
		}

		resp, err := next(c, v)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, wire.APIResponse[*RESP]{
			Success: true,
			Data:    resp,
		})
	}
}

func Typed[RESP any](next func(c *Context) (*RESP, error)) HandlerFunc {
	return func(c *Context) error {
		resp, err := next(c)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, wire.APIResponse[*RESP]{
			Success: true,
			Data:    resp,
		})
	}
}

type BodyValidate interface {
	Validate() error
}

type BodySanitize interface {
	Sanitize()
}

func RateLimitByUser(tokens uint64, interval time.Duration) MiddlewareFunc {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   tokens,
		Interval: interval,
	})
	if err != nil {
		log.Fatal(err)
	}

	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			userID := c.Session.UserID

			tokens, remaining, reset, ok, err := store.Take(c.Context(), userID)
			if err != nil {
				slog.With("error", err).Error("failed to take rate limit token")
				return ErrInternal(fmt.Sprintf("failed to take rate limit token: %v", err))
			}

			resetTime := time.Unix(0, int64(reset)).UTC().Format(time.RFC1123)

			c.SetHeader("X-RateLimit-Limit", fmt.Sprint(tokens))
			c.SetHeader("X-RateLimit-Remaining", fmt.Sprint(remaining))
			c.SetHeader("X-RateLimit-Reset", resetTime)

			if !ok {
				c.SetHeader("Retry-After", resetTime)
				return ErrRateLimit("Rate limit exceeded, try again later")
			}

			return next(c)
		}
	}
}
