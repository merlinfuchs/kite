package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

type Context struct {
	r *http.Request
	w http.ResponseWriter

	Session       *model.Session
	App           *model.App
	UserAppRole   model.AppCollaboratorRole
	Command       *model.Command
	Variable      *model.Variable
	Message       *model.Message
	EventListener *model.EventListener
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

func (c *Context) Header(name string) string {
	return c.r.Header.Get(name)
}

func (c *Context) SetHeader(name, value string) {
	c.w.Header().Set(name, value)
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

func (c *Context) ParseBody(v interface{}) error {
	contentType := c.r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return fmt.Errorf("invalid content type: %s", contentType)
	}

	decoder := json.NewDecoder(c.r.Body)
	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to decode request body: %w", err)
	}

	return nil
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
