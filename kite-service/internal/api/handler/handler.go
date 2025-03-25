package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

type HandlerFunc = func(c *Context) error

func APIHandler(f HandlerFunc) http.Handler {
	wrapped := func(w http.ResponseWriter, r *http.Request) {
		ctx := &Context{
			r: r,
			w: w,
		}
		if err := f(ctx); err != nil {
			var aerr *Error
			// Check if the error is an API error, if not, convert it to one
			if !errors.As(err, &aerr) {
				// Special handling for validation errors
				var verr validation.Errors
				if errors.As(err, &verr) {
					aerr = &Error{
						Status:  http.StatusBadRequest,
						Code:    "validation_failed",
						Message: err.Error(),
						Data:    err,
					}
				} else {
					aerr = ErrInternal(err.Error())
				}
			}

			resp := wire.APIResponse[struct{}]{
				Success: false,
				Error: &wire.APIError{
					Code:    aerr.Code,
					Message: aerr.Message,
					Data:    aerr.Data,
				},
			}

			if err := ctx.JSON(aerr.Status, resp); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}

	return http.HandlerFunc(wrapped)
}

type MiddlewareFunc = func(HandlerFunc) HandlerFunc

type HandlerGroup struct {
	mux         *http.ServeMux
	pathPrefix  string
	middlewares []MiddlewareFunc
}

func Group(mux *http.ServeMux, pathPrefix string, middlewares ...MiddlewareFunc) HandlerGroup {
	return HandlerGroup{
		mux:         mux,
		pathPrefix:  pathPrefix,
		middlewares: middlewares,
	}
}

func (g HandlerGroup) Handle(method string, path string, f HandlerFunc, middlewares ...MiddlewareFunc) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		f = middlewares[i](f)
	}
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		f = g.middlewares[i](f)
	}

	pattern := fmt.Sprintf("%s %s%s", method, g.pathPrefix, path)
	pattern = strings.TrimRight(pattern, "/")

	g.mux.Handle(pattern, APIHandler(f))
}

func (g HandlerGroup) Get(path string, f HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Handle("GET", path, f, middlewares...)
}

func (g HandlerGroup) Post(path string, f HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Handle("POST", path, f, middlewares...)
}

func (g HandlerGroup) Put(path string, f HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Handle("PUT", path, f, middlewares...)
}

func (g HandlerGroup) Delete(path string, f HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Handle("DELETE", path, f, middlewares...)
}

func (g HandlerGroup) Patch(path string, f HandlerFunc, middlewares ...MiddlewareFunc) {
	g.Handle("PATCH", path, f, middlewares...)
}

func (g HandlerGroup) Group(pathPrefix string, middlewares ...MiddlewareFunc) HandlerGroup {
	combinedMiddleware := make([]MiddlewareFunc, len(g.middlewares)+len(middlewares))
	copy(combinedMiddleware, g.middlewares)
	copy(combinedMiddleware[len(g.middlewares):], middlewares)

	return HandlerGroup{
		mux:         g.mux,
		pathPrefix:  g.pathPrefix + pathPrefix,
		middlewares: combinedMiddleware,
	}
}
