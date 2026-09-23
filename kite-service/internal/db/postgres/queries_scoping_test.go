package postgres

import (
	"os"
	"strings"
	"testing"
)

// variable_values and message_instances have no app_id, so every query touching them
// must scope through their parent. Checked against the .sql files since there is no
// database in tests, and default-deny so new queries are covered too.

// scopedTables must never be reached without also naming app_id.
var scopedTables = []string{"variable_values", "message_instances"}

// exemptQueries must not be able to leak across apps.
var exemptQueries = map[string]string{
	// Only counts the variable's own values, the variable itself is checked by VariableAccess.
	"GetVariable": "aggregates the parent's own children, never keyed by child id",
}

// splitQueries returns each query by name with comments stripped, so a comment
// mentioning app_id can't satisfy the check.
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
