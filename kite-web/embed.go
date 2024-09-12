//go:build embedweb
// +build embedweb

package kiteweb

import (
	"embed"
	"io/fs"
	"net/http"

	gonextstatic "github.com/merlinfuchs/go-next-static"
)

//go:embed all:out
var OutFS embed.FS

func NewHandler() (http.Handler, error) {
	subFS, err := fs.Sub(OutFS, "out")
	if err != nil {
		return nil, err
	}

	handler, err := gonextstatic.NewHandler(subFS)
	if err != nil {
		return nil, err
	}

	return handler, nil
}
