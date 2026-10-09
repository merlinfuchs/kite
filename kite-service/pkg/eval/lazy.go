package eval

import (
	"reflect"

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

func fieldOptions() []expr.Option {
	return []expr.Option{
		expr.Function(fetchFieldFunc, fetchField),
		expr.Patch(fieldPatcher{}),
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
	if _, ok := member.Property.(*ast.StringNode); !ok {
		return
	}

	t := member.Node.Type()
	if t != nil && t.Kind() != reflect.Interface && !t.Implements(lazyFielderType) {
		return
	}

	ast.Patch(node, &ast.CallNode{
		Callee:    &ast.IdentifierNode{Value: fetchFieldFunc},
		Arguments: []ast.Node{member.Node, member.Property, &ast.BoolNode{Value: member.Optional}},
	})
}

func fetchField(params ...any) (any, error) {
	from, name, optional := params[0], params[1].(string), params[2].(bool)

	if runtime.IsNil(from) {
		if optional {
			return nil, nil
		}
		// Outside of a server the server is nil, and its fields are empty
		// like the member fields of the user, instead of failing the flow.
		v := reflect.ValueOf(from)
		if v.Kind() != reflect.Ptr || !v.Type().Implements(lazyFielderType) {
			return runtime.Fetch(from, name), nil
		}
		from = reflect.New(v.Type().Elem()).Interface()
	}

	if l, ok := from.(lazyFielder); ok {
		if v, ok, err := l.lazyField(name); ok {
			return v, err
		}
	}
	return runtime.Fetch(from, name), nil
}
