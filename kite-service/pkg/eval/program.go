package eval

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/builtin"
	"github.com/expr-lang/expr/parser"
	"github.com/expr-lang/expr/vm"
)

// programs caches compiled expressions, which are most of the cost of
// evaluating one: about 30 µs of 32 on kite.onl's server for a short one.
// A compiled program only depends on the expression, the patchers and the
// types of the env's values, which decide how fields are read, see
// fieldPatcher. The values are read when the program runs.
var programs = sync.OnceValue(func() *ristretto.Cache {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 300_000,
		// Bytes, as estimated by programCost. The cache only keeps hashes of
		// keys.
		MaxCost: 64 << 20,
		// Otherwise ristretto adds the size of its own entry to the cost.
		IgnoreInternalCost: true,
		BufferItems:        64,
	})
	if err != nil {
		panic(err)
	}
	return cache
})

// programCost estimates the memory of a program, which grows with its
// expression: about 2 KB for user.mention, and up to 850 KB for 10,000
// characters of +xp+xp, as it keeps the parsed expression.
func programCost(expression string) int64 {
	return 1<<10 + 100*int64(len(expression))
}

func compile(expression string, c Context) (*vm.Program, error) {
	key := programKey(expression, c)
	if p, ok := programs().Get(key); ok {
		return p.(*vm.Program), nil
	}

	program, err := compileUncached(expression, c)
	if err != nil {
		return nil, err
	}
	programs().Set(key, program, programCost(expression))
	return program, nil
}

func compileUncached(expression string, c Context) (*vm.Program, error) {
	// Invalid expressions are left to expr.Compile, which explains the error.
	if tree, err := parser.Parse(expression); err == nil && checkCost(tree.Node) > maxCheckCost {
		return nil, ErrExpressionTooNested
	}

	opts := []expr.Option{
		expr.Env(c.Env),
		expr.AllowUndefinedVariables(),
		expr.WithContext("ctx"),
		expr.Timezone("UTC"),
		calculateFunc,
		durationFunc,
	}
	opts = append(opts, patchOptions(c.Patchers)...)

	return expr.Compile(expression, opts...)
}

// programKey is the expression with the types of the patchers and of the
// env's values by name. Patchers must not have settings, as only their type
// is part of the key.
func programKey(expression string, c Context) string {
	names := make([]string, 0, len(c.Env))
	for name := range c.Env {
		names = append(names, name)
	}
	slices.Sort(names)

	var b strings.Builder
	b.WriteString(expression)
	for _, p := range c.Patchers {
		b.WriteByte(0)
		writeType(&b, reflect.TypeOf(p))
	}
	b.WriteByte(0)
	for _, name := range names {
		b.WriteByte(0)
		b.WriteString(name)
		b.WriteByte('=')
		writeType(&b, reflect.TypeOf(c.Env[name]))
	}
	return b.String()
}

// writeType writes the address of t, which identifies it. Its name doesn't,
// as *eval.GuildEnv has no package path. Types are never freed.
func writeType(b *strings.Builder, t reflect.Type) {
	if t == nil {
		b.WriteString("nil")
		return
	}
	var buf [16]byte
	b.Write(strconv.AppendUint(buf[:0], uint64(reflect.ValueOf(t).Pointer()), 16))
}

// maxCheckCost bounds checkCost, so that compiling an expression takes at most
// about 60 ms on kite.onl's server. Compiling can't be stopped, not even
// by the timeout of the flow.
const maxCheckCost = 20_000

// checkCost estimates the work of expr's type checker for an expression. It
// tries the overloads of a builtin like replace or sort one after another and
// checks all arguments again for each, so its work multiplies with every level
// of nested builtins. Calls of other functions are checked once. Remove this
// once expr's checker reuses the types of a builtin's arguments.
func checkCost(node ast.Node) int {
	var children []ast.Node
	factor := 1

	switch n := node.(type) {
	case *ast.UnaryNode:
		children = []ast.Node{n.Node}
	case *ast.BinaryNode:
		children = []ast.Node{n.Left, n.Right}
	case *ast.ChainNode:
		children = []ast.Node{n.Node}
	case *ast.MemberNode:
		children = []ast.Node{n.Node, n.Property}
	case *ast.SliceNode:
		children = []ast.Node{n.Node, n.From, n.To}
	case *ast.CallNode:
		children = append([]ast.Node{n.Callee}, n.Arguments...)
	case *ast.BuiltinNode:
		children = n.Arguments
		if i, ok := builtin.Index[n.Name]; ok {
			factor = max(len(builtin.Builtins[i].Types), 1)
		}
	case *ast.PredicateNode:
		children = []ast.Node{n.Node}
	case *ast.VariableDeclaratorNode:
		children = []ast.Node{n.Value, n.Expr}
	case *ast.SequenceNode:
		children = n.Nodes
	case *ast.ConditionalNode:
		children = []ast.Node{n.Cond, n.Exp1, n.Exp2}
	case *ast.ArrayNode:
		children = n.Nodes
	case *ast.MapNode:
		children = n.Pairs
	case *ast.PairNode:
		children = []ast.Node{n.Key, n.Value}
	}

	cost := 1
	for _, child := range children {
		cost += factor * checkCost(child)
	}
	// Saturated, so a deep expression can't overflow it.
	return min(cost, maxCheckCost+1)
}
