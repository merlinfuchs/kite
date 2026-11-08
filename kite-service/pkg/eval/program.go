package eval

import (
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto"
	"github.com/expr-lang/expr"
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
