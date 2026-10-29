package flow

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRandomGenerate:     executeActionRandomGenerate,
		FlowNodeTypeActionExpressionEvaluate: executeActionExpressionEvaluate,
		FlowNodeTypeActionLog:                executeActionLog,
		FlowNodeTypeActionValueLookup:        executeActionValueLookup,
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
	case RandomModeItems:
		res, err = randomItems(n, ctx)
	default:
		err = fmt.Errorf("unknown random mode %q", n.Data.RandomMode)
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
		return thing.Null, err
	}

	max, err := ctx.EvalTemplate(n.Data.RandomMax)
	if err != nil {
		return thing.Null, err
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
		return thing.Null, err
	}

	items := listOrLines(value)
	if len(items) == 0 {
		return thing.Null, errors.New("no items to pick from")
	}

	return items[rand.IntN(len(items))], nil
}

// randomItems picks up to random_count items in random order, never the same
// entry of the list twice. With fewer items, it returns all of them shuffled,
// and an empty list without any, so a giveaway without entries isn't an error.
func randomItems(n *CompiledFlowNode, ctx *FlowContext) (thing.Thing, error) {
	countValue, err := ctx.EvalTemplate(n.Data.RandomCount)
	if err != nil {
		return thing.Null, err
	}

	count, err := strconv.Atoi(strings.TrimSpace(countValue.String()))
	if err != nil || count < 1 || count > maxRandomCount {
		return thing.Null, fmt.Errorf("the number of items to pick must be a whole number from 1 to %d", maxRandomCount)
	}

	value, err := ctx.EvalTemplate(n.Data.RandomItems)
	if err != nil {
		return thing.Null, err
	}

	// A copy, as the items can be the stored result of another block.
	items := slices.Clone(listOrLines(value))
	rand.Shuffle(len(items), func(i, j int) {
		items[i], items[j] = items[j], items[i]
	})
	return thing.NewArray(items[:min(count, len(items))]), nil
}

// listOrLines returns the items of a list placeholder, keeping their types,
// or else the lines of the text, trimmed and without empty ones. Settings
// that take a list of values, one per line, use it.
func listOrLines(value thing.Thing) []thing.Thing {
	if items, ok := listItems(value); ok {
		return items
	}

	var items []thing.Thing
	for line := range strings.SplitSeq(value.String(), "\n") {
		if item := strings.TrimSpace(line); item != "" {
			items = append(items, thing.NewString(item))
		}
	}
	return items
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
		return err
	}

	res, err := eval.Eval(ctx, expression.String(), ctx.EvalCtx)
	if err != nil {
		return err
	}

	ctx.StoreNodeResult(n, res)
	return n.ExecuteChildren(ctx)
}

func executeActionValueLookup(n *CompiledFlowNode, ctx *FlowContext) error {
	lookupValue, err := ctx.EvalTemplate(n.Data.LookupValue)
	if err != nil {
		return err
	}

	value := lookupValue.String()
	if n.Data.LookupIgnoreCase {
		value = strings.ToLower(value)
	}

	// Only the value of the row that matches is evaluated, the first one wins.
	// Without a match or default the result is empty, which a condition can
	// check.
	result := n.Data.LookupDefault
	for _, row := range n.Data.LookupRows {
		rowKey, err := ctx.EvalTemplate(row.Key)
		if err != nil {
			return err
		}

		key := rowKey.String()
		// A key that is empty after placeholders are filled in would match
		// every value with contains and starts with.
		if key == "" {
			continue
		}
		if n.Data.LookupIgnoreCase {
			key = strings.ToLower(key)
		}

		var matches bool
		switch n.Data.LookupMode {
		case ComparsionModeContains:
			matches = strings.Contains(value, key)
		case ComparsionModeStartsWith:
			matches = strings.HasPrefix(value, key)
		default:
			matches = value == key
		}
		if matches {
			result = row.Value
			break
		}
	}

	res, err := ctx.EvalTemplate(result)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, thing.NewString(res.String()))
	return n.ExecuteChildren(ctx)
}

func executeActionLog(n *CompiledFlowNode, ctx *FlowContext) error {
	logMessage, err := ctx.EvalTemplate(n.Data.LogMessage)
	if err != nil {
		return err
	}

	ctx.Log.CreateLogEntry(ctx, n.Data.LogLevel, logMessage.String())
	return n.ExecuteChildren(ctx)
}
