package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type Context struct {
	r *http.Request
	w http.ResponseWriter

	Session        *model.Session
	App            *model.App
	UserAppRole    model.AppCollaboratorRole
	Features       model.Features
	Command        *model.Command
	Variable       *model.Variable
	Message        *model.Message
	EventListener  *model.EventListener
	PluginInstance *model.PluginInstance
}

func (c *Context) Context() context.Context {
	return c.r.Context()
}

func (c *Context) Param(name string) string {
	return c.r.PathValue(name)
}

func (c *Context) Query(name string) string {
	return c.r.URL.Query().Get(name)
}

func (c *Context) QueryValues() url.Values {
	return c.r.URL.Query()
}

func (c *Context) Header(name string) string {
	return c.r.Header.Get(name)
}

func (c *Context) Headers() http.Header {
	return c.r.Header
}

func (c *Context) SetHeader(name, value string) {
	c.w.Header().Set(name, value)
}

// SetUntrustedContentHeaders keeps browsers from rendering a response built
// by users, like an asset, as a page of the API's origin.
func (c *Context) SetUntrustedContentHeaders() {
	c.SetHeader("X-Content-Type-Options", "nosniff")
	c.SetHeader("X-Frame-Options", "DENY")
	c.SetHeader("Content-Security-Policy", "default-src 'none'")
}

func (c *Context) Cookie(name string) string {
	cookie, err := c.r.Cookie(name)
	if err != nil {
		return ""
	}

	return cookie.Value
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}

func (c *Context) DeleteCookie(name string) {
	http.SetCookie(c.w, &http.Cookie{
		Name:   name,
		Value:  "",
		MaxAge: -1,
	})
}

func (c *Context) Redirect(url string, code int) {
	http.Redirect(c.w, c.r, url, code)
}

// MaxJSONBodySize bounds a JSON request body.
//
// TypedWithBody parses the body before the handler runs, and several routes
// parse before any access check, so without a bound an anonymous caller can
// make the server buffer an arbitrary amount of memory. Set well above the
// largest legitimate payload, which is a bulk import of flows.
const MaxJSONBodySize = 8 * 1024 * 1024

// LimitBody caps how much of the request body can be read, failing the read
// once the limit is passed rather than truncating it silently.
func (c *Context) LimitBody(n int64) {
	c.r.Body = http.MaxBytesReader(c.w, c.r.Body, n)
}

func (c *Context) ParseBody(v interface{}) error {
	ct := c.r.Header.Get("Content-Type")
	if ct != "application/json" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "application/json" {
			return ErrUnsupportedMediaType("unsupported_media_type", "Content-Type must be application/json")
		}
	}

	c.LimitBody(MaxJSONBodySize)

	decoder := json.NewDecoder(c.r.Body)
	if err := decoder.Decode(v); err != nil {
		return decodeJSONError(err)
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return ErrBodyTooLarge(MaxJSONBodySize)
		}
		return ErrBadRequest("invalid_json", "request body must only contain a single JSON value")
	}

	return nil
}

func decodeJSONError(err error) error {
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		return ErrBodyTooLarge(MaxJSONBodySize)
	}
	var invalidUnmarshalErr *json.InvalidUnmarshalError
	if errors.As(err, &invalidUnmarshalErr) {
		return err
	}
	if errors.Is(err, io.EOF) {
		return ErrBadRequest("invalid_json", "request body is empty")
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return ErrBadRequest("invalid_json", fmt.Sprintf("invalid JSON syntax at offset %d: %s", syntaxErr.Offset, syntaxErr.Error()))
	}
	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalErr) {
		if unmarshalErr.Field != "" {
			return ErrBadRequest("invalid_json", fmt.Sprintf("invalid value for field %q: expected %s", unmarshalErr.Field, unmarshalErr.Type))
		}
		return ErrBadRequest("invalid_json", fmt.Sprintf("invalid value: expected %s", unmarshalErr.Type))
	}
	return ErrBadRequest("invalid_json", fmt.Sprintf("invalid JSON: %s", err.Error()))
}

// ReadBody reads the whole request body, whatever its content type, and
// fails if it's larger than limit.
func (c *Context) ReadBody(limit int64) ([]byte, error) {
	c.LimitBody(limit)

	body, err := io.ReadAll(c.r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return nil, ErrBodyTooLarge(limit)
		}
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	return body, nil
}

func (c *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return c.r.FormFile(name)
}

func (c *Context) JSON(status int, v interface{}) error {
	c.w.Header().Set("Content-Type", "application/json")
	c.w.WriteHeader(status)

	if err := json.NewEncoder(c.w).Encode(v); err != nil {
		return fmt.Errorf("failed to encode response body: %w", err)
	}

	return nil
}

func (c *Context) Send(status int, body []byte) error {
	c.w.WriteHeader(status)

	if _, err := c.w.Write(body); err != nil {
		return fmt.Errorf("failed to write response body: %w", err)
	}

	return nil
}
