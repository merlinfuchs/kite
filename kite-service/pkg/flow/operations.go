package flow

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
)

// operationsJSON has the operations of every integration's API that Kite
// knows about, and what calls of them from code cost and act on. Request
// blocks run one of them. It's generated from kite-web/src/lib/integrations,
// run `pnpm test -u` in kite-web to update it. See "Operations for code" in
// design/integrations.md.
//
//go:embed operations.json
var operationsJSON []byte

type APIParamType string

const (
	apiParamTypeSnowflake APIParamType = "snowflake"
	apiParamTypeInteger   APIParamType = "integer"
	apiParamTypeNumber    APIParamType = "number"
	apiParamTypeBoolean   APIParamType = "boolean"
	apiParamTypeArray     APIParamType = "array"
)

type APIParam struct {
	Name     string       `json:"name"`
	Type     APIParamType `json:"type"`
	Required bool         `json:"required"`
}

// Operation is an endpoint of an integration's API.
type Operation struct {
	Integration string     `json:"integration"`
	ID          string     `json:"id"`
	Method      string     `json:"method"`
	Path        string     `json:"path"`
	PathParams  []APIParam `json:"path_params"`
	QueryParams []APIParam `json:"query_params"`
	HasBody     bool       `json:"has_body"`
	// Interaction is set for Discord endpoints of the interaction the flow
	// runs with, whose {application_id} and {interaction_token} Kite fills in.
	Interaction bool `json:"interaction"`
	// FlowOnly is set for Discord endpoints that can't be called from
	// outside of flows, like by MCP clients.
	FlowOnly bool            `json:"flow_only"`
	Policy   OperationPolicy `json:"policy"`
}

// OperationPolicy is what a call of an operation from code costs and acts on.
// Blocks have their own credits.
type OperationPolicy struct {
	Credits int `json:"credits"`
	// What the call acts on: guild, channel, webhook, interaction, user or
	// none. It isn't checked yet, as every app has its own bot, which can act
	// in any server it's in. A shared bot would have to check it's the app's.
	Target string `json:"target"`
	// It deletes or removes something, or acts on many things at once.
	Destructive bool `json:"destructive"`
}

// operations are the operations by integration and ID.
var operations = func() map[string]map[string]Operation {
	var data struct {
		Operations []Operation `json:"operations"`
	}
	if err := json.Unmarshal(operationsJSON, &data); err != nil {
		panic(fmt.Sprintf("failed to parse operations.json: %v", err))
	}

	res := make(map[string]map[string]Operation)
	for _, op := range data.Operations {
		if res[op.Integration] == nil {
			res[op.Integration] = make(map[string]Operation)
		}
		res[op.Integration][op.ID] = op
	}
	return res
}()

// sendOperation sends a request to op with the path that operationPath
// built, with the app's bot token for Discord and the app's credential for
// other integrations. secrets are values in the body to redact from errors.
func (ctx *FlowContext) sendOperation(op Operation, path string, body []byte, reason api.AuditLogReason, secrets ...string) ([]byte, error) {
	integration, ok := integrations[op.Integration]
	if !ok {
		return nil, fmt.Errorf("unknown integration: %s", op.Integration)
	}

	if integration.Auth.Type == "discord_bot" {
		return ctx.Discord.APIRequest(ctx, provider.DiscordAPIRequest{
			Method: op.Method,
			Path:   path,
			Body:   body,
			Reason: reason,
		})
	}
	return integrationRequest(ctx, integration, op.Method, path, body, secrets...)
}
