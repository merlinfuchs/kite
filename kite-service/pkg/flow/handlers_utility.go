package flow

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRandomGenerate:     executeActionRandomGenerate,
		FlowNodeTypeActionExpressionEvaluate: executeActionExpressionEvaluate,
		FlowNodeTypeActionLog:                executeActionLog,
	})
}

func executeActionRandomGenerate(n *CompiledFlowNode, ctx *FlowContext) error {
	var res thing.Thing
	var err error
	switch n.Data.RandomMode {
	case "", RandomModeNumber:
		res, err = randomNumber(n, ctx)
	case RandomModeItem:
		res, err = randomItem(n, ctx)
	default:
		err = traceError(n, fmt.Errorf("unknown random mode %q", n.Data.RandomMode))
	}
	if err != nil {
		return err
	}

	ctx.StoreNodeResult(n, res)
	return n.ExecuteChildren(ctx)
}

func randomNumber(n *CompiledFlowNode, ctx *FlowContext) (thing.Thing, error) {
	min, err := ctx.EvalTemplate(n.Data.RandomMin)
	if err != nil {
		return thing.Null, traceError(n, err)
	}

	max, err := ctx.EvalTemplate(n.Data.RandomMax)
	if err != nil {
		return thing.Null, traceError(n, err)
	}

	minInt := int(min.Int())
	maxInt := int(max.Int())
	if maxInt <= 0 || minInt >= maxInt {
		return thing.Null, &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "random_generate_max must be greater than random_generate_min and greater than 0",
		}
	}

	return thing.NewInt(rand.IntN(maxInt-minInt) + minInt), nil
}

// randomItem picks one of the items, which are the lines of the text, or
// the items of a list placeholder with their types, e.g. a member.
func randomItem(n *CompiledFlowNode, ctx *FlowContext) (thing.Thing, error) {
	value, err := ctx.EvalTemplate(n.Data.RandomItems)
	if err != nil {
		return thing.Null, traceError(n, err)
	}

	items, ok := listItems(value)
	if !ok {
		for _, line := range strings.Split(value.String(), "\n") {
			if item := strings.TrimSpace(line); item != "" {
				items = append(items, thing.NewString(item))
			}
		}
	}
	if len(items) == 0 {
		return thing.Null, traceError(n, errors.New("no items to pick from"))
	}

	return items[rand.IntN(len(items))], nil
}

// listItems returns the items of a list, keeping their types, unlike
// apiListItems. Lists built in expressions, like {{['a', 'b']}}, aren't
// wrapped in a thing.
func listItems(value thing.Thing) ([]thing.Thing, bool) {
	if value.Type == thing.TypeArray {
		return value.Array(), true
	}
	if value.Type != thing.TypeAny {
		return nil, false
	}

	rv := reflect.ValueOf(value.Value)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return nil, false
	}
	items := make([]thing.Thing, rv.Len())
	for i := range items {
		items[i] = thing.NewGuessTypeWithFallback(rv.Index(i).Interface())
	}
	return items, true
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
