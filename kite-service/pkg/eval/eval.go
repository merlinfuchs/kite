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

const templateStartTag = "{{"
const templateEndTag = "}}"

func Eval(ctx context.Context, expression string, env Env) (thing.Any, error) {
	env["ctx"] = ctx

	program, err := expr.Compile(
		expression,
		expr.Env(env),
		expr.AllowUndefinedVariables(),
		expr.WithContext("ctx"),
		expr.Timezone("UTC"),
	)
	if err != nil {
		return thing.Null, fmt.Errorf("eval error: %w", err)
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return thing.Null, fmt.Errorf("eval error: %w", err)
	}

	return thing.New(result), nil
}

func EvalTemplate(ctx context.Context, template string, env Env) (any, error) {
	template = strings.TrimSpace(template)
	if template == "" {
		return "", nil
	}

	// Special case when template only contains one placeholder
	// We can just evaluate the expression directly and return the result with the original type
	if strings.HasPrefix(template, templateStartTag) &&
		strings.HasSuffix(template, templateEndTag) &&
		strings.Count(template, templateStartTag) == 1 &&
		strings.Count(template, templateEndTag) == 1 {
		template = template[len(templateStartTag) : len(template)-len(templateEndTag)]
		res, err := Eval(ctx, template, env)
		if err != nil {
			return nil, err
		}

		return res, nil
	}

	res, err := fasttemplate.ExecuteFuncStringWithErr(
		template,
		templateStartTag,
		templateEndTag,
		func(w io.Writer, tag string) (int, error) {
			res, err := Eval(ctx, tag, env)
			if err != nil {
				return 0, err
			}

			// This will call the String() method if it exists
			val := fmt.Sprintf("%v", res)
			return w.Write([]byte(val))
		},
	)
	if err != nil {
		return "", err
	}

	return res, nil
}

func EvalTemplateToString(ctx context.Context, template string, env Env) (string, error) {
	res, err := EvalTemplate(ctx, template, env)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", res), nil
}
