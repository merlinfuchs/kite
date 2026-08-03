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
// This is default-deny over the query files rather than a list of query names:
// a query added later is covered the day it is written, without anyone
// remembering to come back here. The query files are also the source of truth,
// and a regenerate would carry a mistake straight through to the generated
// code.

// scopedTables must never be reached without also naming app_id.
var scopedTables = []string{"variable_values", "message_instances"}

// exemptQueries opt out of the rule above. Every entry needs a reason, and the
// reason has to be that the query cannot leak across tenants -- not that
// scoping it is inconvenient.
var exemptQueries = map[string]string{
	// Keyed on the parent, not the child: the join only counts the variable's
	// own values, so no child row is reachable by its own id. Whether the
	// parent lookup itself is scoped is VariableAccess's problem, not this
	// query's.
	"GetVariable": "aggregates the parent's own children, never keyed by child id",
}

// splitQueries breaks a .sql file into one chunk per `-- name:` block, keyed by
// the query name, with `--` comments stripped so a rationale mentioning app_id
// cannot pass for an actual predicate.
func splitQueries(t *testing.T, content string) map[string]string {
	t.Helper()

	queries := make(map[string]string)
	for _, block := range strings.Split(content, "-- name: ")[1:] {
		name, body, ok := strings.Cut(block, "\n")
		if !ok {
			t.Fatalf("malformed query block: %.60s", block)
		}
		name, _, _ = strings.Cut(name, " ")

		var stripped strings.Builder
		for _, line := range strings.Split(body, "\n") {
			if code, _, found := strings.Cut(line, "--"); found {
				stripped.WriteString(code)
			} else {
				stripped.WriteString(line)
			}
			stripped.WriteString("\n")
		}
		queries[name] = stripped.String()
	}
	return queries
}

func assertScopedByApp(t *testing.T, file string) {
	t.Helper()

	content, err := os.ReadFile("queries/" + file)
	if err != nil {
		t.Fatalf("failed to read %s: %v", file, err)
	}

	for name, body := range splitQueries(t, string(content)) {
		if _, exempt := exemptQueries[name]; exempt {
			continue
		}
		for _, table := range scopedTables {
			if strings.Contains(body, table) && !strings.Contains(body, "app_id") {
				t.Errorf("%s: query %q touches %s but is not scoped by app_id:\n%s", file, name, table, body)
			}
		}
	}
}

func TestVariableQueriesAreScopedByApp(t *testing.T) {
	assertScopedByApp(t, "variables.sql")
}

func TestMessageQueriesAreScopedByApp(t *testing.T) {
	assertScopedByApp(t, "messages.sql")
}
