package flow

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var limitedIntegration = Integration{ID: "limited", Name: "Limited", RateLimitHeaders: true}

func rateLimitResponse(status int, headers map[string]string) *http.Response {
	resp := &http.Response{StatusCode: status, Header: http.Header{}}
	for k, v := range headers {
		resp.Header.Set(k, v)
	}
	return resp
}

func acquireLimit(t *testing.T, l *rateLimits, credential string, route string) func(*http.Response) {
	t.Helper()
	done, err := l.acquire(context.Background(), limitedIntegration, credential, route)
	require.NoError(t, err)
	return done
}

func TestRateLimitWaitsForRetryAfter(t *testing.T) {
	l := newRateLimits()
	acquireLimit(t, l, "key", "POST /command")(rateLimitResponse(429, map[string]string{
		"Retry-After":        "0.2",
		"X-RateLimit-Bucket": "command-key",
	}))

	start := time.Now()
	acquireLimit(t, l, "key", "POST /command")(nil)
	assert.GreaterOrEqual(t, time.Since(start), 150*time.Millisecond)

	// Other servers have their own command limit.
	start = time.Now()
	acquireLimit(t, l, "other", "POST /command")(nil)
	assert.Less(t, time.Since(start), 100*time.Millisecond)
}

func TestRateLimitFailsWhenTooFarAway(t *testing.T) {
	l := newRateLimits()
	reset := time.Now().Add(30 * time.Second).Unix()
	acquireLimit(t, l, "key", "POST /command")(rateLimitResponse(200, map[string]string{
		"X-RateLimit-Bucket":    "command-key",
		"X-RateLimit-Remaining": "0",
		"X-RateLimit-Reset":     strconv.FormatInt(reset, 10),
	}))

	start := time.Now()
	_, err := l.acquire(context.Background(), limitedIntegration, "key", "POST /command")
	assert.ErrorContains(t, err, "Limited is rate limited, try again in")
	assert.Less(t, time.Since(start), 100*time.Millisecond)
}

// A global limit stops the requests of every app, as does a 429 that doesn't
// name its bucket.
func TestRateLimitGlobal(t *testing.T) {
	for _, bucket := range []string{"global", ""} {
		l := newRateLimits()
		acquireLimit(t, l, "key", "GET /server")(rateLimitResponse(429, map[string]string{
			"Retry-After":        "60",
			"X-RateLimit-Bucket": bucket,
		}))

		_, err := l.acquire(context.Background(), limitedIntegration, "other", "POST /command")
		assert.ErrorContains(t, err, "try again in 60 seconds", bucket)
	}
}

// Requests of a route run one at a time, so the next knows the limits of the
// one before.
func TestRateLimitOneRequestPerRoute(t *testing.T) {
	l := newRateLimits()
	done := acquireLimit(t, l, "key", "POST /command")

	acquired := make(chan error)
	go func() {
		done, err := l.acquire(context.Background(), limitedIntegration, "key", "POST /command")
		if err == nil {
			done(nil)
		}
		acquired <- err
	}()

	select {
	case <-acquired:
		t.Fatal("second request didn't wait for the first")
	case <-time.After(50 * time.Millisecond):
	}

	done(rateLimitResponse(429, map[string]string{"Retry-After": "0.1", "X-RateLimit-Bucket": "command-key"}))
	start := time.Now()
	require.NoError(t, <-acquired)
	assert.GreaterOrEqual(t, time.Since(start), 50*time.Millisecond)

	// The route's limit is over, so it's forgotten.
	l.mu.Lock()
	defer l.mu.Unlock()
	assert.Empty(t, l.routes)
}

func TestRateLimitIgnoresOtherResponses(t *testing.T) {
	l := newRateLimits()
	acquireLimit(t, l, "key", "GET /server")(rateLimitResponse(200, map[string]string{
		"X-RateLimit-Bucket":    "global",
		"X-RateLimit-Remaining": "34",
	}))
	assert.Empty(t, l.global)
	assert.Empty(t, l.routes)
}

func TestRateLimitOnlyWithHeaders(t *testing.T) {
	l := newRateLimits()
	done, err := l.acquire(context.Background(), Integration{ID: "unlimited"}, "key", "GET /server")
	require.NoError(t, err)
	done(rateLimitResponse(429, map[string]string{"Retry-After": "60"}))
	assert.Empty(t, l.global)
	assert.Empty(t, l.routes)
}

// Long blocks are waited out in full, up to a day, as retrying early makes
// them longer.
func TestRateLimitClampsRetryAfter(t *testing.T) {
	l := newRateLimits()
	acquireLimit(t, l, "key", "GET /server")(rateLimitResponse(429, map[string]string{
		"Retry-After":        "172800",
		"X-RateLimit-Bucket": "global",
	}))

	_, err := l.acquire(context.Background(), limitedIntegration, "key", "GET /server")
	assert.ErrorIs(t, err, ErrRateLimited)
	assert.ErrorContains(t, err, "try again in 86400 seconds")
}
