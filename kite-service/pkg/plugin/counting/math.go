package counting

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/expr-lang/expr"
)

// maxExpressionLength bounds a counting message that is evaluated as math.
// Counting messages come from any member of the server, so they are kept far
// below what a flow expression may be.
const maxExpressionLength = 200

// mathCharset is everything a math message may contain. It leaves out quotes,
// so there are no strings, and comparison and ternary operators, so a message
// can only ever be arithmetic.
var mathCharset = regexp.MustCompile(`^[0-9a-z\s+\-*/%^().,]+$`)

var errNotMath = errors.New("not a math expression")

var mathBuiltins = []string{"abs", "ceil", "floor", "round", "max", "min"}

var mathOptions = func() []expr.Option {
	opts := []expr.Option{
		expr.Env(map[string]any{
			"pi": math.Pi,
			"e":  math.E,
		}),
		expr.DisableAllBuiltins(),
		mathFunction("sqrt", 1, func(a []float64) float64 { return math.Sqrt(a[0]) }),
		mathFunction("cbrt", 1, func(a []float64) float64 { return math.Cbrt(a[0]) }),
		mathFunction("pow", 2, func(a []float64) float64 { return math.Pow(a[0], a[1]) }),
		mathFunction("log", 1, func(a []float64) float64 { return math.Log(a[0]) }),
		mathFunction("log2", 1, func(a []float64) float64 { return math.Log2(a[0]) }),
		mathFunction("log10", 1, func(a []float64) float64 { return math.Log10(a[0]) }),
		mathFunction("sin", 1, func(a []float64) float64 { return math.Sin(a[0]) }),
		mathFunction("cos", 1, func(a []float64) float64 { return math.Cos(a[0]) }),
		mathFunction("tan", 1, func(a []float64) float64 { return math.Tan(a[0]) }),
	}
	for _, name := range mathBuiltins {
		opts = append(opts, expr.EnableBuiltin(name))
	}
	return opts
}()

func mathFunction(name string, args int, fn func([]float64) float64) expr.Option {
	return expr.Function(name, func(params ...any) (any, error) {
		if len(params) != args {
			return nil, fmt.Errorf("%s takes %d arguments", name, args)
		}

		values := make([]float64, len(params))
		for i, p := range params {
			v, ok := toFloat(p)
			if !ok {
				return nil, fmt.Errorf("%s takes numbers", name)
			}
			values[i] = v
		}

		return fn(values), nil
	})
}

// parseCount reads a counting message as a number. A message that is just a
// number is that number. Otherwise it is evaluated as a math expression, like
// "5+5" or "sqrt(81)+1". isExpression reports whether it was an expression,
// and errNotMath means the message isn't a counting attempt at all, like
// normal chat, so it must not break the count.
func parseCount(content string) (value float64, isExpression bool, err error) {
	content = strings.TrimSpace(content)

	if num, err := strconv.ParseInt(content, 10, 64); err == nil {
		return float64(num), false, nil
	}

	if len(content) > maxExpressionLength {
		return 0, false, errNotMath
	}

	content = strings.ToLower(content)
	content = strings.NewReplacer("×", "*", "÷", "/", "−", "-").Replace(content)

	// Without a digit it's a word, not math. ".." is expr's range operator,
	// which allocates the whole range.
	if !mathCharset.MatchString(content) ||
		!strings.ContainsAny(content, "0123456789") ||
		strings.Contains(content, "..") {
		return 0, false, errNotMath
	}

	program, err := expr.Compile(content, mathOptions...)
	if err != nil {
		return 0, false, errNotMath
	}

	result, err := expr.Run(program, map[string]any{
		"pi": math.Pi,
		"e":  math.E,
	})
	if err != nil {
		return 0, false, errNotMath
	}

	v, ok := toFloat(result)
	if !ok {
		return 0, false, errNotMath
	}

	return v, true, nil
}

// matchesCount reports whether value is the integer n, allowing for float
// rounding like 0.1*30.
func matchesCount(value float64, n int64) bool {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return false
	}
	return math.Abs(value-float64(n)) < 1e-9
}

// formatCount formats an evaluated value for the ruined message.
func formatCount(value float64) string {
	if value == math.Trunc(value) && math.Abs(value) < 1e15 {
		return strconv.FormatInt(int64(value), 10)
	}
	return strconv.FormatFloat(value, 'g', 10, 64)
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}
