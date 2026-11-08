package eval

import (
	"context"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/expr-lang/expr/ast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A cached program reads the values of the env it runs with, not the one it
// was compiled with.
func TestCachedProgramUsesNewValues(t *testing.T) {
	for _, name := range []string{"a", "b"} {
		env := Env{
			"user": NewUserEnv(discord.User{ID: 1, Username: name}),
			"n":    len(name),
			"get":  func(s string) (any, error) { return s + name, nil },
			"m":    map[string]any{"k": name},
		}
		res, err := Eval(context.Background(), `user.username + get("x") + m.k + string(n)`, NewContext(env))
		require.NoError(t, err)
		assert.Equal(t, name+"x"+name+name+"1", res.String())
		waitForCache()
	}
}

// The same expression compiles again when the types of the env differ, as
// they decide how fields are read.
func TestCachedProgramPerEnvTypes(t *testing.T) {
	expression := `guild?.name ?? "none"`

	res, err := Eval(context.Background(), expression, NewContext(Env{"guild": &GuildEnv{Name: "Kite"}}))
	require.NoError(t, err)
	assert.Equal(t, "Kite", res.String())
	waitForCache()

	res, err = Eval(context.Background(), expression, NewContext(Env{"guild": map[string]any{"name": "Map"}}))
	require.NoError(t, err)
	assert.Equal(t, "Map", res.String())
	waitForCache()

	res, err = Eval(context.Background(), expression, NewContext(Env{"guild": nil}))
	require.NoError(t, err)
	assert.Equal(t, "none", res.String())
	waitForCache()

	res, err = Eval(context.Background(), expression, NewContext(Env{}))
	require.NoError(t, err)
	assert.Equal(t, "none", res.String())
}

func TestProgramKey(t *testing.T) {
	user := NewUserEnv(discord.User{ID: 1})
	a := programKey("x", Context{Env: Env{"user": user, "n": 1}})
	assert.Equal(t, a, programKey("x", Context{Env: Env{"n": 2, "user": NewUserEnv(discord.User{ID: 2})}}))
	assert.NotEqual(t, a, programKey("y", Context{Env: Env{"user": user, "n": 1}}))
	assert.NotEqual(t, a, programKey("x", Context{Env: Env{"user": user, "n": "1"}}))
	assert.NotEqual(t, a, programKey("x", Context{Env: Env{"user": user}}))
	assert.NotEqual(t, a, programKey("x", Context{Env: Env{"user": user, "n": 1}, Patchers: []ast.Visitor{fieldPatcher{}}}))
}

// Sets of ristretto are buffered, so a Get right after can miss.
func waitForCache() {
	programs.Wait()
	time.Sleep(time.Millisecond)
}

func benchEnv() Env {
	return Env{
		"user":  NewUserEnv(discord.User{ID: 123456789012345678, Username: "kite"}),
		"guild": &GuildEnv{ID: "1", Name: "Kite"},
		"xp":    12345,
		"items": []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		"arg":   func(name string) (any, error) { return 3, nil },
	}
}

func BenchmarkEval(b *testing.B) {
	for _, c := range []struct{ name, expression string }{
		{"field", `user.mention`},
		{"math", `(xp / 100) ** 0.5 * 2 + 1`},
		{"text", `upper(user.username) + " has " + string(xp) + " XP in " + guild.name`},
		{"list", `join(map(filter(items, # % 2 == 0), string(# * 10)), ", ")`},
	} {
		b.Run(c.name, func(b *testing.B) {
			env := benchEnv()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := Eval(context.Background(), c.expression, NewContext(env)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
