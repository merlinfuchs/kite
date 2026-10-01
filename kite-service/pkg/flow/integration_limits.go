package flow

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// maxRateLimitWait is how long a request waits for a rate limit before its
// block fails. Flows time out after 30 seconds.
const maxRateLimitWait = 10 * time.Second

// integrationLimits keeps requests to integrations with RateLimitHeaders
// within the limits the services report. Apps are pinned to one cluster, so
// it knows all requests of an app, while each cluster learns about limits
// shared by all apps on its own.
var integrationLimits = newRateLimits()

type rateLimits struct {
	mu sync.Mutex
	// When requests can be sent again, by integration for limits shared by
	// all apps, and by route for the others.
	until map[string]time.Time
	// Requests of a route run one at a time, so each knows the limits the
	// one before was told about.
	routes map[string]*routeLock
}

type routeLock struct {
	ch    chan struct{}
	users int
}

func newRateLimits() *rateLimits {
	return &rateLimits{
		until:  make(map[string]time.Time),
		routes: make(map[string]*routeLock),
	}
}

// acquire waits until a request to the route can be sent, and returns a
// function to call with its response, or nil if there's none. It fails
// without waiting if that's more than maxRateLimitWait away.
func (l *rateLimits) acquire(ctx context.Context, integration Integration, credential string, route string) (func(*http.Response), error) {
	routeKey := integration.ID + "\x00" + credential + "\x00" + route
	deadline := time.Now().Add(maxRateLimitWait)

	l.mu.Lock()
	lock, ok := l.routes[routeKey]
	if !ok {
		lock = &routeLock{ch: make(chan struct{}, 1)}
		l.routes[routeKey] = lock
	}
	lock.users++
	l.mu.Unlock()

	locked := false
	release := func() {
		if locked {
			<-lock.ch
		}
		l.mu.Lock()
		defer l.mu.Unlock()
		lock.users--
		if lock.users == 0 {
			delete(l.routes, routeKey)
			if l.until[routeKey].Before(time.Now()) {
				delete(l.until, routeKey)
			}
		}
	}

	timer := time.NewTimer(maxRateLimitWait)
	defer timer.Stop()
	select {
	case lock.ch <- struct{}{}:
		locked = true
	case <-timer.C:
		release()
		return nil, rateLimitedError(integration, maxRateLimitWait)
	case <-ctx.Done():
		release()
		return nil, ctx.Err()
	}

	l.mu.Lock()
	until := l.until[routeKey]
	if global := l.until[integration.ID]; global.After(until) {
		until = global
	}
	l.mu.Unlock()

	if until.After(deadline) {
		release()
		return nil, rateLimitedError(integration, time.Until(until))
	}
	if wait := time.Until(until); wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			release()
			return nil, ctx.Err()
		}
	}

	return func(resp *http.Response) {
		if resp != nil {
			l.record(integration.ID, routeKey, resp)
		}
		release()
	}, nil
}

// record remembers until when a response says not to send requests. Limits
// of a bucket other than "global" are the route's, and the others are shared
// by all apps, as is a 429 that doesn't name its bucket.
func (l *rateLimits) record(integrationID string, routeKey string, resp *http.Response) {
	var until time.Time
	if resp.StatusCode == http.StatusTooManyRequests {
		seconds, err := strconv.ParseFloat(resp.Header.Get("Retry-After"), 64)
		if err != nil || seconds < 0 || seconds > 24*60*60 {
			// Waits a while rather than retrying right away, which ER:LC
			// punishes.
			seconds = 60
		}
		until = time.Now().Add(time.Duration(seconds * float64(time.Second)))
	}
	if resp.Header.Get("X-RateLimit-Remaining") == "0" {
		reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64)
		if err == nil && time.Unix(reset, 0).After(until) {
			until = time.Unix(reset, 0)
		}
	}
	if until.IsZero() {
		return
	}

	key := integrationID
	if bucket := resp.Header.Get("X-RateLimit-Bucket"); bucket != "" && bucket != "global" {
		key = routeKey
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if until.After(l.until[key]) {
		l.until[key] = until
	}
}

func rateLimitedError(integration Integration, wait time.Duration) error {
	return fmt.Errorf("%s is rate limited, try again in %d seconds", integration.Name, int(math.Ceil(wait.Seconds())))
}
