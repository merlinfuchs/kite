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
// shared by all apps on its own. arikawa's limiter does the same for Discord,
// but reads Discord's headers and keeps every bucket forever.
var integrationLimits = newRateLimits()

type rateLimits struct {
	mu sync.Mutex
	// When requests to an integration can be sent again, for limits shared by
	// all apps.
	global map[string]time.Time
	routes map[string]*routeLimit
}

// routeLimit is the limit of the requests with one credential to one
// endpoint. They run one at a time, so each knows the limit the one before
// was told about.
type routeLimit struct {
	lock  chan struct{}
	users int
	until time.Time
}

func newRateLimits() *rateLimits {
	return &rateLimits{
		global: make(map[string]time.Time),
		routes: make(map[string]*routeLimit),
	}
}

// acquire waits until a request to the route can be sent, and returns a
// function to call with its response, or nil if there's none. It fails
// without waiting if that's more than maxRateLimitWait away.
func (l *rateLimits) acquire(ctx context.Context, integration Integration, credential string, route string) (func(*http.Response), error) {
	if !integration.RateLimitHeaders {
		return func(*http.Response) {}, nil
	}

	key := integration.ID + "\x00" + credential + "\x00" + route
	l.mu.Lock()
	r, ok := l.routes[key]
	if !ok {
		// Routes whose limit ended while nobody used them are kept until
		// another route is added.
		for k, idle := range l.routes {
			if idle.users == 0 && idle.until.Before(time.Now()) {
				delete(l.routes, k)
			}
		}
		r = &routeLimit{lock: make(chan struct{}, 1)}
		l.routes[key] = r
	}
	r.users++
	l.mu.Unlock()

	deadline := time.Now().Add(maxRateLimitWait)
	waitCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	select {
	case r.lock <- struct{}{}:
	case <-waitCtx.Done():
		l.leave(key, r)
		return nil, waitError(ctx, integration, maxRateLimitWait)
	}

	l.mu.Lock()
	until := r.until
	if global := l.global[integration.ID]; global.After(until) {
		until = global
	}
	l.mu.Unlock()

	var err error
	if until.After(deadline) {
		err = rateLimitedError(integration, time.Until(until))
	} else if wait := time.Until(until); wait > 0 {
		select {
		case <-time.After(wait):
		case <-waitCtx.Done():
			err = waitError(ctx, integration, wait)
		}
	}
	if err != nil {
		<-r.lock
		l.leave(key, r)
		return nil, err
	}

	return func(resp *http.Response) {
		if resp != nil {
			l.record(integration.ID, r, resp)
		}
		<-r.lock
		l.leave(key, r)
	}, nil
}

// leave forgets the route once nobody uses it and its limit is over.
func (l *rateLimits) leave(key string, r *routeLimit) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r.users--
	if r.users == 0 && r.until.Before(time.Now()) {
		delete(l.routes, key)
	}
}

// record remembers until when a response says not to send requests, in
// ER:LC's format: limits of a bucket other than "global" are the route's,
// and the others are shared by all apps, as is a 429 that doesn't name its
// bucket.
func (l *rateLimits) record(integrationID string, r *routeLimit, resp *http.Response) {
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

	l.mu.Lock()
	defer l.mu.Unlock()
	if bucket := resp.Header.Get("X-RateLimit-Bucket"); bucket != "" && bucket != "global" {
		if until.After(r.until) {
			r.until = until
		}
		return
	}
	if until.After(l.global[integrationID]) {
		l.global[integrationID] = until
	}
}

// waitError is the error of a wait that ended early: the flow's, if it was
// cancelled, and otherwise that the limit is too far away.
func waitError(ctx context.Context, integration Integration, wait time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return rateLimitedError(integration, wait)
}

func rateLimitedError(integration Integration, wait time.Duration) error {
	return fmt.Errorf("%s is rate limited, try again in %d seconds", integration.Name, int(math.Ceil(wait.Seconds())))
}

// Do sends a request to the integration, after waiting for its rate limits
// if it has RateLimitHeaders, and records those of the response.
func (i Integration) Do(ctx context.Context, credential string, req *http.Request, send func(*http.Request) (*http.Response, error)) (*http.Response, error) {
	done, err := integrationLimits.acquire(ctx, i, credential, req.Method+" "+req.URL.Path)
	if err != nil {
		return nil, err
	}
	resp, err := send(req)
	if err != nil {
		done(nil)
		return nil, err
	}
	done(resp)
	return resp, nil
}
