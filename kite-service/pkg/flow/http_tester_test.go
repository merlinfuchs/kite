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
)

type testClientProvider struct{}

func (testClientProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req.WithContext(ctx))
}

func runTest(data *HTTPRequestData, values map[string]string) HTTPRequestTestResult {
	return RunHTTPRequestTest(context.Background(), HTTPRequestTestOpts{
		Data:    data,
		Values:  values,
		HTTP:    testClientProvider{},
		Timeout: 5 * time.Second,
	})
}

func TestRunHTTPRequestTest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/fail" {
			w.WriteHeader(http.StatusTeapot)
			w.Write([]byte(`{"error":"nope"}`))
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"path":   r.URL.Path,
			"query":  r.URL.Query().Get("q"),
			"header": r.Header.Get("X-Test"),
			"body":   string(body),
		})
	}))
	defer srv.Close()

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
		"var('token')":    "secret",
		"result('count')": "5",
	})

	if res.Error != "" {
		t.Fatalf("unexpected error at %s: %s", res.ErrorStage, res.Error)
	}
	if res.Method != "POST" || !strings.HasSuffix(strings.Split(res.URL, "?")[0], "/users/123456789012345678") {
		t.Fatalf("unexpected request line: %s %s", res.Method, res.URL)
	}
	if res.RequestHeaders["Content-Type"] != "application/json" {
		t.Fatalf("content type not set: %v", res.RequestHeaders)
	}
	if res.RequestBody != `{"n": 5, "name": "a \"quoted\" name"}` {
		t.Fatalf("unexpected body: %s", res.RequestBody)
	}
	if res.Response == nil || res.Response.StatusCode != 200 {
		t.Fatalf("unexpected response: %+v", res.Response)
	}
	if !strings.Contains(res.ResponseBody, `"query":"hello world"`) ||
		!strings.Contains(res.ResponseBody, `"header":"secret"`) {
		t.Fatalf("placeholders not applied: %s", res.ResponseBody)
	}
	if string(res.Result) != `"/users/123456789012345678"` {
		t.Fatalf("unexpected transform result: %s", res.Result)
	}

	// The response is kept when fail on error status stops the request.
	res = runTest(&HTTPRequestData{URL: srv.URL + "/fail", FailOnErrorStatus: true}, nil)
	if res.ErrorStage != HTTPRequestTestStageStatus || res.Response == nil ||
		res.Response.StatusCode != http.StatusTeapot || !strings.Contains(res.ResponseBody, "nope") {
		t.Fatalf("unexpected status failure: %+v", res)
	}

	// A missing test value fails while building, before anything is sent.
	res = runTest(&HTTPRequestData{URL: srv.URL + "/{{interaction.user.id}}"}, nil)
	if res.ErrorStage != HTTPRequestTestStageBuild || res.Response != nil {
		t.Fatalf("expected build failure, got %+v", res)
	}

	// Transform errors keep the response too.
	res = runTest(&HTTPRequestData{URL: srv.URL, ResponseTransform: "response.data().a.b.c"}, nil)
	if res.ErrorStage != HTTPRequestTestStageTransform || res.Response == nil {
		t.Fatalf("expected transform failure, got %+v", res)
	}
}

func TestApplyTestValues(t *testing.T) {
	state := NewFlowContextState()
	env := map[string]any{}
	applyTestValues(env, state, map[string]string{
		"user.id":          "1",
		"user":             "ignored",
		"nodes.abc.result": `{"a":[1,2]}`,
		"input('x')":       "true",
		"not a key!":       "x",
	})

	user, ok := env["user"].(map[string]any)
	if !ok || user["id"] != int64(1) {
		t.Fatalf("unexpected user: %#v", env["user"])
	}
	if state.GetNodeResult("abc").IsNil() {
		t.Fatal("node result not set")
	}
	input, ok := env["input"].(func(string) any)
	if !ok || input("x") != true || input("y") != nil {
		t.Fatal("input lookup not set")
	}
	if len(env) != 2 {
		t.Fatalf("unexpected env keys: %v", env)
	}

	if parseTestValue("123456789012345678") != "123456789012345678" {
		t.Fatal("large numbers must stay text")
	}
	if parseTestValue("NaN") != "NaN" {
		t.Fatal("NaN must stay text")
	}
}
