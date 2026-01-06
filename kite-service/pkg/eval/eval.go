package eval

import (
	"context"
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
)

func Eval(ctx context.Context, expression string, c Context) (thing.Thing, error) {
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

	result, err := expr.Run(program, c.Env)
	if err != nil {
		return thing.Null, fmt.Errorf("eval error: %w", err)
	}

	res := thing.NewGuessTypeWithFallback(result)
	return replaceNewlines(res), nil
}

func EvalTemplate(ctx context.Context, template string, c Context) (thing.Thing, error) {
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
