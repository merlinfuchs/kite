package flow

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// discordAPIJSON lists the Discord API endpoints the Discord API Request block
// can call. It's generated from Discord's OpenAPI spec, run
// `node scripts/discord-api.mjs` and `pnpm test -u` in kite-web to update it.
//
//go:embed discord_api.json
var discordAPIJSON []byte

type discordAPIParamType string

const (
	discordAPIParamTypeSnowflake discordAPIParamType = "snowflake"
	discordAPIParamTypeInteger   discordAPIParamType = "integer"
	discordAPIParamTypeNumber    discordAPIParamType = "number"
	discordAPIParamTypeBoolean   discordAPIParamType = "boolean"
	discordAPIParamTypeArray     discordAPIParamType = "array"
	discordAPIParamTypeString    discordAPIParamType = "string"
)

type discordAPIParam struct {
	Name     string              `json:"name"`
	Type     discordAPIParamType `json:"type"`
	Required bool                `json:"required"`
}

type discordAPIOperation struct {
	ID          string            `json:"id"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	PathParams  []discordAPIParam `json:"path_params"`
	QueryParams []discordAPIParam `json:"query_params"`
	// "required", "optional" or empty if the endpoint takes no body.
	Body string `json:"body"`
}

var discordAPIOperations = func() map[string]discordAPIOperation {
	var data struct {
		Operations []discordAPIOperation `json:"operations"`
	}
	if err := json.Unmarshal(discordAPIJSON, &data); err != nil {
		panic(fmt.Sprintf("failed to parse discord_api.json: %v", err))
	}

	res := make(map[string]discordAPIOperation, len(data.Operations))
	for _, op := range data.Operations {
		res[op.ID] = op
	}
	return res
}()

var (
	snowflakeRe = regexp.MustCompile(`^[0-9]+$`)
	pathParamRe = regexp.MustCompile(`\{([a-z0-9_]+)\}`)
)

// discordAPIPath builds the path of a request to op, relative to the API base
// URL, from the evaluated path and query parameters.
func discordAPIPath(op discordAPIOperation, pathParams map[string]thing.Thing, query map[string]thing.Thing) (string, error) {
	params := make(map[string]string, len(op.PathParams))
	for _, p := range op.PathParams {
		value, ok := pathParams[p.Name]
		if !ok {
			return "", fmt.Errorf("missing path parameter %s", p.Name)
		}

		v, err := discordAPIParamValue(p, value)
		if err != nil {
			return "", err
		}
		// Values like "../.." could point the request at another endpoint.
		// Escaping alone doesn't cover "." and "..", and proxies may decode an
		// escaped "/", so none of these are allowed at all.
		if v == "." || v == ".." || strings.ContainsAny(v, `/\`) {
			return "", fmt.Errorf("invalid value for path parameter %s", p.Name)
		}
		params[p.Name] = url.PathEscape(v)
	}
	for name := range pathParams {
		if _, ok := params[name]; !ok {
			return "", fmt.Errorf("unknown path parameter %s", name)
		}
	}

	path := pathParamRe.ReplaceAllStringFunc(op.Path, func(m string) string {
		return params[m[1:len(m)-1]]
	})

	values := url.Values{}
	for _, p := range op.QueryParams {
		value, ok := query[p.Name]
		if !ok {
			if p.Required {
				return "", fmt.Errorf("missing query parameter %s", p.Name)
			}
			continue
		}

		v, err := discordAPIParamValue(p, value)
		if err != nil {
			return "", err
		}
		values.Set(p.Name, v)
	}
	for name := range query {
		if !values.Has(name) {
			return "", fmt.Errorf("unknown query parameter %s", name)
		}
	}

	if len(values) > 0 {
		path += "?" + values.Encode()
	}
	return path, nil
}

func discordAPIParamValue(p discordAPIParam, value thing.Thing) (string, error) {
	var (
		v   string
		err error
	)
	switch p.Type {
	case discordAPIParamTypeSnowflake:
		v, err = discordAPISnowflake(value)
	case discordAPIParamTypeInteger:
		v = strings.TrimSpace(value.String())
		_, err = strconv.ParseInt(v, 10, 64)
	case discordAPIParamTypeNumber:
		v = strings.TrimSpace(value.String())
		_, err = strconv.ParseFloat(v, 64)
	case discordAPIParamTypeBoolean:
		v = strings.TrimSpace(value.String())
		if v != "true" && v != "false" {
			err = fmt.Errorf("must be true or false")
		}
	default:
		v = value.String()
		if v == "" {
			err = fmt.Errorf("must not be empty")
		}
	}
	if err != nil {
		return "", fmt.Errorf("invalid value for parameter %s: %w", p.Name, err)
	}
	return v, nil
}

// discordAPISnowflake returns the ID of a Discord object, or a string that
// already is an ID.
func discordAPISnowflake(value thing.Thing) (string, error) {
	switch value.Type {
	case thing.TypeString:
		v := strings.TrimSpace(value.String())
		if !snowflakeRe.MatchString(v) {
			return "", fmt.Errorf("must be an ID")
		}
		return v, nil
	case thing.TypeInt:
		if value.Int() <= 0 {
			return "", fmt.Errorf("must be an ID")
		}
		return value.String(), nil
	case thing.TypeDiscordMessage, thing.TypeDiscordUser, thing.TypeDiscordMember,
		thing.TypeDiscordChannel, thing.TypeDiscordGuild, thing.TypeDiscordRole:
		return value.Snowflake().String(), nil
	default:
		return "", fmt.Errorf("must be an ID")
	}
}

// evalDiscordAPIBody evaluates the placeholders in the string values of a JSON
// body. A value that is a single placeholder keeps the type of its result, so
// numbers and lists can be filled in too.
func evalDiscordAPIBody(ctx *FlowContext, body json.RawMessage) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	// Keeps IDs written as numbers exact.
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to parse body: %w", err)
	}

	v, err := evalDiscordAPIBodyValue(ctx, v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func evalDiscordAPIBodyValue(ctx *FlowContext, v any) (any, error) {
	switch v := v.(type) {
	case map[string]any:
		for key, value := range v {
			res, err := evalDiscordAPIBodyValue(ctx, value)
			if err != nil {
				return nil, err
			}
			v[key] = res
		}
		return v, nil
	case []any:
		for i, value := range v {
			res, err := evalDiscordAPIBodyValue(ctx, value)
			if err != nil {
				return nil, err
			}
			v[i] = res
		}
		return v, nil
	case string:
		// An empty template evaluates to null, which isn't what a plain string
		// should become.
		if !strings.Contains(v, "{{") {
			return v, nil
		}

		res, err := ctx.EvalTemplateKeepSpace(v)
		if err != nil {
			return nil, err
		}
		return discordAPIBodyJSON(res), nil
	default:
		return v, nil
	}
}

func discordAPIBodyJSON(t thing.Thing) any {
	switch t.Type {
	case thing.TypeString, thing.TypeInt, thing.TypeFloat, thing.TypeBool:
		return t.Value
	case thing.TypeArray:
		arr := t.Array()
		res := make([]any, len(arr))
		for i, item := range arr {
			res[i] = discordAPIBodyJSON(item)
		}
		return res
	case thing.TypeObject:
		obj := t.Object()
		res := make(map[string]any, len(obj))
		for key, item := range obj {
			res[key] = discordAPIBodyJSON(item)
		}
		return res
	case thing.TypeDiscordMessage, thing.TypeDiscordUser, thing.TypeDiscordMember,
		thing.TypeDiscordChannel, thing.TypeDiscordGuild, thing.TypeDiscordRole:
		return t.Snowflake().String()
	}

	// Lists and maps built in expressions, like {{[1, 2]}}, aren't wrapped.
	switch v := t.Value.(type) {
	case nil:
		return nil
	case []any:
		res := make([]any, len(v))
		for i, item := range v {
			res[i] = discordAPIBodyJSON(thing.NewGuessTypeWithFallback(item))
		}
		return res
	case map[string]any:
		res := make(map[string]any, len(v))
		for key, item := range v {
			res[key] = discordAPIBodyJSON(thing.NewGuessTypeWithFallback(item))
		}
		return res
	default:
		return t.String()
	}
}

// discordAPIResult turns the JSON of a response into a result that later
// blocks can read fields of, like {{result('id').code}}.
func discordAPIResult(body []byte) (thing.Thing, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return thing.Null, nil
	}

	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return thing.Null, fmt.Errorf("failed to parse response: %w", err)
	}
	return discordAPIResultValue(v), nil
}

func discordAPIResultValue(v any) thing.Thing {
	switch v := v.(type) {
	case map[string]any:
		res := make(map[string]thing.Thing, len(v))
		for key, item := range v {
			res[key] = discordAPIResultValue(item)
		}
		return thing.NewObject(res)
	case []any:
		res := make([]thing.Thing, len(v))
		for i, item := range v {
			res[i] = discordAPIResultValue(item)
		}
		return thing.NewArray(res)
	default:
		return thing.NewGuessTypeWithFallback(v)
	}
}
