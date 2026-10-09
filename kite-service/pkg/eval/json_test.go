package eval

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func TestExpandJSONTemplate(t *testing.T) {
	values := map[string]thing.Thing{
		"name":   thing.NewString(`Jack "the dev" \ o/`),
		"multi":  thing.NewString("line one\nline two"),
		"num":    thing.NewInt(42),
		"float":  thing.NewFloat(1.5),
		"bool":   thing.NewBool(true),
		"nil":    thing.Null,
		"list":   thing.NewArray([]thing.Thing{thing.NewInt(1), thing.NewString("two")}),
		"object": thing.NewAny(map[string]any{"a": 1}),
		"html":   thing.NewString("<b>&</b>"),
		"resp": thing.NewHTTPResponse(thing.HTTPResponseValue{
			StatusCode: 200,
			Body:       []byte(`{"ok":true}`),
		}),
	}

	evalFn := func(expression string) (thing.Thing, error) {
		v, ok := values[expression]
		if !ok {
			return thing.Null, errors.New("unknown " + expression)
		}
		return v, nil
	}

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"no placeholders", `{"a": 1}`, `{"a": 1}`},
		{"escaped in string", `{"content": "Hi {{name}}!"}`, `{"content": "Hi Jack \"the dev\" \\ o/!"}`},
		{"newline in string", `{"c": "{{multi}}"}`, `{"c": "line one\nline two"}`},
		{"number in string stays string", `{"c": "{{num}}"}`, `{"c": "42"}`},
		{"typed number", `{"c": {{num}}}`, `{"c": 42}`},
		{"typed float", `{"c": {{float}}}`, `{"c": 1.5}`},
		{"typed bool", `{"c": {{bool}}}`, `{"c": true}`},
		{"typed null", `{"c": {{nil}}}`, `{"c": null}`},
		{"typed string", `{"c": {{name}}}`, `{"c": "Jack \"the dev\" \\ o/"}`},
		{"typed array", `{"c": {{list}}}`, `{"c": [1,"two"]}`},
		{"typed object", `{"c": {{object}}}`, `{"c": {"a":1}}`},
		{"no html escaping", `{"c": "{{html}}"}`, `{"c": "<b>&</b>"}`},
		{"http response passthrough", `{"c": {{resp}}}`, `{"c": {"ok":true}}`},
		{"quote after escaped quote", `{"a": "x\"", "b": {{num}}}`, `{"a": "x\"", "b": 42}`},
		{"placeholder key", `{"{{name}}": 1}`, `{"Jack \"the dev\" \\ o/": 1}`},
		{"top level array", `[{{num}}, "{{num}}"]`, `[42, "42"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandJSONTemplate(tt.template, evalFn)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
			if !json.Valid(got) {
				t.Fatalf("result is not valid JSON: %s", got)
			}
		})
	}
}

func TestExpandJSONTemplateUnclosed(t *testing.T) {
	_, err := expandJSONTemplate(`{"a": "{{name"}`, func(string) (thing.Thing, error) {
		return thing.Null, nil
	})
	if err == nil {
		t.Fatal("expected an error for an unclosed placeholder")
	}
}

func TestEvalJSONTemplate(t *testing.T) {
	c := NewContext(Env{
		"name":  `Jack "the dev"`,
		"count": 3,
		"list":  []any{1, "two"},
		"user":  NewUserEnv(discord.User{ID: 123456789012345678, Username: "jack"}),
	})

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{"string and number", `{"content": "Hi {{name}}", "n": {{count}}}`, `{"content": "Hi Jack \"the dev\"", "n": 3}`},
		{"expression", `{"n": {{count * 2}}, "s": "{{count * 2}}"}`, `{"n": 6, "s": "6"}`},
		{"list", `{"l": {{list}}}`, `{"l": [1,"two"]}`},
		{"list in string", `{"l": "{{list}}"}`, `{"l": "[1,\"two\"]"}`},
		{"missing value is null", `{"x": {{missing}}}`, `{"x": null}`},
		{"missing value in string is empty", `{"x": "{{missing}}"}`, `{"x": ""}`},
		{"discord object is its id", `{"user": {{user}}}`, `{"user": "123456789012345678"}`},
		{"discord object in string is its mention", `{"user": "{{user}}"}`, `{"user": "<@123456789012345678>"}`},
		{"string literal in expression", `{"s": "{{"a}" + "b"}}"}`, `{"s": "a}b"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalJSONTemplate(context.Background(), tt.template, c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEvalJSONTemplateEmpty(t *testing.T) {
	for _, template := range []string{"", "  \n "} {
		got, err := EvalJSONTemplate(context.Background(), template, NewContext(Env{}))
		if err != nil || got != nil {
			t.Fatalf("%q: got %s, %v, want no body", template, got, err)
		}
	}
}

func TestEvalJSONTemplateErrors(t *testing.T) {
	c := NewContext(Env{"text": "a, b"})

	tests := []struct {
		name     string
		template string
		want     error
	}{
		{"invalid template", `{"a": }`, ErrInvalidJSONTemplate},
		// Text outside quotes is inserted as a JSON string, so the result
		// is only invalid when the document around it is.
		{"invalid around placeholder", `{"a": {{text}} "b": 1}`, ErrInvalidJSONTemplate},
		{"too long", `"` + strings.Repeat("a", MaxTemplateLength) + `"`, ErrTemplateTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EvalJSONTemplate(context.Background(), tt.template, c)
			if !errors.Is(err, tt.want) {
				t.Fatalf("got %v, want %v", err, tt.want)
			}
		})
	}

	_, err := EvalJSONTemplate(context.Background(), `{"a": {{1 +}}}`, c)
	if err == nil {
		t.Fatal("expected the expression error")
	}
}

// Every placeholder is bounded on its own, so the total is checked too.
func TestExpandJSONTemplateOutputTooLong(t *testing.T) {
	big := thing.NewString(strings.Repeat("b", MaxTemplateOutputLength/2+1))
	_, err := expandJSONTemplate(`["{{a}}", "{{a}}"]`, func(string) (thing.Thing, error) {
		return big, nil
	})
	if !errors.Is(err, ErrTemplateOutputTooLong) {
		t.Fatalf("got %v, want %v", err, ErrTemplateOutputTooLong)
	}
}

func TestThingToJSON(t *testing.T) {
	tests := []struct {
		name  string
		value thing.Thing
		want  string
	}{
		{"null", thing.Null, `null`},
		{"string", thing.NewString(`a "b" <c>`), `"a \"b\" <c>"`},
		{"int", thing.NewInt(7), `7`},
		{"bool", thing.NewBool(false), `false`},
		{"nested", thing.NewArray([]thing.Thing{thing.NewObject(map[string]thing.Thing{"a": thing.NewInt(1)})}), `[{"a":1}]`},
		{"discord user", thing.NewDiscordUser(discord.User{ID: 42}), `"42"`},
		{"json response", thing.NewHTTPResponse(thing.HTTPResponseValue{Body: []byte(` {"ok": true} `)}), `{"ok": true}`},
		{"text response", thing.NewHTTPResponse(thing.HTTPResponseValue{Body: []byte("plain")}), `"plain"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ThingToJSON(tt.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestTemplateString(t *testing.T) {
	if got := templateString(thing.Null); got != "" {
		t.Fatalf("null: got %q", got)
	}
	if got := templateString(thing.NewAny(map[string]any{"a": 1})); got != `{"a":1}` {
		t.Fatalf("map: got %q", got)
	}
	if got := templateString(thing.NewArray([]thing.Thing{thing.NewInt(1)})); got != `[1]` {
		t.Fatalf("array: got %q", got)
	}
	if got := templateString(thing.NewFloat(1.5)); got != "1.5" {
		t.Fatalf("float: got %q", got)
	}
}
