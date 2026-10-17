package eval

import (
	"context"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// Results of blocks are read through node(), which returns them untyped, like
// {{nodes.id.result.data.name}} in a flow.
func TestHTTPResponsePlaceholders(t *testing.T) {
	results := map[string]thing.Thing{
		"array": thing.NewHTTPResponse(thing.HTTPResponseValue{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       []byte(`[{"q":"Be kind","a":"Someone"}]`),
		}),
		"object": thing.NewHTTPResponse(thing.HTTPResponseValue{
			Status:     "200 OK",
			StatusCode: 200,
			Body:       []byte(`{"setup":"Why?","punchline":"Because","count":3}`),
		}),
	}

	c := Context{Env: Env{
		"node": func(id string) (any, error) {
			return map[string]any{"result": NewThingEnv(results[id])}, nil
		},
	}}

	cases := map[string]string{
		"{{node('array').result.data[0].q}}":         "Be kind",
		"— {{node('array').result.data[0].a}}":       "— Someone",
		"{{node('object').result.data.punchline}}":   "Because",
		"{{node('object').result.data.count + 1}}":   "4",
		"{{node('object').result.data().setup}}":     "Why?",
		"{{node('object').result.body}}":             `{"setup":"Why?","punchline":"Because","count":3}`,
		"{{node('object').result.status_code}}":      "200",
		"{{node('object').result.data?.missing?.x}}": "",
	}
	for template, want := range cases {
		got, err := EvalTemplate(context.Background(), template, c)
		if err != nil {
			t.Errorf("%s: %v", template, err)
			continue
		}
		if got.String() != want {
			t.Errorf("%s = %q, want %q", template, got.String(), want)
		}
	}
}

func TestHTTPResponseDataNotJSON(t *testing.T) {
	c := Context{Env: Env{
		"node": func(id string) (any, error) {
			return map[string]any{"result": NewThingEnv(thing.NewHTTPResponse(thing.HTTPResponseValue{
				StatusCode: 200,
				Body:       []byte("not json"),
			}))}, nil
		},
	}}

	if _, err := EvalTemplate(context.Background(), "{{node('x').result.data.name}}", c); err == nil {
		t.Error("reading data of a body that isn't JSON should fail")
	}
}
