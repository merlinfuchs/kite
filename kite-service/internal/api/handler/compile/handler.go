package compile

import (
	"encoding/base64"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type CompileHandler struct{}

func NewHandler() *CompileHandler {
	return &CompileHandler{}
}
func (h *CompileHandler) HandleJSCompile(c *fiber.Ctx, req wire.CompileJSRequest) error {
	cmd := exec.Command("kitejs-compiler", "-", "-", "--optimize")
	cmd.Stdin = strings.NewReader(req.Source)

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	wasmData := base64.StdEncoding.EncodeToString(output)

	return c.JSON(wire.CompileJSResponse{
		Success: true,
		Data: wire.CompileJSResponseData{
			WASMBytes: wasmData,
		},
	})
}
