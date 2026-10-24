package flow

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/exp/rand"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRandomID:           executeActionRandomID,
		FlowNodeTypeActionRandomGenerate:     executeActionRandomGenerate,
		FlowNodeTypeActionExpressionEvaluate: executeActionExpressionEvaluate,
		FlowNodeTypeActionLog:                executeActionLog,
	})
}

const maxRandomIDLength = 256

// Alphabets of the random ID types other than UUIDs, and their default
// lengths.
var randomIDFormats = map[RandomIDType]struct {
	alphabet      string
	defaultLength int
}{
	RandomIDTypeNanoID:       {"_-0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 21},
	RandomIDTypeAlphanumeric: {"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz", 16},
	RandomIDTypeNumeric:      {"0123456789", 6},
	RandomIDTypeHex:          {"0123456789abcdef", 32},
}

func executeActionRandomID(n *CompiledFlowNode, ctx *FlowContext) error {
	idType := n.Data.RandomIDType
	if idType == "" || idType == RandomIDTypeUUIDv4 {
		ctx.StoreNodeResult(n, thing.NewString(uuid.NewString()))
		return n.ExecuteChildren(ctx)
	}

	format, ok := randomIDFormats[idType]
	if !ok {
		return traceError(n, fmt.Errorf("unknown random ID type %q", idType))
	}

	length := format.defaultLength
	if n.Data.RandomIDLength != "" {
		raw, err := ctx.EvalTemplate(n.Data.RandomIDLength)
		if err != nil {
			return traceError(n, err)
		}

		length, err = parseRandomIDLength(raw.String())
		if err != nil {
			return traceError(n, err)
		}
	}

	id, err := gonanoid.Generate(format.alphabet, length)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, thing.NewString(id))
	return n.ExecuteChildren(ctx)
}

// A length that's set but empty, invalid or out of range is an error rather
// than falling back to the default.
func parseRandomIDLength(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("random ID length is empty")
	}

	length, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("random ID length must be a whole number, got %q", raw)
	}
	if length < 1 || length > maxRandomIDLength {
		return 0, fmt.Errorf("random ID length must be between 1 and %d, got %d", maxRandomIDLength, length)
	}

	return length, nil
}

func executeActionRandomGenerate(n *CompiledFlowNode, ctx *FlowContext) error {
	min, err := ctx.EvalTemplate(n.Data.RandomMin)
	if err != nil {
		return traceError(n, err)
	}

	max, err := ctx.EvalTemplate(n.Data.RandomMax)
	if err != nil {
		return traceError(n, err)
	}

	minInt := int(min.Int())
	maxInt := int(max.Int())
	if maxInt <= 0 || minInt >= maxInt {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "random_generate_max must be greater than random_generate_min and greater than 0",
		}
	}

	ctx.StoreNodeResult(n, thing.NewInt(rand.Intn(maxInt-minInt)+minInt))
	return n.ExecuteChildren(ctx)
}

func executeActionExpressionEvaluate(n *CompiledFlowNode, ctx *FlowContext) error {
	expression, err := ctx.EvalTemplate(n.Data.Expression)
	if err != nil {
		return traceError(n, err)
	}

	res, err := eval.Eval(ctx, expression.String(), ctx.EvalCtx)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, res)
	return n.ExecuteChildren(ctx)
}

func executeActionLog(n *CompiledFlowNode, ctx *FlowContext) error {
	logMessage, err := ctx.EvalTemplate(n.Data.LogMessage)
	if err != nil {
		return traceError(n, err)
	}

	ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, logMessage.String())
	return n.ExecuteChildren(ctx)
}
