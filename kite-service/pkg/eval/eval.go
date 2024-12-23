package eval

import (
	"context"
	"fmt"
	"io"

	"github.com/expr-lang/expr"
	"github.com/valyala/fasttemplate"
)

const templateStartTag = "{{"
const templateEndTag = "}}"

func Eval(ctx context.Context, expression string, env any) (any, error) {
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

	return result, nil
}

func EvalTemplate(ctx context.Context, template string, env any) (string, error) {
	res, err := fasttemplate.ExecuteFuncStringWithErr(
		template,
		templateStartTag,
		templateEndTag,
		func(w io.Writer, tag string) (int, error) {
			res, err := Eval(ctx, tag, env)
			if err != nil {
				return 0, err
			}

			// TODO: do proper casting to string
			val := fmt.Sprintf("%v", res)

			return w.Write([]byte(val))
		},
	)
	if err != nil {
		return "", err
	}

	return res, nil
}
