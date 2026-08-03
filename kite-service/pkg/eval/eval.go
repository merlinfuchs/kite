package eval

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/valyala/fasttemplate"
)

const (
	templateStartTag = "{{"
	templateEndTag   = "}}"

	// MaxExpressionLength bounds a single expression handed to expr.Compile.
	// The compiler allocates AST nodes proportional to the input, so an
	// unbounded expression lets a flow author exhaust memory at execution time.
	//
	// The value is deliberately loose. Closing the parser blowup only requires
	// keeping input out of the megabytes, so anything in the low thousands would
	// do; the limit is set well above realistic usage (Discord caps message
	// content at 2000 and a whole embed at 6000) so that no flow already stored
	// in the database starts failing at runtime.
	MaxExpressionLength = 10_000

	// MaxTemplateLength bounds a whole template before its placeholders are
	// evaluated. Each placeholder is bounded by MaxExpressionLength, but a
	// template may contain arbitrarily many of them.
	MaxTemplateLength = 100_000

	// MaxTemplateOutputLength bounds what a template expands to, which the
	// input bounds above do not imply. Every placeholder is evaluated by its
	// own expr VM with its own memory budget, so the budgets never see each
	// other and the results all accumulate into one buffer. A legal 100k
	// template of ~4700 placeholders each emitting ~1MB expands to several GB
	// for the cost of a single flow operation, which template evaluation is
	// not metered as.
	//
	// Tied to the HTTP response body cap, since the case this has to keep
	// working is interpolating a fetched body into a larger string. A template
	// that is nothing but one placeholder returns before it gets here, so
	// passing a body straight through is unaffected either way.
	MaxTemplateOutputLength = thing.MaxBodySize
)

// ErrExpressionTooLong is returned when an expression exceeds its length limit.
// Flows are user-authored, so this is reachable from user input.
var ErrExpressionTooLong = errors.New("expression too long")

// ErrTemplateTooLong is returned when a template exceeds MaxTemplateLength.
var ErrTemplateTooLong = errors.New("template too long")

// ErrTemplateOutputTooLong is returned when a template's placeholders expand
// past MaxTemplateOutputLength. Like ErrExpressionTooLong this is reachable
// from user input, so it is an ordinary flow error rather than a panic.
var ErrTemplateOutputTooLong = errors.New("template output too long")

func Eval(ctx context.Context, expression string, c Context) (thing.Thing, error) {
	if len(expression) > MaxExpressionLength {
		return thing.Null, fmt.Errorf(
			"eval error: %w: %d characters, limit is %d",
			ErrExpressionTooLong, len(expression), MaxExpressionLength,
		)
	}

	c.Env["ctx"] = proxyContext{ctx: ctx}

	opts := []expr.Option{
		expr.Env(c.Env),
		expr.AllowUndefinedVariables(),
		expr.WithContext("ctx"),
		expr.Timezone("UTC"),
	}
	for _, p := range c.Patchers {
		opts = append(opts, expr.Patch(p))
	}

	program, err := expr.Compile(expression, opts...)
	if err != nil {
		return thing.Null, fmt.Errorf("eval error: %w", err)
	}

	// The env has to be handed to Run as a plain map[string]any, not as Env.
	// expr's OpLoadFast opcode type-asserts the env to map[string]any without
	// a comma-ok, so a named map type panics on every identifier lookup and
	// surfaces as "interface conversion: interface {} is eval.Env". The
	// conversion is free -- Env's underlying type is already map[string]any.
	result, err := expr.Run(program, map[string]any(c.Env))
	if err != nil {
		return thing.Null, fmt.Errorf("eval error: %w", err)
	}

	res := thing.NewGuessTypeWithFallback(result)
	return replaceNewlines(res), nil
}

func EvalTemplate(ctx context.Context, template string, c Context) (thing.Thing, error) {
	if len(template) > MaxTemplateLength {
		return thing.Null, fmt.Errorf(
			"eval error: %w: %d characters, limit is %d",
			ErrTemplateTooLong, len(template), MaxTemplateLength,
		)
	}

	template = strings.TrimSpace(template)
	if template == "" {
		return thing.Null, nil
	}

	// Special case when template only contains one placeholder
	// We can just evaluate the expression directly and return the result with the original type
	if strings.HasPrefix(template, templateStartTag) &&
		strings.HasSuffix(template, templateEndTag) &&
		strings.Count(template, templateStartTag) == 1 &&
		strings.Count(template, templateEndTag) == 1 {
		template = template[len(templateStartTag) : len(template)-len(templateEndTag)]
		res, err := Eval(ctx, template, c)
		if err != nil {
			return thing.Null, err
		}

		return res, nil
	}

	// Literal spans between placeholders are already covered by the input
	// bound, so only what the placeholders expand to has to be tracked.
	var emitted int

	res, err := fasttemplate.ExecuteFuncStringWithErr(
		template,
		templateStartTag,
		templateEndTag,
		func(w io.Writer, tag string) (int, error) {
			res, err := Eval(ctx, tag, c)
			if err != nil {
				return 0, err
			}

			if res.IsNil() {
				return w.Write([]byte(""))
			}

			// This will call the String() method if it exists
			val := fmt.Sprintf("%v", res)

			emitted += len(val)
			if emitted > MaxTemplateOutputLength {
				return 0, fmt.Errorf(
					"eval error: %w: expanded to at least %d bytes, limit is %d",
					ErrTemplateOutputTooLong, emitted, MaxTemplateOutputLength,
				)
			}

			return w.Write([]byte(val))
		},
	)
	if err != nil {
		return thing.Null, err
	}

	return thing.NewString(res), nil
}

// replaceNewlines replaces \n with actual newlines in a string
// It also replaces \\n with \n so escaping is preserved
func replaceNewlines(s thing.Thing) thing.Thing {
	if s.Type == thing.TypeString {
		str := s.String()
		var result strings.Builder
		result.Grow(len(str))

		runes := []rune(str)
		for i := 0; i < len(runes); i++ {
			// Check for escaped backslash followed by n
			if i < len(runes)-2 && runes[i] == '\\' && runes[i+1] == '\\' && runes[i+2] == 'n' {
				result.WriteString("\\n")
				i += 2 // Skip the next two runes
			} else if i < len(runes)-1 && runes[i] == '\\' && runes[i+1] == 'n' {
				// Regular \n case
				result.WriteString("\n")
				i++ // Skip the next rune
			} else {
				result.WriteRune(runes[i])
			}
		}
		return thing.NewString(result.String())
	}
	return s
}

func EvalTemplateToString(ctx context.Context, template string, c Context) (string, error) {
	res, err := EvalTemplate(ctx, template, c)
	if err != nil {
		return "", err
	}

	if res.IsNil() {
		return "", nil
	}

	return fmt.Sprintf("%v", res), nil
}
