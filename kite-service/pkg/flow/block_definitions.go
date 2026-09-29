package flow

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// blockDefinitionsJSON describes every block: how it runs, which integrations
// it needs, and the fields and result of blocks that send a request. It's
// generated from kite-web/src/lib/blocks, run `pnpm test -u` in kite-web to
// update it. See design/integrations.md.
//
//go:embed block_definitions.json
var blockDefinitionsJSON []byte

type blockField struct {
	// Name of the setting in the node's data, like "channel_target" or
	// "max_age". Target is its name in the request.
	Name     string `json:"name"`
	In       string `json:"in"`
	Target   string `json:"target"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	// "guild" or "channel": the one the flow runs in, if the field is empty.
	Fallback  string `json:"fallback"`
	Min       *int64 `json:"min"`
	Max       *int64 `json:"max"`
	MaxLength *int   `json:"max_length"`
}

// blockRequest is how a block runs: a request to the API of an integration
// for kind "request", or the handler in nodeHandlers for kind "custom".
type blockRequest struct {
	Kind        string `json:"kind"`
	Integration string `json:"integration"`
	Operation   string `json:"operation"`
	Method      string `json:"method"`
	Path        string `json:"path"`
}

type blockDefinition struct {
	Type FlowNodeType `json:"type"`
	// Nil for blocks whose cost depends on their settings, see CreditsCost.
	Credits        *int `json:"credits"`
	AuditLogReason bool `json:"audit_log_reason"`
	// Integrations the block needs, including the one of its request.
	Requires []string     `json:"requires"`
	Run      blockRequest `json:"run"`
	Fields   []blockField `json:"fields"`
	Result   *struct {
		Thing string `json:"thing"`
		List  bool   `json:"list"`
	} `json:"result"`
}

var blockDefinitions = func() map[FlowNodeType]blockDefinition {
	var data struct {
		Blocks []blockDefinition `json:"blocks"`
	}
	if err := json.Unmarshal(blockDefinitionsJSON, &data); err != nil {
		panic(fmt.Sprintf("failed to parse block_definitions.json: %v", err))
	}

	res := make(map[FlowNodeType]blockDefinition, len(data.Blocks))
	for _, block := range data.Blocks {
		res[block.Type] = block
	}
	return res
}()

var listSeparatorRe = regexp.MustCompile(`[,\s]+`)

func (n *CompiledFlowNode) executeBlockDefinition(ctx *FlowContext, block blockDefinition) error {
	// Other integrations need app credentials, which don't exist yet.
	if block.Run.Kind != "request" || block.Run.Integration != "discord" {
		return traceError(n, fmt.Errorf("unsupported block run: %s %s", block.Run.Kind, block.Run.Integration))
	}

	pathParams := make(map[string]string)
	query := url.Values{}
	body := make(map[string]any)
	hasBody := false

	for _, field := range block.Fields {
		hasBody = hasBody || field.In == "body"

		raw := n.Data.Setting(field.Name)
		value, err := ctx.evalFieldValue(raw)
		if err != nil {
			return traceError(n, err)
		}
		// Only a setting that's left empty falls back, not a placeholder
		// that turns out empty, which could point e.g. a ban at the wrong
		// server.
		if raw == nil || raw == "" {
			value = field.fallbackValue(ctx)
		}
		if isEmptyFieldValue(value) {
			if field.Required || field.In == "path" {
				return traceError(n, fmt.Errorf("%s is required", field.Name))
			}
			continue
		}

		v, err := field.value(value)
		if err != nil {
			return traceError(n, fmt.Errorf("invalid value for %s: %w", field.Name, err))
		}

		switch field.In {
		case "path":
			pathParams[field.Target], err = apiPathSegment(field.Name, fmt.Sprint(v))
			if err != nil {
				return traceError(n, err)
			}
		case "query":
			if list, ok := v.([]string); ok {
				for _, item := range list {
					query.Add(field.Target, item)
				}
			} else {
				query.Set(field.Target, fmt.Sprint(v))
			}
		case "body":
			body[field.Target] = v
		}
	}

	path := pathParamRe.ReplaceAllStringFunc(block.Run.Path, func(m string) string {
		return pathParams[m[1:len(m)-1]]
	})
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var reqBody []byte
	if hasBody && block.Run.Method != http.MethodGet {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return traceError(n, err)
		}
	}

	var reason api.AuditLogReason
	if block.AuditLogReason {
		auditLogReason, err := ctx.EvalTemplate(n.Data.AuditLogReason)
		if err != nil {
			return traceError(n, err)
		}
		reason = api.AuditLogReason(auditLogReason.String())
	}

	resBody, err := ctx.Discord.APIRequest(ctx, provider.DiscordAPIRequest{
		Method: block.Run.Method,
		Path:   path,
		Body:   reqBody,
		Reason: reason,
	})
	if err != nil {
		return traceError(n, err)
	}

	if block.Result != nil {
		result, err := blockResult(block, resBody)
		if err != nil {
			return traceError(n, err)
		}
		ctx.StoreNodeResult(n, result)
	}

	return n.ExecuteChildren(ctx)
}

func (ctx *FlowContext) evalFieldValue(raw any) (thing.Thing, error) {
	switch v := raw.(type) {
	case string:
		return ctx.EvalTemplate(v)
	case []any:
		// The items of lists can be templates too.
		items := make([]thing.Thing, len(v))
		for i, item := range v {
			value, err := ctx.evalFieldValue(item)
			if err != nil {
				return thing.Null, err
			}
			items[i] = value
		}
		return thing.NewArray(items), nil
	case nil, bool, float64, json.Number, map[string]any:
		return thing.NewFromJSONValue(v), nil
	}

	// Settings of FlowNodeData that are structs, like emoji_data, are read
	// like their JSON.
	b, err := json.Marshal(raw)
	if err != nil {
		return thing.Null, err
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return thing.Null, err
	}
	return thing.NewFromJSONValue(v), nil
}

func isEmptyFieldValue(value thing.Thing) bool {
	return !value.IsDiscordEntity() && strings.TrimSpace(value.String()) == ""
}

func (f blockField) fallbackValue(ctx *FlowContext) thing.Thing {
	var id discord.Snowflake
	switch f.Fallback {
	case "guild":
		id = discord.Snowflake(ctx.Data.GuildID())
	case "channel":
		id = discord.Snowflake(ctx.Data.ChannelID())
	}
	if !id.IsValid() {
		return thing.Null
	}
	return thing.NewString(id.String())
}

// value converts an evaluated field to the JSON value the request needs.
func (f blockField) value(value thing.Thing) (any, error) {
	switch f.Type {
	case "snowflake":
		return discordAPISnowflake(value)
	case "snowflake_list":
		items, ok := apiListItems(value)
		if !ok {
			for _, s := range listSeparatorRe.Split(strings.TrimSpace(value.String()), -1) {
				items = append(items, thing.NewString(s))
			}
		}

		ids := make([]string, len(items))
		for i, item := range items {
			id, err := discordAPISnowflake(item)
			if err != nil {
				return nil, err
			}
			ids[i] = id
		}
		if err := f.checkRange(int64(len(ids)), "IDs"); err != nil {
			return nil, err
		}
		return ids, nil
	case "integer":
		n, err := strconv.ParseInt(strings.TrimSpace(value.String()), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("must be a whole number")
		}
		if err := f.checkRange(n, ""); err != nil {
			return nil, err
		}
		return n, nil
	case "emoji":
		// A custom emoji is sent as "name:id", a standard one as itself.
		emoji := value.Object()
		name := strings.TrimSpace(emoji["name"].String())
		if name == "" {
			return nil, fmt.Errorf("must be an emoji")
		}
		if id := strings.TrimSpace(emoji["id"].String()); id != "" {
			return name + ":" + id, nil
		}
		return name, nil
	case "seconds":
		seconds, err := parseSeconds(value)
		if err != nil {
			return nil, err
		}
		if err := f.checkRange(seconds, ""); err != nil {
			return nil, err
		}
		return seconds, nil
	case "seconds_until":
		seconds, err := parseSeconds(value)
		if err != nil {
			return nil, err
		}
		if err := f.checkRange(seconds, ""); err != nil {
			return nil, err
		}
		return discord.Timestamp(time.Now().UTC().Add(time.Duration(seconds) * time.Second)), nil
	case "boolean":
		switch strings.TrimSpace(value.String()) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
		return nil, fmt.Errorf("must be true or false")
	default:
		s := value.String()
		if f.MaxLength != nil && utf8.RuneCountInString(s) > *f.MaxLength {
			return nil, fmt.Errorf("must be at most %d characters", *f.MaxLength)
		}
		return s, nil
	}
}

// parseSeconds drops fractions, like the blocks did before they were
// requests.
func parseSeconds(value thing.Thing) (int64, error) {
	seconds, err := strconv.ParseFloat(strings.TrimSpace(value.String()), 64)
	// Also rejects NaN and values that don't fit into a duration.
	if err != nil || !(math.Abs(seconds) < 1e12) {
		return 0, fmt.Errorf("must be a number of seconds")
	}
	return int64(seconds), nil
}

func (f blockField) checkRange(n int64, unit string) error {
	if unit != "" {
		unit = " " + unit
	}
	if f.Min != nil && n < *f.Min {
		return fmt.Errorf("must be at least %d%s", *f.Min, unit)
	}
	if f.Max != nil && n > *f.Max {
		return fmt.Errorf("must be at most %d%s", *f.Max, unit)
	}
	return nil
}

func blockResult(block blockDefinition, body []byte) (thing.Thing, error) {
	switch block.Result.Thing {
	case "discord_message":
		return decodeDiscordResult(body, block.Result.List, thing.NewDiscordMessage)
	case "discord_role":
		return decodeDiscordResult(body, block.Result.List, thing.NewDiscordRole)
	case "discord_channel":
		return decodeDiscordResult(body, block.Result.List, thing.NewDiscordChannel)
	default:
		return discordAPIResult(body)
	}
}

func decodeDiscordResult[T any](body []byte, list bool, wrap func(T) thing.Thing) (thing.Thing, error) {
	if !list {
		var v T
		if err := json.Unmarshal(body, &v); err != nil {
			return thing.Null, fmt.Errorf("failed to parse response: %w", err)
		}
		return wrap(v), nil
	}

	var v []T
	if err := json.Unmarshal(body, &v); err != nil {
		return thing.Null, fmt.Errorf("failed to parse response: %w", err)
	}
	res := make([]thing.Thing, len(v))
	for i, item := range v {
		res[i] = wrap(item)
	}
	return thing.NewArray(res), nil
}
