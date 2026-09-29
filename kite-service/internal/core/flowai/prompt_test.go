package flowai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterCatalogKeepsOrder(t *testing.T) {
	catalog := []byte(`{"nodes": {
		"b": {"title": "B"},
		"a": {"title": "A"},
		"c": {"title": "C"}
	}}`)

	res, err := filterCatalog(catalog, func(nodeType string) bool { return nodeType != "a" })
	require.NoError(t, err)
	assert.Equal(t, `{"nodes":{"b":{"title":"B"},"c":{"title":"C"}}}`, res)
}

// Apps with the same integrations share the instructions, which the model
// provider caches.
func TestInstructionsForIsCached(t *testing.T) {
	a := instructionsFor([]string{"x", "y"})
	b := instructionsFor([]string{"y", "x"})
	assert.Equal(t, a, b)
	assert.Contains(t, a, "Block catalog:")
	assert.Contains(t, a, `"action_message_create"`)
}

func TestFilterCatalogUnknownKey(t *testing.T) {
	_, err := filterCatalog([]byte(`{"nodes": {}, "types": {}}`), func(string) bool { return true })
	assert.ErrorContains(t, err, "unexpected catalog key: types")
}
