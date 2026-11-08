package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRobloxUserGet:   executeActionRobloxUserGet,
		FlowNodeTypeActionBloxlinkUserGet: executeActionBloxlinkUserGet,
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

	storeRobloxUser(n, ctx, user)
	return n.ExecuteChildren(ctx)
}

// storeRobloxUser stores the user with their avatar as the block's result, or
// null if there's no user.
func storeRobloxUser(n *CompiledFlowNode, ctx *FlowContext, user *thing.RobloxUserValue) {
	if user == nil {
		ctx.StoreNodeResult(n, thing.Null)
		return
	}

	// The avatar is extra, flows that only use the rest shouldn't fail when
	// Roblox can't give it.
	var err error
	user.AvatarURL, err = ctx.Roblox.UserAvatarURL(ctx, user.ID)
	if err != nil {
		ctx.Log.CreateLogEntry(ctx, provider.LogLevelWarn, fmt.Sprintf("%s returned no avatar: %v", blockDefinitions[n.Type].Title, err))
	}
	ctx.StoreNodeResult(n, thing.NewRobloxUser(*user))
}

func executeActionBloxlinkUserGet(n *CompiledFlowNode, ctx *FlowContext) error {
	userID, err := evalSnowflake(ctx, n.Data.UserTarget, "user")
	if err != nil {
		return err
	}
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}
	if !guildID.IsValid() {
		return errors.New("needs a server, Bloxlink's API keys belong to one")
	}

	var res struct {
		RobloxID json.Number `json:"robloxID"`
	}
	path := fmt.Sprintf("/v4/public/guilds/%s/discord-to-roblox/%s", guildID, userID)
	err = integrationJSONRequest(ctx, integrations["bloxlink"], http.MethodGet, path, nil, &res)
	var integrationErr *integrationError
	switch {
	case errors.As(err, &integrationErr) && integrationErr.Status == http.StatusNotFound:
		// The member hasn't verified with Bloxlink.
	case err != nil:
		return err
	}

	var user *thing.RobloxUserValue
	if res.RobloxID != "" {
		id, err := res.RobloxID.Int64()
		if err != nil {
			return fmt.Errorf("Bloxlink returned an invalid Roblox ID %q", res.RobloxID)
		}
		user, err = ctx.Roblox.UserByID(ctx, id)
		if errors.Is(err, provider.ErrNotFound) {
			// The account was deleted. Its ID still tells the member apart
			// from one who didn't verify.
			user = &thing.RobloxUserValue{ID: id}
		} else if err != nil {
			return err
		}
	}

	storeRobloxUser(n, ctx, user)
	return n.ExecuteChildren(ctx)
}
