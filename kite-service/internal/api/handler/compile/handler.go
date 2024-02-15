package compile

import (
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type CompileHandler struct{}

func NewHandler() *CompileHandler {
	return &CompileHandler{}
}
func (h *CompileHandler) HandleCompile(c *fiber.Ctx, req wire.CompileRequest) error {
	var cmd *exec.Cmd
	switch req.Type {
	case "JS":
		cmd = exec.Command("kitejs-compiler", "-", "-", "--optimize")
		cmd.Stdin = strings.NewReader(req.Source)
	case "FLOW":
		cmd = exec.Command("kiteflow-compiler", "-", "-", "--optimize")
		cmd.Stdin = strings.NewReader(req.Source)
	default:
		return helpers.BadRequest("unknown_type", "Unsupported compiler type provided")
	}

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return c.JSON(wire.CompileResponse{
		Success: true,
		Data: wire.CompileResponseData{
			WASMBytes: output,
		},
	})
}
