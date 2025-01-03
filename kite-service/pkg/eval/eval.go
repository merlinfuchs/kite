package eval

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/valyala/fasttemplate"
)

const templateStartTag = "{{"
const templateEndTag = "}}"

func Eval(ctx context.Context, expression string, env Env) (any, error) {
	fmt.Println("eval expression: ", expression)

	env["ctx"] = ctx

	program, err := expr.Compile(
		expression,
		expr.Env(env),
		expr.AllowUndefinedVariables(),
		expr.WithContext("ctx"),
		expr.Timezone("UTC"),
	)
	if err != nil {
		return nil, err
	}

	result, err := expr.Run(program, env)
	if err != nil {
		return nil, err
	}

	fmt.Printf("eval result: %T\n", result)
	return result, nil
}

func EvalTemplate(ctx context.Context, template string, env Env) (any, error) {
	// Special case when template only contains one placeholder
	if strings.Count(template, templateStartTag) == 1 && strings.Count(template, templateEndTag) == 1 {
		template = template[len(templateStartTag) : len(template)-len(templateEndTag)]
		res, err := Eval(ctx, template, env)
		if err != nil {
			return nil, err
		}

		fmt.Printf("eval template result: %T\n", res)
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
