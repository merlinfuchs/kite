package handler

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	ristretto_store "github.com/eko/gocache/store/ristretto/v4"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/sethvargo/go-limiter/memorystore"
)

func ClusterIndexProvider(clusterCount int, clusterIndex int) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			c.SetHeader("X-Cluster-Count", fmt.Sprintf("%d", clusterCount))
			c.SetHeader("X-Cluster-Index", fmt.Sprintf("%d", clusterIndex))
			return next(c)
		}
	}
}

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
				return ErrRateLimit("You are making too many requests. Please try again later.")
			}

			return next(c)
		}
	}
}

func NewCacheManager(capacity int64) (*cache.Cache[cachedResponse], error) {
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: capacity * 2,
		MaxCost:     capacity,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ristretto cache: %w", err)
	}

	ristrettoStore := ristretto_store.NewRistretto(ristrettoCache)

	return cache.New[cachedResponse](ristrettoStore), nil
}

var cacheNotFound = store.NotFound{}

func CacheByUser(manager *cache.Cache[cachedResponse], expiration time.Duration) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			cacheKey := getUserCacheKey(c)

			cached, err := manager.Get(c.Context(), cacheKey)
			if err != nil && !cacheNotFound.Is(err) {
				return fmt.Errorf("failed to get cache: %w", err)
			}

			if err == nil {
				for key, values := range cached.Headers {
					for _, value := range values {
						c.SetHeader(key, value)
					}
				}

				return c.Send(http.StatusOK, cached.Body)
			}

			proxyWriter := &cacheProxyWriter{w: c.w}
			c.w = proxyWriter

			err = next(c)
			if err != nil {
				return err
			}

			if proxyWriter.statusCode == http.StatusOK {
				cacheValue := cachedResponse{
					Headers: proxyWriter.Header(),
					Body:    proxyWriter.body,
				}
				err = manager.Set(
					c.Context(),
					cacheKey,
					cacheValue,
					store.WithCost(1),
					store.WithExpiration(expiration),
				)
				if err != nil {
					return fmt.Errorf("failed to set cache: %w", err)
				}
			}

			return nil
		}
	}
}

func getUserCacheKey(c *Context) string {
	return fmt.Sprintf("%s:%s", c.r.URL.String(), c.Session.UserID)
}

type cacheProxyWriter struct {
	w http.ResponseWriter

	statusCode int
	body       []byte
}

func (w *cacheProxyWriter) Header() http.Header {
	return w.w.Header()
}

func (w *cacheProxyWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}

	w.body = append(w.body, b...)
	return w.w.Write(b)
}

func (w *cacheProxyWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.w.WriteHeader(statusCode)
}

type cachedResponse struct {
	Headers http.Header
	Body    []byte
}
