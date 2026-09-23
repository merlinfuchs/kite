package eval

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func testContext() Context {
	// Env, not map[string]any: the named type is what expr's OpLoadFast used to
	// panic on, so every case has to exercise it.
	return Context{Env: Env{}}
}

func TestEvalRejectsOverlongExpression(t *testing.T) {
	// A long chain of additions is valid syntax, so it reaches expr.Compile
	// unless the length bound stops it first.
	expression := "1" + strings.Repeat("+1", MaxExpressionLength)

	_, err := Eval(context.Background(), expression, testContext())
	if err == nil {
		t.Fatal("expected error for overlong expression, got nil")
	}
	if !errors.Is(err, ErrExpressionTooLong) {
		t.Fatalf("expected ErrExpressionTooLong, got %v", err)
	}
}

func TestEvalAllowsExpressionAtLimit(t *testing.T) {
	// Padding with spaces keeps the expression trivial to compile while landing
	// exactly on the limit, so we verify the boundary is inclusive.
	expression := "1" + strings.Repeat(" ", MaxExpressionLength-1)
	if len(expression) != MaxExpressionLength {
		t.Fatalf("setup: got length %d, want %d", len(expression), MaxExpressionLength)
	}

	res, err := Eval(context.Background(), expression, testContext())
	if err != nil {
		t.Fatalf("expected expression at limit to be accepted, got %v", err)
	}
	if res.String() != "1" {
		t.Errorf("got %q, want %q", res.String(), "1")
	}
}

func TestEvalTemplateRejectsOverlongTemplate(t *testing.T) {
	template := strings.Repeat("a", MaxTemplateLength+1)

	_, err := EvalTemplate(context.Background(), template, testContext())
	if err == nil {
		t.Fatal("expected error for overlong template, got nil")
	}
	if !errors.Is(err, ErrTemplateTooLong) {
		t.Fatalf("expected ErrTemplateTooLong, got %v", err)
	}
}

func TestEvalTemplateRejectsOverlongPlaceholder(t *testing.T) {
	// The template itself is under MaxTemplateLength, but the placeholder inside
	// it exceeds MaxExpressionLength, so the bound in Eval must catch it.
	inner := "1" + strings.Repeat("+1", MaxExpressionLength)
	template := "prefix {{" + inner + "}} suffix"
	if len(template) > MaxTemplateLength {
		t.Fatalf("setup: template length %d exceeds MaxTemplateLength", len(template))
	}

	_, err := EvalTemplate(context.Background(), template, testContext())
	if err == nil {
		t.Fatal("expected error for overlong placeholder, got nil")
	}
	if !errors.Is(err, ErrExpressionTooLong) {
		t.Fatalf("expected ErrExpressionTooLong, got %v", err)
	}
}

// The single-placeholder fast path in EvalTemplate strips the delimiters and
// calls Eval directly, bypassing fasttemplate, so it needs its own coverage.
func TestEvalTemplateRejectsOverlongSinglePlaceholder(t *testing.T) {
	inner := "1" + strings.Repeat("+1", MaxExpressionLength)
	template := "{{" + inner + "}}"
	if len(template) > MaxTemplateLength {
		t.Fatalf("setup: template length %d exceeds MaxTemplateLength", len(template))
	}

	_, err := EvalTemplate(context.Background(), template, testContext())
	if err == nil {
		t.Fatal("expected error for overlong single placeholder, got nil")
	}
	if !errors.Is(err, ErrExpressionTooLong) {
		t.Fatalf("expected ErrExpressionTooLong, got %v", err)
	}
}

// Ordinary expressions must keep working; the bound is far above real usage.
func TestEvalAcceptsNormalExpressions(t *testing.T) {
	cases := []struct{ expression, want string }{
		{`1 + 2`, "3"},
		{`"a" + "b"`, "ab"},
		{`[1,2,3] | filter(# > 1) | len()`, "2"},
		{`max([1,5,3])`, "5"},
		{`upper("hi")`, "HI"},
		{`true ? "y" : "n"`, "y"},
	}

	for _, c := range cases {
		res, err := Eval(context.Background(), c.expression, testContext())
		if err != nil {
			t.Errorf("%s: unexpected error %v", c.expression, err)
			continue
		}
		if res.String() != c.want {
			t.Errorf("%s: got %q, want %q", c.expression, res.String(), c.want)
		}
	}
}

func TestEvalTemplateAcceptsNormalTemplates(t *testing.T) {
	res, err := EvalTemplate(context.Background(), `hello {{ "world" }}`, testContext())
	if err != nil {
		t.Fatal(err)
	}
	if res.String() != "hello world" {
		t.Errorf("got %q, want %q", res.String(), "hello world")
	}
}

