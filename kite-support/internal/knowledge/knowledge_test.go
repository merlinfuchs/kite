package knowledge

import (
	"os"
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

	for _, url := range []string{"https://docs.kite.onl", "https://docs.kite.onl/reference/blocks", "https://docs.kite.onl/reference/blocks/actions/action_message_pin"} {
		if !HasURL(k, url) {
			t.Errorf("page %s is missing", url)
		}
	}
	if HasURL(k, "https://docs.kite.onl/reference/blocks/actions/action_message_pin.md") {
		t.Error("HasURL accepts a URL that isn't a page")
	}
}
