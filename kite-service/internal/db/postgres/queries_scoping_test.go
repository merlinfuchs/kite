package postgres

import (
	"os"
	"strings"
	"testing"
)

// variable_values and message_instances carry no app_id column, so they are
// only tenant-safe when the query reaches the app through variables/messages.
// The IDs these queries take come from tenant-authored flow config and from
// Discord interactions, so dropping the join silently reopens cross-tenant
// reads and writes -- and nothing else in the test suite touches SQL, since
// there is no database here.
//
// This asserts on the query files rather than the generated code: they are the
// source of truth, and a regenerate would carry a mistake straight through.

func readQueryFile(t *testing.T, name string) string {
	t.Helper()

	content, err := os.ReadFile("queries/" + name)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	return string(content)
}

// splitQueries breaks a .sql file into one chunk per `-- name:` block, keyed by
// the query name.
func splitQueries(t *testing.T, content string) map[string]string {
	t.Helper()

	queries := make(map[string]string)
	for _, block := range strings.Split(content, "-- name: ")[1:] {
		name, _, ok := strings.Cut(block, " ")
		if !ok {
			t.Fatalf("malformed query block: %.60s", block)
		}
		queries[name] = block
	}
	return queries
}

func assertScopedByApp(t *testing.T, file string, names ...string) {
	t.Helper()

	queries := splitQueries(t, readQueryFile(t, file))

	for _, name := range names {
		body, ok := queries[name]
		if !ok {
			t.Errorf("%s: query %q not found -- renamed without updating this test?", file, name)
			continue
		}
		if !strings.Contains(body, "app_id") {
			t.Errorf("%s: query %q is not scoped by app_id:\n%s", file, name, body)
		}
	}
}

func TestVariableValueQueriesAreScopedByApp(t *testing.T) {
	assertScopedByApp(t, "variables.sql",
		"GetVariableValue",
		"GetVariableValueForUpdate",
		"GetVariableValues",
		"SetVariableValue",
		"DeleteVariableValue",
		"DeleteAllVariableValues",
	)
}

func TestMessageQueriesAreScopedByApp(t *testing.T) {
	assertScopedByApp(t, "messages.sql",
		"GetMessage",
		"GetMessageInstance",
		"GetMessageInstancesByMessage",
		"GetMessageInstancesByMessageWithHidden",
		"GetMessageInstanceByDiscordMessageId",
		"UpdateMessageInstance",
		"DeleteMessageInstance",
		"DeleteMessageInstanceByDiscordMessageId",
	)
}
