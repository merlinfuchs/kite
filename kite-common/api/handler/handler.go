package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-common/api/wire"
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

func (g HandlerGroup) Handle(method string, path string, f HandlerFunc) {
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		f = g.middlewares[i](f)
	}

	pattern := fmt.Sprintf("%s %s%s", method, g.pathPrefix, path)
	pattern = strings.TrimRight(pattern, "/")

	g.mux.Handle(pattern, APIHandler(f))
}

func (g HandlerGroup) Get(path string, f HandlerFunc) {
	g.Handle("GET", path, f)
}

func (g HandlerGroup) Post(path string, f HandlerFunc) {
	g.Handle("POST", path, f)
}

func (g HandlerGroup) Put(path string, f HandlerFunc) {
	g.Handle("PUT", path, f)
}

func (g HandlerGroup) Delete(path string, f HandlerFunc) {
	g.Handle("DELETE", path, f)
}

func (g HandlerGroup) Patch(path string, f HandlerFunc) {
	g.Handle("PATCH", path, f)
}

func (g HandlerGroup) Group(pathPrefix string, middlewares ...MiddlewareFunc) HandlerGroup {
	return HandlerGroup{
		mux:         g.mux,
		pathPrefix:  g.pathPrefix + pathPrefix,
		middlewares: append(g.middlewares, middlewares...),
	}
}
