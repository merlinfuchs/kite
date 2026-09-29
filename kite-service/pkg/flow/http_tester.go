package flow

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// Stages at which a tested request can fail, so the editor can point at the
// part of the block that needs fixing.
const (
	HTTPRequestTestStageBuild     = "build"
	HTTPRequestTestStageSend      = "send"
	HTTPRequestTestStageStatus    = "status"
	HTTPRequestTestStageTransform = "transform"
)

// maxHTTPRequestTestPreview bounds how much of a request or response body is
// sent back to the editor. The flow itself still sees the whole body.
const maxHTTPRequestTestPreview = 256 * 1024

// HTTPRequestTestOpts configures a test run of an HTTP request block outside
// of any flow, e.g. from the editor before the flow is deployed.
type HTTPRequestTestOpts struct {
	Data *HTTPRequestData
	// Values stand in for the placeholders the block uses, which have no
	// interaction or event to come from during a test. See applyTestValues
	// for the accepted keys.
	Values  map[string]string
	HTTP    provider.HTTPProvider
	Timeout time.Duration
}

// HTTPRequestTestResult is everything the editor shows about a test run. It
// is always returned, even when the request fails part way, so whatever was
// built or received before the failure can still be shown.
type HTTPRequestTestResult struct {
	Method         string
	URL            string
	RequestHeaders map[string]string
	RequestBody    string
	RequestBodyCut bool

	Response        *thing.HTTPResponseValue
	ResponseBody    string
	ResponseBodyCut bool
	ResponseBinary  bool

	// Result is the JSON of what the transform expression returned. It's
	// only set when the block has a transform, otherwise the response itself
	// is the result.
	Result json.RawMessage

	Error      string
	ErrorStage string
	Duration   time.Duration
}

// RunHTTPRequestTest builds and sends the request of an HTTP request block the
// same way a flow would, and applies its response options.
func RunHTTPRequestTest(ctx context.Context, opts HTTPRequestTestOpts) HTTPRequestTestResult {
	var res HTTPRequestTestResult

	if opts.Data == nil {
		res.Error = "the block has no request configured"
		res.ErrorStage = HTTPRequestTestStageBuild
		return res
	}

	state := NewFlowContextState()
	env := eval.Env{}
	applyTestValues(env, state, opts.Values)

	fctx := NewContext(
		ctx,
		opts.Timeout,
		nil,
		FlowProviders{HTTP: opts.HTTP},
		FlowContextLimits{},
		eval.NewContext(env),
		state,
	)
	defer fctx.Cancel()

	req, body, err := fctx.buildHTTPRequest(opts.Data)
	if err != nil {
		res.Error = err.Error()
		res.ErrorStage = HTTPRequestTestStageBuild
		return res
	}

	res.Method = req.Method
	res.URL = req.URL.String()
	res.RequestHeaders = make(map[string]string, len(req.Header))
	for k, v := range req.Header {
		res.RequestHeaders[k] = strings.Join(v, ",")
	}
	res.RequestBody, res.RequestBodyCut, _ = previewBody(body)

	start := time.Now()
	resp, err := fctx.sendHTTPRequest(req)
	res.Duration = time.Since(start)
	if err != nil {
		res.Error = describeSendError(err, opts.Timeout)
		res.ErrorStage = HTTPRequestTestStageSend
		return res
	}

	res.Response = &resp
	res.ResponseBody, res.ResponseBodyCut, res.ResponseBinary = previewBody(resp.Body)

	if opts.Data.FailOnErrorStatus {
		if err := checkHTTPResponseStatus(resp); err != nil {
			res.Error = err.Error()
			res.ErrorStage = HTTPRequestTestStageStatus
			return res
		}
	}

	if opts.Data.ResponseTransform != "" {
		result, err := fctx.transformHTTPResponse(opts.Data.ResponseTransform, resp)
		if err != nil {
			res.Error = err.Error()
			res.ErrorStage = HTTPRequestTestStageTransform
			return res
		}

		b, err := eval.ThingToJSON(result)
		if err != nil {
			b, _ = json.Marshal(result.String())
		}
		res.Result = b
	}

	return res
}

