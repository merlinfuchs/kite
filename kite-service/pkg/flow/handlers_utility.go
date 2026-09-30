package flow

import (
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"golang.org/x/exp/rand"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRandomGenerate:     executeActionRandomGenerate,
		FlowNodeTypeActionExpressionEvaluate: executeActionExpressionEvaluate,
		FlowNodeTypeActionLog:                executeActionLog,
	})
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
