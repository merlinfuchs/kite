package flow

import (
	"errors"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRobloxUserGet: executeActionRobloxUserGet,
	})
}

func executeActionRobloxUserGet(n *CompiledFlowNode, ctx *FlowContext) error {
	userID, err := ctx.EvalTemplate(n.Data.RobloxUserTarget)
	if err != nil {
		return traceError(n, err)
	}

	res := thing.Null

	switch n.Data.RobloxLookupMode {
	case RobloxLookupTypeID:
		user, err := ctx.Roblox.UserByID(ctx, userID.Int())
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}
		if user != nil {
			res = thing.NewRobloxUser(*user)
		}
	case RobloxLookupTypeName:
		users, err := ctx.Roblox.UsersByUsername(ctx, userID.String())
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return traceError(n, err)
		}
		if len(users) > 0 {
			res = thing.NewRobloxUser(users[0])
		}
	}

	ctx.StoreNodeResult(n, res)
	return n.ExecuteChildren(ctx)
}
