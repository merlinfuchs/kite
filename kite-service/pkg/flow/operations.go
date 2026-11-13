package flow

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// operationsJSON has the operations of every integration's API that Kite
// knows about, and what calls of them from code cost and act on. Request
// blocks run one of them. It's generated from kite-web/src/lib/integrations,
// run `pnpm test -u` in kite-web to update it. See "Operations for code" in
// design/integrations.md.
//
//go:embed operations.json
var operationsJSON []byte

type apiParamType string

const (
	apiParamTypeSnowflake apiParamType = "snowflake"
	apiParamTypeInteger   apiParamType = "integer"
	apiParamTypeNumber    apiParamType = "number"
	apiParamTypeBoolean   apiParamType = "boolean"
	apiParamTypeArray     apiParamType = "array"
)

type apiParam struct {
	Name     string       `json:"name"`
	Type     apiParamType `json:"type"`
	Required bool         `json:"required"`
}

// operation is an endpoint of an integration's API, or for kind "handler" Go
// code in operationHandlers, which has none of the request's fields.
type operation struct {
	Kind        string `json:"kind"`
	Integration string `json:"integration"`
	ID          string `json:"id"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	// Credits each request costs, from request blocks and from code.
	Credits     int        `json:"credits"`
	PathParams  []apiParam `json:"path_params"`
	QueryParams []apiParam `json:"query_params"`
	HasBody     bool       `json:"has_body"`
	// Interaction is set for Discord endpoints of the interaction the flow
	// runs with, whose {application_id} and {interaction_token} Kite fills in.
	Interaction bool `json:"interaction"`
	// FlowOnly is set for Discord endpoints that can't be called from
	// outside of flows, like by MCP clients.
	FlowOnly bool            `json:"flow_only"`
	Policy   operationPolicy `json:"policy"`
}

// operationPolicy is what a call of an operation from code acts on.
type operationPolicy struct {
	// Code can call it, see OperationPolicy in
	// kite-web/src/lib/integrations/operations.ts.
	Code bool `json:"code"`
	// What the call acts on, see OperationPolicy in
	// kite-web/src/lib/integrations/operations.ts. It isn't checked yet, as
	// every app has its own bot, which can act in any server it's in.
	Target string `json:"target"`
	// It deletes, removes or bans.
	Destructive bool `json:"destructive"`
	// Code can pass the audit log reason of the call as "reason".
	Reason bool `json:"reason"`
}

// operations are the operations by integration and ID.
var operations = func() map[string]map[string]operation {
	var data struct {
		Operations []operation `json:"operations"`
	}
	if err := json.Unmarshal(operationsJSON, &data); err != nil {
		panic(fmt.Sprintf("failed to parse operations.json: %v", err))
	}

	res := make(map[string]map[string]operation)
	for _, op := range data.Operations {
		if res[op.Integration] == nil {
			res[op.Integration] = make(map[string]operation)
		}
		res[op.Integration][op.ID] = op
	}
	return res
}()

// sendOperation sends a request to op with the path its caller built, with
// the app's bot token for Discord and the app's credential for other
// integrations. secrets are values in the body to redact from errors. It's
// only the sending, javaScriptHost runs an operation from its parameters with
// its policy.
func (ctx *FlowContext) sendOperation(op operation, path string, body []byte, reason api.AuditLogReason, secrets ...string) ([]byte, error) {
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

// operationRequest builds the path, body and audit log reason of a call of op
// from code with arg, a JSON object with its path and query parameters by
// name, its body under "body" and the audit log reason under "reason" if its
// policy allows one, like {"channel_id": "1", "body": {"content": "Hi"}}. See
// TestOperationParamsAreNotReserved.
func operationRequest(op operation, arg json.RawMessage) (path string, body []byte, reason api.AuditLogReason, err error) {
	params := map[string]json.RawMessage{}
	if arg = bytes.TrimSpace(arg); len(arg) > 0 && string(arg) != "null" {
		if err := json.Unmarshal(arg, &params); err != nil {
			return "", nil, "", fmt.Errorf("the argument must be an object of its parameters")
		}
	}

	take := func(list []apiParam) map[string]thing.Thing {
		res := make(map[string]thing.Thing, len(list))
		for _, p := range list {
			if v, ok := params[p.Name]; ok {
				res[p.Name] = jsonParamThing(v)
				delete(params, p.Name)
			}
		}
		return res
	}
	pathParams := take(op.PathParams)
	query := take(op.QueryParams)
	if path, err = operationPath(op, pathParams, query, nil); err != nil {
		return "", nil, "", err
	}

	if v, ok := params["body"]; ok && op.HasBody {
		body = v
		delete(params, "body")
	}
	if v, ok := params["reason"]; ok && op.Policy.Reason {
		var s string
		if err := json.Unmarshal(v, &s); err != nil {
			return "", nil, "", fmt.Errorf("the reason must be text")
		}
		reason = util.AuditLogReason(s)
		delete(params, "reason")
	}
	if len(params) > 0 {
		names := slices.Sorted(maps.Keys(params))
		return "", nil, "", fmt.Errorf("it has no parameter %s", names[0])
	}
	return path, body, reason, nil
}
