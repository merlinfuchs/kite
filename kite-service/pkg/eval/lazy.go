package eval

import (
	"reflect"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm/runtime"
)

// lazyFielder is implemented by envs with fields that are only computed when
// an expression reads them, like the channel count of a server, which takes
// all of its channels from the cache. expr reads struct fields directly, so
// fieldPatcher turns reading a field into a call of fetchField, which asks
// lazyField first.
type lazyFielder interface {
	// lazyField returns false for fields it doesn't compute.
	lazyField(name string) (any, bool, error)
}

var lazyFielderType = reflect.TypeOf((*lazyFielder)(nil)).Elem()

// fetchFieldFunc can't be written in an expression, only fieldPatcher adds it.
const fetchFieldFunc = "$fetch_field"

// patchOptions adds fieldPatcher after patchers. They run as one, as expr
// checks the types of the whole expression again before each patcher.
func patchOptions(patchers []ast.Visitor) []expr.Option {
	return []expr.Option{
		expr.Function(fetchFieldFunc, fetchField),
		expr.Patch(visitors(append(slices.Clip(patchers), fieldPatcher{}))),
	}
}

type visitors []ast.Visitor

func (v visitors) Visit(node *ast.Node) {
	for _, visitor := range v {
		visitor.Visit(node)
	}
}

// fieldPatcher patches reading a field of an env with lazy fields, or of a
// value that isn't known before the expression runs, like origin.guild.
type fieldPatcher struct{}

func (fieldPatcher) Visit(node *ast.Node) {
	member, ok := (*node).(*ast.MemberNode)
	if !ok || member.Method {
		return
	}

	_, named := member.Property.(*ast.StringNode)
	t := member.Node.Type()
	lazy := t == nil || t.Kind() == reflect.Interface || t.Implements(lazyFielderType)
	// Patching takes away the jump of an optional field to the end of its
	// chain, so what follows it is patched too, to be nil in a?.b.c or a?.b[0]
	// if a is nil.
	afterOptional := isOptionalFetch(member.Node)
	if !afterOptional && !(named && lazy) {
		return
	}

	ast.Patch(node, &ast.CallNode{
		Callee:    &ast.IdentifierNode{Value: fetchFieldFunc},
		Arguments: []ast.Node{member.Node, member.Property, &ast.BoolNode{Value: member.Optional || afterOptional}},
	})
}

func isOptionalFetch(node ast.Node) bool {
	call, ok := node.(*ast.CallNode)
	if !ok {
		return false
	}
	if callee, ok := call.Callee.(*ast.IdentifierNode); !ok || callee.Value != fetchFieldFunc {
		return false
	}
	optional, ok := call.Arguments[2].(*ast.BoolNode)
	return ok && optional.Value
}

func fetchField(params ...any) (any, error) {
	from, name, optional := params[0], params[1], params[2].(bool)

	if runtime.IsNil(from) {
		if optional {
			return nil, nil
		}
		// Outside of a server the server is nil, and its fields are empty
		// like the member fields of the user, instead of failing the flow.
		if v := reflect.ValueOf(from); v.Kind() == reflect.Ptr && v.Type().Implements(lazyFielderType) {
			from = reflect.New(v.Type().Elem()).Interface()
		}
	}

	if l, ok := from.(lazyFielder); ok {
		if name, ok := name.(string); ok {
			if v, ok, err := l.lazyField(name); ok {
				return v, err
			}
		}
	}
	return runtime.Fetch(from, name), nil
}
