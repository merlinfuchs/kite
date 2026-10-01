package flow

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"gopkg.in/guregu/null.v4"
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
	// The values a string field can have, if they're limited.
	Options []string `json:"options"`
}

// blockRequest is how a block runs: a request to the API of an integration
// for kind "request", or the handler in nodeHandlers for kind "custom".
type blockRequest struct {
	Kind        string `json:"kind"`
	Integration string `json:"integration"`
	Operation   string `json:"operation"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	// Values Kite adds to the request, which aren't settings.
	Inject []blockRequestInject `json:"inject"`
}

// blockRequestInject is a value Kite adds to a request. Value
// "discord_bot_token" is the app's bot token, for services that call Discord
// for the app.
type blockRequestInject struct {
	In    string `json:"in"`
	Name  string `json:"name"`
	Value string `json:"value"`
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

// Integration is a service blocks can talk to.
type Integration struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	BaseURL     string          `json:"base_url"`
	Auth        IntegrationAuth `json:"auth"`
	// One of the Availability constants.
	Availability string `json:"availability"`
	// A GET endpoint, relative to BaseURL, that checks a credential.
	TestPath string `json:"test_path"`
}

// IntegrationAuth is how requests to an integration prove who they are:
// "discord_bot", "none", or an app credential in a "header" or "query"
// parameter.
type IntegrationAuth struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
	Label  string `json:"label"`
}

const (
	// Every app can use the integration.
	AvailabilityAlways = "always"
	// Apps can use it until they turn it off.
	AvailabilityDefault = "default"
	// Apps turn it on, by entering the credential if it needs one.
	AvailabilityOptIn = "opt_in"
)

// NeedsCredential reports whether the app has to enter a credential to use
// the integration.
func (i Integration) NeedsCredential() bool {
	return i.Auth.Type == "header" || i.Auth.Type == "query"
}

// Enabled reports whether an app can use the integration, given whether it
// enabled or disabled it. Without a choice, the integration's default applies.
// Integrations that need a credential are opt-in, and have a choice exactly
// when the app connected them, as the credential references it.
func (i Integration) Enabled(choice null.Bool) bool {
	switch {
	case i.Availability == AvailabilityAlways:
		return true
	case choice.Valid:
		return choice.Bool
	default:
		return i.Availability == AvailabilityDefault
	}
}

// NewRequest creates a request to the integration's API, with the app's
// credential if the integration needs one.
func (i Integration) NewRequest(ctx context.Context, method string, path string, credential string, body []byte) (*http.Request, error) {
	u, err := url.Parse(strings.TrimSuffix(i.BaseURL, "/") + path)
	if err != nil {
		return nil, err
	}
	if i.Auth.Type == "query" {
		query := u.Query()
		query.Set(i.Auth.Name, credential)
		u.RawQuery = query.Encode()
	}

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if i.Auth.Type == "header" {
		req.Header.Set(i.Auth.Name, i.Auth.Prefix+credential)
	}
	return req, nil
}

var blockDefinitions, integrations = func() (map[FlowNodeType]blockDefinition, map[string]Integration) {
	var data struct {
		Integrations []Integration     `json:"integrations"`
		Blocks       []blockDefinition `json:"blocks"`
	}
	if err := json.Unmarshal(blockDefinitionsJSON, &data); err != nil {
		panic(fmt.Sprintf("failed to parse block_definitions.json: %v", err))
	}

	blocks := make(map[FlowNodeType]blockDefinition, len(data.Blocks))
	for _, block := range data.Blocks {
		blocks[block.Type] = block
	}
	integrations := make(map[string]Integration, len(data.Integrations))
	for _, integration := range data.Integrations {
		integrations[integration.ID] = integration
	}
	return blocks, integrations
}()

// AddIntegration adds an integration until the returned function is called.
// It's for tests of other packages, which need an integration with a
// credential while Kite has none.
func AddIntegration(integration Integration) (remove func()) {
	integrations[integration.ID] = integration
	return func() { delete(integrations, integration.ID) }
}

// GetIntegration returns the integration with the given ID.
func GetIntegration(id string) (Integration, bool) {
	integration, ok := integrations[id]
	return integration, ok
}

// Integrations returns all integrations, sorted by ID.
func Integrations() []Integration {
	res := make([]Integration, 0, len(integrations))
	for _, integration := range integrations {
		res = append(res, integration)
	}
	slices.SortFunc(res, func(a, b Integration) int { return strings.Compare(a.ID, b.ID) })
	return res
}

// BlockIntegrations returns the IDs of the integrations a block needs.
func BlockIntegrations(nodeType FlowNodeType) []string {
	return blockDefinitions[nodeType].Requires
}

var listSeparatorRe = regexp.MustCompile(`[,\s]+`)

func (n *CompiledFlowNode) executeBlockDefinition(ctx *FlowContext, block blockDefinition) error {
	integration, ok := integrations[block.Run.Integration]
	if block.Run.Kind != "request" || !ok {
		return traceError(n, fmt.Errorf("unsupported block run: %s %s", block.Run.Kind, block.Run.Integration))
	}

	pathParams := make(map[string]string)
	query := url.Values{}
	body := make(map[string]any)
	hasBody := false

	for _, field := range block.Fields {
		hasBody = hasBody || field.In == "body"

		raw := n.Data.Setting(field.Name)
		if field.Type == "json_object" {
			props, err := ctx.evalJSONObjectField(raw)
			if err != nil {
				return traceError(n, fmt.Errorf("invalid value for %s: %w", field.Name, err))
			}
			if props == nil && field.Required {
				return traceError(n, fmt.Errorf("%s is required", field.Name))
			}
			maps.Copy(body, props)
			continue
		}

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

	var secrets []string
	for _, inject := range block.Run.Inject {
		if inject.In != "body" || inject.Value != "discord_bot_token" {
			return traceError(n, fmt.Errorf("unsupported injected value: %s in %s", inject.Value, inject.In))
		}
		token := ctx.Discord.BotToken()
		secrets = append(secrets, token)
		body[inject.Name] = token
		hasBody = true
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

	var resBody []byte
	var err error
	if integration.Auth.Type == "discord_bot" {
		resBody, err = ctx.Discord.APIRequest(ctx, provider.DiscordAPIRequest{
			Method: block.Run.Method,
			Path:   path,
			Body:   reqBody,
			Reason: reason,
		})
	} else {
		resBody, err = integrationRequest(ctx, integration, block.Run.Method, path, reqBody, secrets...)
	}
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

// evalJSONObjectField reads a JSON object stored as text, with the
// placeholders in its strings filled in. It's nil if the setting is empty.
func (ctx *FlowContext) evalJSONObjectField(raw any) (map[string]any, error) {
	text, ok := raw.(string)
	if !ok && raw != nil {
		return nil, fmt.Errorf("must be a JSON object")
	}
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}

	b, err := ctx.EvalJSONTemplate(json.RawMessage(text))
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	// Keeps IDs written as numbers exact.
	dec.UseNumber()
	var props map[string]any
	if err := dec.Decode(&props); err != nil || props == nil {
		return nil, fmt.Errorf("must be a JSON object")
	}
	return props, nil
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
		if len(f.Options) > 0 && !slices.Contains(f.Options, s) {
			return nil, fmt.Errorf("must be one of %s", strings.Join(f.Options, ", "))
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

// integrationRequest sends a request to an integration other than Discord,
// with the app's credential if it needs one. The credential only goes to the
// integration's own host: its base URL comes from its definition, and
// redirects aren't followed. Errors don't contain the credential or the
// other secrets of the request, like a bot token in its body.
func integrationRequest(ctx *FlowContext, integration Integration, method string, path string, body []byte, secrets ...string) ([]byte, error) {
	var credential string
	if integration.NeedsCredential() {
		var err error
		credential, err = integrationCredential(ctx, integration)
		if err != nil {
			return nil, err
		}
	}
	redact := &requestSecrets{values: map[string]string{"credential": credential}}
	for i, secret := range secrets {
		redact.values[fmt.Sprint(i)] = secret
	}

	req, err := integration.NewRequest(ctx, method, path, credential, body)
	if err != nil {
		return nil, redact.Redact(err)
	}

	resp, err := ctx.HTTP.HTTPRequestWithoutRedirects(ctx, req)
	if err != nil {
		return nil, redact.Redact(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, thing.MaxBodySize+1))
	if err != nil {
		return nil, redact.Redact(err)
	}
	if len(data) > thing.MaxBodySize {
		return nil, fmt.Errorf("body size exceeds max body size of %d bytes", thing.MaxBodySize)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// Redacted before it's cut, which could leave part of the credential.
		msg := []rune(redact.redactString(string(data)))
		if len(msg) > 300 {
			msg = msg[:300]
		}
		return nil, fmt.Errorf("%s returned %s: %s", integration.Name, resp.Status, string(msg))
	}
	return data, nil
}

func integrationCredential(ctx *FlowContext, integration Integration) (string, error) {
	if ctx.Integration == nil {
		return "", notEnabledError(integration)
	}

	credential, err := ctx.Integration.Credential(ctx, integration.ID)
	if err != nil {
		if errors.Is(err, provider.ErrNotFound) {
			return "", notEnabledError(integration)
		}
		return "", err
	}
	return credential, nil
}

// checkIntegrations fails if the app didn't enable an integration the block
// needs.
func (n *CompiledFlowNode) checkIntegrations(ctx *FlowContext) error {
	for _, id := range blockDefinitions[n.Type].Requires {
		integration, ok := integrations[id]
		if !ok || integration.Availability == AvailabilityAlways {
			continue
		}

		var choice null.Bool
		if ctx.Integration != nil {
			var err error
			choice, err = ctx.Integration.Choice(ctx, id)
			if err != nil {
				return err
			}
		}
		if !integration.Enabled(choice) {
			return notEnabledError(integration)
		}
	}
	return nil
}

func notEnabledError(integration Integration) error {
	return fmt.Errorf("%s isn't enabled, enable it in the app's integrations", integration.Name)
}
