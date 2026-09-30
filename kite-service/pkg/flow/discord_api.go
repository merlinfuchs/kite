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
	HasBody     bool              `json:"has_body"`
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

var pathParamRe = regexp.MustCompile(`\{([a-z0-9_]+)\}`)

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
		params[p.Name], err = apiPathSegment(p.Name, v)
		if err != nil {
			return "", err
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

		// A list is sent as the parameter repeated for each item.
		items := []thing.Thing{value}
		if p.Type == discordAPIParamTypeArray {
			if list, ok := apiListItems(value); ok {
				items = list
			}
		}
		for _, item := range items {
			v, err := discordAPIParamValue(p, item)
			if err != nil {
				return "", err
			}
			values.Add(p.Name, v)
		}
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

// apiPathSegment escapes a value for a path parameter.
func apiPathSegment(name string, v string) (string, error) {
	// Values like "../.." could point the request at another endpoint.
	// Escaping alone doesn't cover "." and "..", and proxies may decode an
	// escaped "/", so none of these are allowed at all.
	if v == "." || v == ".." || strings.ContainsAny(v, `/\`) {
		return "", fmt.Errorf("invalid value for path parameter %s", name)
	}
	return url.PathEscape(v), nil
}

// apiListItems returns the items of a list, like the result of a placeholder
// such as {{[1, 2]}}.
func apiListItems(value thing.Thing) ([]thing.Thing, bool) {
	list, ok := value.JSONValue().([]any)
	if !ok {
		return nil, false
	}

	items := make([]thing.Thing, len(list))
	for i, item := range list {
		items[i] = thing.NewGuessTypeWithFallback(item)
	}
	return items, true
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
	if value.IsDiscordEntity() || value.Type == thing.TypeObject {
		id := value.Snowflake()
		if !id.IsValid() {
			return "", fmt.Errorf("must be an ID")
		}
		return id.String(), nil
	}

	// Also covers placeholders like {{channel}}, whose text is the ID.
	v := strings.TrimSpace(value.String())
	if id, err := strconv.ParseUint(v, 10, 64); err != nil || id == 0 {
		return "", fmt.Errorf("must be an ID")
	}
	return v, nil
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
	return thing.NewFromJSONValue(v), nil
}
