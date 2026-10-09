package flow

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testClientProvider struct{}

func (testClientProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req.WithContext(ctx))
}

func (p testClientProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	return p.HTTPRequest(ctx, req)
}

// echoServer answers with what it received, like an API that reflects the
// request.
func echoServer(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Echo-Auth", r.Header.Get("Authorization"))
		if r.URL.Path == "/fail" {
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte(`{"error":"nope"}`))
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"path":   r.URL.Path,
			"query":  r.URL.Query().Get("q"),
			"header": r.Header.Get("X-Test"),
			"auth":   r.Header.Get("Authorization"),
			"body":   string(body),
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func runTest(data *HTTPRequestData, values map[string]string, secrets map[string]string) HTTPRequestTestResult {
	return RunHTTPRequestTest(context.Background(), HTTPRequestTestOpts{
		Data:    data,
		Values:  values,
		HTTP:    testClientProvider{},
		Secret:  &secretTestProvider{values: secrets},
		Timeout: 5 * time.Second,
	})
}

func TestRunHTTPRequestTest(t *testing.T) {
	srv := echoServer(t)

	res := runTest(&HTTPRequestData{
		Method:            "POST",
		URL:               srv.URL + "/users/{{user.id}}",
		Query:             []HTTPRequestDataKeyValue{{Key: "q", Value: "{{arg('term')}}"}},
		Headers:           []HTTPRequestDataKeyValue{{Key: "X-Test", Value: "{{var('token')}}"}},
		BodyType:          HTTPRequestBodyTypeJSON,
		Body:              `{"n": {{result('count')}}, "name": "{{user.name}}"}`,
		ResponseTransform: "response.data().path",
	}, map[string]string{
		"user.id":         "123456789012345678",
		"user.name":       `a "quoted" name`,
		"arg('term')":     "hello world",
		"var('token')":    "tok",
		"result('count')": "5",
	}, nil)

	require.Empty(t, res.Error, "stage %s", res.ErrorStage)
	assert.Equal(t, "POST", res.Method)
	assert.True(t, strings.HasPrefix(res.URL, srv.URL+"/users/123456789012345678?"))
	assert.Equal(t, "application/json", res.RequestHeaders["Content-Type"])
	assert.Equal(t, `{"n": 5, "name": "a \"quoted\" name"}`, res.RequestBody)

	require.NotNil(t, res.Response)
	assert.Equal(t, 200, res.Response.StatusCode)
	assert.Nil(t, res.Response.Body, "the body is only returned as a preview")
	assert.Equal(t, len(res.ResponseBody), res.ResponseBodySize)
	assert.Contains(t, res.ResponseBody, `"query":"hello world"`)
	assert.Contains(t, res.ResponseBody, `"header":"tok"`)
	assert.JSONEq(t, `"/users/123456789012345678"`, string(res.Result))
}

func TestRunHTTPRequestTestFailures(t *testing.T) {
	srv := echoServer(t)

	// The response is kept when fail on error status stops the request.
	res := runTest(&HTTPRequestData{URL: srv.URL + "/fail", FailOnErrorStatus: true}, nil, nil)
	assert.Equal(t, HTTPRequestTestStageStatus, res.ErrorStage)
	require.NotNil(t, res.Response)
	assert.Equal(t, http.StatusTeapot, res.Response.StatusCode)
	assert.Contains(t, res.ResponseBody, "nope")

	// A placeholder without a test value fails before anything is sent.
	res = runTest(&HTTPRequestData{URL: srv.URL + "/{{interaction.user.id}}"}, nil, nil)
	assert.Equal(t, HTTPRequestTestStageBuild, res.ErrorStage)
	assert.Nil(t, res.Response)

	// So does an invalid JSON body.
	res = runTest(&HTTPRequestData{URL: srv.URL, BodyType: HTTPRequestBodyTypeJSON, Body: `{"a":}`}, nil, nil)
	assert.Equal(t, HTTPRequestTestStageBuild, res.ErrorStage)

	// Transform errors keep the response.
	res = runTest(&HTTPRequestData{URL: srv.URL, ResponseTransform: "response.data().a.b.c"}, nil, nil)
	assert.Equal(t, HTTPRequestTestStageTransform, res.ErrorStage)
	assert.NotNil(t, res.Response)

	// Unreachable hosts fail while sending.
	res = runTest(&HTTPRequestData{URL: "http://127.0.0.1:1"}, nil, nil)
	assert.Equal(t, HTTPRequestTestStageSend, res.ErrorStage)

	res = runTest(nil, nil, nil)
	assert.Equal(t, HTTPRequestTestStageBuild, res.ErrorStage)

	// A secret the app doesn't have.
	res = runTest(&HTTPRequestData{URL: srv.URL + "/?k={{secrets.MISSING}}"}, nil, nil)
	assert.Equal(t, HTTPRequestTestStageBuild, res.ErrorStage)
	assert.Contains(t, res.Error, "MISSING")
}

// Secrets can't be read back once saved, so a test sends them but never shows
// them, even when the server echoes them.
func TestRunHTTPRequestTestRedactsSecrets(t *testing.T) {
	srv := echoServer(t)
	secrets := map[string]string{"API_KEY": "s3cr3t-value"}

	res := runTest(&HTTPRequestData{
		Method:            "POST",
		URL:               srv.URL + "/echo/{{secrets.API_KEY}}?key={{secrets.API_KEY}}",
		Headers:           []HTTPRequestDataKeyValue{{Key: "Authorization", Value: "Bearer {{secrets.API_KEY}}"}},
		BodyType:          HTTPRequestBodyTypeJSON,
		Body:              `{"token": "{{secrets.API_KEY}}"}`,
		ResponseTransform: "response.data().auth",
	}, map[string]string{"secrets.API_KEY": "overridden"}, secrets)

	require.Empty(t, res.Error)
	// The real secret was sent, the test value didn't replace it.
	assert.Contains(t, res.ResponseBody, "Bearer [secret]")

	shown, err := json.Marshal(res)
	require.NoError(t, err)
	assert.NotContains(t, string(shown), "s3cr3t")
	assert.NotContains(t, string(shown), "overridden")
	assert.Equal(t, "Bearer [secret]", res.RequestHeaders["Authorization"])
	assert.Equal(t, "Bearer [secret]", res.Response.Headers["X-Echo-Auth"])
	assert.JSONEq(t, `"Bearer [secret]"`, string(res.Result))
}

func TestApplyTestValues(t *testing.T) {
	state := NewFlowContextState()
	env := eval.Env{}
	applyTestValues(env, state, map[string]string{
		"user.id":          "1",
		"user":             "ignored",
		"nodes.abc.result": `{"a":[1,2]}`,
		"input('x')":       "true",
		"secrets.API_KEY":  "x",
		"not a key!":       "x",
	})

	user, ok := env["user"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, int64(1), user["id"])
	assert.False(t, state.GetNodeResult("abc").IsNil())

	input, ok := env["input"].(func(string) any)
	require.True(t, ok)
	assert.Equal(t, true, input("x"))
	assert.Nil(t, input("y"))

	assert.Len(t, env, 2, "only user and input are set")
}

func TestParseTestValue(t *testing.T) {
	assert.Equal(t, int64(42), parseTestValue("42"))
	assert.Equal(t, -1.5, parseTestValue("-1.5"))
	assert.Equal(t, true, parseTestValue(" true "))
	assert.Nil(t, parseTestValue("null"))
	assert.Equal(t, []any{1.0, "a"}, parseTestValue(`[1, "a"]`))
	// Discord IDs stay text so they aren't rounded.
	assert.Equal(t, "123456789012345678", parseTestValue("123456789012345678"))
	assert.Equal(t, "NaN", parseTestValue("NaN"))
	assert.Equal(t, "{not json", parseTestValue("{not json"))
}

func TestPreviewBody(t *testing.T) {
	text, cut, binary := previewBody([]byte("hi"))
	assert.Equal(t, "hi", text)
	assert.False(t, cut)
	assert.False(t, binary)

	_, _, binary = previewBody([]byte{0xff, 0xfe})
	assert.True(t, binary)

	text, cut, binary = previewBody([]byte(strings.Repeat("é", maxHTTPRequestTestPreview)))
	assert.True(t, cut)
	assert.False(t, binary, "a cut must not split a character")
	assert.LessOrEqual(t, len(text), maxHTTPRequestTestPreview)
}
