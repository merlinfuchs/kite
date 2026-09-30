package knowledge

import (
	"os"
	"slices"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	k, err := Build("../../../kite-docs/docs", "../../../kite-service/pkg/flow/catalog.json", "https://docs.kite.onl")
	if err != nil {
		t.Fatal(err)
	}

	for _, leftover := range []string{"import ", "<EmbedFlowNode", "<NodeInfoExplorer", ":::"} {
		if strings.Contains(k, leftover) {
			t.Errorf("knowledge contains %q", leftover)
		}
	}

	raw, err := os.ReadFile("../../../kite-service/pkg/flow/catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := ParseCatalog(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range catalog.Nodes {
		if !strings.Contains(k, "Block `"+n.Title+"`: ") {
			t.Errorf("block %q is missing", n.Title)
		}
	}

	urls := PageURLs(k)
	for _, url := range []string{"https://docs.kite.onl", "https://docs.kite.onl/reference/blocks", "https://docs.kite.onl/reference/blocks/actions/action_message_pin"} {
		if !slices.Contains(urls, url) {
			t.Errorf("page %s is missing", url)
		}
	}
	if strings.Contains(k, ".md)") {
		t.Error("knowledge contains a link to a markdown file instead of a page")
	}
}

func TestPageURLs(t *testing.T) {
	k := "=== Page: A\nURL: https://a\n\nURL: https://not-a-page\n\n=== Page: B\nURL: https://b\n"
	if got := PageURLs(k); !slices.Equal(got, []string{"https://a", "https://b"}) {
		t.Errorf("PageURLs = %v", got)
	}
}
