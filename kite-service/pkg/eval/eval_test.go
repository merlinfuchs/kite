package eval

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func testContext() Context {
	return Context{Env: map[string]any{}}
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
	if !errors.Is(err, ErrExpressionTooLong) {
		t.Fatalf("expected ErrExpressionTooLong, got %v", err)
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
