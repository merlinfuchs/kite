package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionTopggVoteGet: executeActionTopggVoteGet,
	})
}

// executeActionTopggVoteGet checks whether a Discord user has an active vote
// for the listing of the app's Top.gg token. Top.gg answers 404 when the user
// never voted or can vote again, which isn't an error here.
func executeActionTopggVoteGet(n *CompiledFlowNode, ctx *FlowContext) error {
	userID, err := evalSnowflake(ctx, n.Data.UserTarget, "user")
	if err != nil {
		return err
	}

	integration, ok := integrations["topgg"]
	if !ok {
		return errors.New("the Top.gg integration doesn't exist")
	}

	res := map[string]any{
		"voted":      false,
		"created_at": "",
		"expires_at": "",
		"weight":     float64(0),
	}

	body, err := integrationRequest(ctx, integration, http.MethodGet, "/projects/@me/votes/"+userID.String()+"?source=discord", nil)
	var ierr *integrationError
	switch {
	case errors.As(err, &ierr) && ierr.Status == http.StatusNotFound:
	case err != nil:
		return err
	default:
		var vote struct {
			CreatedAt string  `json:"created_at"`
			ExpiresAt string  `json:"expires_at"`
			Weight    float64 `json:"weight"`
		}
		if err := json.Unmarshal(body, &vote); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
		res["voted"] = true
		res["created_at"] = vote.CreatedAt
		res["expires_at"] = vote.ExpiresAt
		res["weight"] = vote.Weight
	}

	ctx.StoreNodeResult(n, thing.NewFromJSONValue(res))
	return n.ExecuteChildren(ctx)
}
