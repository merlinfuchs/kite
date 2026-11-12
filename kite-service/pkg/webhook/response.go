package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ResponseTimeout is how long a request waits for its flow to respond.
// Senders give up after a few seconds, GitHub after 10.
const ResponseTimeout = 10 * time.Second

// ErrNotWaiting means the request was answered already, because the flow
// responded before, ended, waited past the timeout or continues after a sleep.
var ErrNotWaiting = fmt.Errorf("the webhook request was already answered, flows can respond once and within %s", ResponseTimeout)

// ResponseContentTypes are the content types a flow can respond with, and the
// header they're sent as. The response comes from the origin of the API, so
// anything a browser renders, like HTML or SVG, would run the script of
// whoever built the flow there.
var ResponseContentTypes = map[string]string{
	"application/json": "application/json",
	"text/plain":       "text/plain; charset=utf-8",
}

// Response is what a flow answers the request that triggered it with.
type Response struct {
	Status      int
	ContentType string
	Body        string
}

func (r *Response) Validate() error {
	// Redirects and other statuses with a meaning beyond the response itself
	// aren't allowed, only success and errors.
	if !(r.Status >= 200 && r.Status <= 299) && !(r.Status >= 400 && r.Status <= 599) {
		return fmt.Errorf("status must be between 200 and 299 or between 400 and 599, got %d", r.Status)
	}
	if (r.Status == 204 || r.Status == 205) && r.Body != "" {
		return fmt.Errorf("status %d can't have a body", r.Status)
	}
	if _, ok := ResponseContentTypes[r.ContentType]; !ok {
		return fmt.Errorf("unsupported content type %q", r.ContentType)
	}
	if r.ContentType == "application/json" && r.Body != "" && !json.Valid([]byte(r.Body)) {
		return errors.New("body isn't valid JSON")
	}
	if len(r.Body) > MaxBodySize {
		return fmt.Errorf("body exceeds the maximum size of %d bytes", MaxBodySize)
	}
	return nil
}

// Responder hands the response of a flow to the request that waits for it.
type Responder struct {
	once sync.Once
	done chan struct{}
	resp *Response
}

func NewResponder() *Responder {
	return &Responder{done: make(chan struct{})}
}

// Respond answers the request with resp. A nil responder belongs to an event
// restored after a sleep, whose request was answered long ago.
func (r *Responder) Respond(resp *Response) error {
	if r == nil {
		return ErrNotWaiting
	}
	if err := resp.Validate(); err != nil {
		return err
	}

	answered := false
	r.once.Do(func() {
		r.resp = resp
		answered = true
		close(r.done)
	})
	if !answered {
		return ErrNotWaiting
	}
	return nil
}

// Close answers the request with the default response, unless the flow
// responded already. It's called when the flow ends or suspends.
func (r *Responder) Close() {
	if r == nil {
		return
	}
	r.once.Do(func() { close(r.done) })
}

// Wait returns the response of the flow, or nil if the flow ends without one
// or doesn't respond within timeout.
func (r *Responder) Wait(ctx context.Context, timeout time.Duration) *Response {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case <-r.done:
	case <-ctx.Done():
		// Unless the flow responded just now.
		r.Close()
	}
	return r.resp
}
