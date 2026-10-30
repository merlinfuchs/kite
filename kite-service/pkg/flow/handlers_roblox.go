package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
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
	userID, err := ctx.EvalTemplate(n.Data.UserTarget)
	if err != nil {
		return err
	}
	if !userID.Snowflake().IsValid() {
		return fmt.Errorf("must be a Discord user ID, got %q", userID.String())
	}

	guildID := discord.Snowflake(ctx.Data.GuildID())
	if n.Data.GuildTarget != "" {
		guild, err := ctx.EvalTemplate(n.Data.GuildTarget)
		if err != nil {
			return err
		}
		guildID = guild.Snowflake()
	}
	if !guildID.IsValid() {
		return errors.New("needs a server, Bloxlink's API keys belong to one")
	}

	path := fmt.Sprintf("/v4/public/guilds/%s/discord-to-roblox/%s", guildID, userID.Snowflake())
	body, err := integrationRequest(ctx, integrations["bloxlink"], http.MethodGet, path, nil)
	var integrationErr *integrationError
	if errors.As(err, &integrationErr) && integrationErr.Status == http.StatusNotFound {
		// The member hasn't verified with Bloxlink.
		storeRobloxUser(n, ctx, nil)
		return n.ExecuteChildren(ctx)
	}
	if err != nil {
		return err
	}

	var res struct {
		RobloxID json.Number `json:"robloxID"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to parse Bloxlink's response: %w", err)
	}

	var user *thing.RobloxUserValue
	if res.RobloxID != "" {
		id, err := res.RobloxID.Int64()
		if err != nil {
			return fmt.Errorf("Bloxlink returned an invalid Roblox ID %q", res.RobloxID)
		}
		user, err = ctx.Roblox.UserByID(ctx, id)
		if err != nil && !errors.Is(err, provider.ErrNotFound) {
			return err
		}
	}

	storeRobloxUser(n, ctx, user)
	return n.ExecuteChildren(ctx)
}