func describeSendError(err error, timeout time.Duration) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "the request timed out after " + timeout.String()
	}
	return err.Error()
}

// previewBody turns a body into text for the editor. Binary bodies are
// replaced with a note, as there's nothing useful to show for them.
func previewBody(body []byte) (text string, cut bool, binary bool) {
	if len(body) == 0 {
		return "", false, false
	}

	if len(body) > maxHTTPRequestTestPreview {
		body = body[:maxHTTPRequestTestPreview]
		cut = true
		// Don't cut a multi-byte character in half.
		for i := 0; i < utf8.UTFMax && len(body) > 0 && !utf8.Valid(body); i++ {
			body = body[:len(body)-1]
		}
	}

	if !utf8.Valid(body) {
		return "", cut, true
	}

	return string(body), cut, false
}

var (
	// arg('name'), input("id"), var('name'), result('node')
	testValueCallRegex = regexp.MustCompile(`^(arg|input|var|result)\(\s*['"]([^'"]+)['"]\s*\)$`)
	// user.id, interaction.guild.name, nodes.abc.result
	testValuePathRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z0-9_]+)*$`)
	// 42, -1.5
	testValueNumberRegex = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?$`)
)

// applyTestValues makes the test values available to the block's
// placeholders. Keys are written the way the placeholder is, without braces:
//
//   - arg('name') and input('id') answer the arg and input functions.
//   - var('name') sets a temporary variable.
//   - result('node') and nodes.node.result set the result of another block.
//   - Any other dotted path, like user.id, sets that path.
//
// Keys that match none of these are ignored.
func applyTestValues(env eval.Env, state *FlowContextState, values map[string]string) {
	funcs := map[string]map[string]any{}

	for rawKey, rawValue := range values {
		key := strings.TrimSpace(rawKey)
		value := parseTestValue(rawValue)

		if m := testValueCallRegex.FindStringSubmatch(key); m != nil {
			name := m[2]
			switch m[1] {
			case "var":
				state.Temporaries[name] = thing.NewGuessTypeWithFallback(value)
			case "result":
				state.GetNodeState(name).Result = thing.NewGuessTypeWithFallback(value)
			default:
				if funcs[m[1]] == nil {
					funcs[m[1]] = map[string]any{}
				}
				funcs[m[1]][name] = value
			}
			continue
		}

		if !testValuePathRegex.MatchString(key) {
			continue
		}

		parts := strings.Split(key, ".")
		if parts[0] == "nodes" && len(parts) >= 2 && len(parts) <= 3 &&
			(len(parts) == 2 || parts[2] == "result") {
			state.GetNodeState(parts[1]).Result = thing.NewGuessTypeWithFallback(value)
			continue
		}

		setTestValuePath(env, parts, value)
	}

	for name, lookup := range funcs {
		env[name] = func(key string) any {
			return lookup[key]
		}
	}
}

func setTestValuePath(env eval.Env, parts []string, value any) {
	current := map[string]any(env)
	for i, part := range parts {
		if i == len(parts)-1 {
			// Don't let a shorter key wipe out the children of a longer one,
			// e.g. user = x after user.id = 1.
			if _, isMap := current[part].(map[string]any); !isMap {
				current[part] = value
			}
			return
		}

		next, ok := current[part].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[part] = next
		}
		current = next
	}
}

// parseTestValue keeps test values as text, except for JSON objects, arrays,
// booleans and small numbers. Large numbers stay text so Discord IDs don't lose
// precision.
func parseTestValue(raw string) any {
	trimmed := strings.TrimSpace(raw)

	switch {
	case trimmed == "true":
		return true
	case trimmed == "false":
		return false
	case trimmed == "null":
		return nil
	case strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "["):
		var v any
		if err := json.Unmarshal([]byte(trimmed), &v); err == nil {
			return v
		}
	case len(trimmed) <= 15 && testValueNumberRegex.MatchString(trimmed):
		if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
			return i
		}
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f
		}
	}

	return raw
}