// Every existing case above evaluates a self-contained expression, none of
// which reads an identifier out of the env. That is exactly the gap that let
// expr 1.17 land green: its OpLoadFast opcode type-asserts the env to
// map[string]any without a comma-ok, which panics for the named Env type and
// broke every template in the product that referenced a variable. Anything
// that pins the env-lookup path shut keeps that from recurring silently.
func TestEvalReadsFromEnv(t *testing.T) {
	c := Context{Env: Env{
		"name":  "world",
		"count": 3,
		"nested": map[string]any{
			"key": "value",
		},
	}}

	cases := []struct {
		expression string
		want       string
	}{
		{`name`, "world"},
		{`upper(name)`, "WORLD"},
		{`count + 1`, "4"},
		{`nested.key`, "value"},
		{`name == "world" ? "yes" : "no"`, "yes"},
	}

	for _, tc := range cases {
		res, err := Eval(context.Background(), tc.expression, c)
		if err != nil {
			t.Errorf("%s: unexpected error %v", tc.expression, err)
			continue
		}
		if res.String() != tc.want {
			t.Errorf("%s: got %q, want %q", tc.expression, res.String(), tc.want)
		}
	}
}

// Same guard for the template path, which is what user-authored flows actually
// go through.
func TestEvalTemplateReadsFromEnv(t *testing.T) {
	c := Context{Env: Env{"name": "world"}}

	res, err := EvalTemplate(context.Background(), `hello {{ name }}, bye {{ name }}`, c)
	if err != nil {
		t.Fatal(err)
	}
	if res.String() != "hello world, bye world" {
		t.Errorf("got %q, want %q", res.String(), "hello world, bye world")
	}
}

func TestEvalTemplateRejectsOverlongOutput(t *testing.T) {
	// Each placeholder is small enough to pass the expression bound and cheap
	// enough to stay inside expr's per-VM memory budget. The budgets are
	// per-placeholder, so only the output bound stops the total.
	const perPlaceholder = 20_000
	one := `{{repeat("x",20000)}}`
	template := strings.Repeat(one, (MaxTemplateOutputLength/perPlaceholder)+2)

	if len(template) > MaxTemplateLength {
		t.Fatalf("setup: template is %d bytes, over the input bound of %d",
			len(template), MaxTemplateLength)
	}

	_, err := EvalTemplate(context.Background(), template, testContext())
	if err == nil {
		t.Fatal("expected error for overlong template output, got nil")
	}
	if !errors.Is(err, ErrTemplateOutputTooLong) {
		t.Fatalf("expected ErrTemplateOutputTooLong, got %v", err)
	}
}

func TestEvalTemplateAllowsOutputUnderLimit(t *testing.T) {
	res, err := EvalTemplate(context.Background(), `a{{repeat("x",1000)}}b{{repeat("y",1000)}}`, testContext())
	if err != nil {
		t.Fatal(err)
	}
	if len(res.String()) != 2002 {
		t.Errorf("got %d bytes, want 2002", len(res.String()))
	}
}

// Appended variable values keep their spaces (#339).
func TestEvalTemplateKeepSpace(t *testing.T) {
	c := Context{Env: Env{"name": "world"}}

	for template, want := range map[string]string{
		" world":          " world",
		"hello {{name}} ": "hello world ",
		" ":               " ",
	} {
		res, err := EvalTemplateKeepSpace(context.Background(), template, c)
		if err != nil {
			t.Fatalf("%q: %v", template, err)
		}
		if got := res.String(); got != want {
			t.Errorf("EvalTemplateKeepSpace(%q) = %q, want %q", template, got, want)
		}
	}
}

// A lone placeholder keeps its value's type even when surrounded by spaces.
func TestEvalTemplateKeepSpaceSinglePlaceholder(t *testing.T) {
	res, err := EvalTemplateKeepSpace(context.Background(), " {{1 + 1}} ", testContext())
	if err != nil {
		t.Fatal(err)
	}
	if res.Int() != 2 || res.String() != "2" {
		t.Errorf("got %q, want the placeholder's value", res.String())
	}
}

func TestResumeContextFallsBackToEarlierInteractions(t *testing.T) {
	origin := Context{Env: Env{
		"user": "origin-user",
		"arg":  func(name string) any { return "origin-" + name },
	}}
	previous := Context{Env: Env{"user": "previous-user"}}

	c := Context{Env: Env{
		"user":  "clicker",
		"arg":   func(name string) any { return nil },
		"input": func(customID string) any { return nil },
	}}
	c.SetResumeContext(origin, previous, []map[string]string{
		{"name": "newest"},
		{"name": "oldest", "age": "30"},
	})

	cases := map[string]string{
		`user`:              "clicker",
		`origin.user`:       "origin-user",
		`previous.user`:     "previous-user",
		`arg("reason")`:     "origin-reason",
		`input("name")`:     "newest",
		`input("age")`:      "30",
		`origin.arg("why")`: "origin-why",
	}
	for expression, want := range cases {
		res, err := Eval(context.Background(), expression, c)
		if err != nil {
			t.Fatalf("%s: %v", expression, err)
		}
		if res.String() != want {
			t.Errorf("%s = %q, want %q", expression, res.String(), want)
		}
	}
}

func TestResumeContextPrefersCurrentModalInput(t *testing.T) {
	c := Context{Env: Env{
		"input": func(customID string) any { return "current" },
	}}
	c.SetResumeContext(Context{Env: Env{}}, Context{Env: Env{}}, []map[string]string{{"name": "earlier"}})

	res, err := Eval(context.Background(), `input("name")`, c)
	if err != nil {
		t.Fatal(err)
	}
	if res.String() != "current" {
		t.Errorf("input(\"name\") = %q, want %q", res.String(), "current")
	}
}
