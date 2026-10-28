package flow

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRobloxUserGet: executeActionRobloxUserGet,
	})
}

func executeActionRobloxUserGet(n *CompiledFlowNode, ctx *FlowContext) error {
	target, err := ctx.EvalTemplate(n.Data.RobloxUserTarget)
	if err != nil {
		return err
	}
	value := strings.TrimSpace(target.String())

	var user *thing.RobloxUserValue

	switch n.Data.RobloxLookupMode {
	case RobloxLookupTypeID:
		id, err := strconv.ParseInt(value, 10, 64)
		if target.Type == thing.TypeRobloxUser {
			id, err = target.RobloxUser().ID, nil
		}
		if err != nil || id <= 0 {
			return fmt.Errorf("must be a Roblox user ID, got %q", value)
		}
		user, err = ctx.Roblox.UserByID(ctx, id)
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return err
		}
	case RobloxLookupTypeName:
		if value == "" {
			return errors.New("username is empty")
		}
		users, err := ctx.Roblox.UsersByUsername(ctx, value)
		if err != nil {
			return err
		}
		if len(users) > 0 {
			// The username lookup leaves out the description, creation date
			// and ban status.
			user, err = ctx.Roblox.UserByID(ctx, users[0].ID)
			if errors.Is(err, provider.ErrNotFound) {
				user = &users[0]
			} else if err != nil {
				return err
			}
		}
	}

	res := thing.Null
	if user != nil {
		user.AvatarURL, err = ctx.Roblox.UserAvatarURL(ctx, user.ID)
		if err != nil {
			return err
		}
		res = thing.NewRobloxUser(*user)
	}

	ctx.StoreNodeResult(n, res)
	return n.ExecuteChildren(ctx)
}
