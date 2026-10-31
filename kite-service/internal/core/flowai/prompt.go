package flowai

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

var (
	instructionsMu    sync.Mutex
	instructionsCache = map[string]string{}
)

// instructionsFor returns the instructions, which come first in every request,
// with the catalog of the blocks of the integrations the app enabled. Other
// integrations are only named after the catalog, so the model can tell the
// user to enable them. Apps missing the same integrations get the same
// instructions, so the model provider can cache them.
func instructionsFor(enabled []string) string {
	var missing []flow.Integration
	var ids []string
	for _, integration := range flow.Integrations() {
		if integration.Availability != flow.AvailabilityAlways && !slices.Contains(enabled, integration.ID) {
			missing = append(missing, integration)
			ids = append(ids, integration.ID)
		}
	}
	key := strings.Join(ids, ",")

	instructionsMu.Lock()
	defer instructionsMu.Unlock()
	if res, ok := instructionsCache[key]; ok {
		return res
	}

	res := buildInstructions(missing, ids)
	instructionsCache[key] = res
	return res
}

func buildInstructions(missing []flow.Integration, missingIDs []string) string {
	catalog, err := filterCatalog(flow.CatalogJSON, func(nodeType string) bool {
		for _, id := range flow.BlockIntegrations(flow.FlowNodeType(nodeType)) {
			if slices.Contains(missingIDs, id) {
				return false
			}
		}
		return true
	})
	if err != nil {
		panic(err)
	}

	res := instructionsText + "\n\nBlock catalog:\n" + catalog
	if len(missing) > 0 {
		res += "\n\nThe app hasn't enabled these integrations, so their blocks aren't in the catalog. If the user asks for something one of them does, tell them to enable it under Integrations in the app first:"
		for _, integration := range missing {
			res += fmt.Sprintf("\n- %s: %s", integration.Name, integration.Description)
		}
	}
	return res
}

// filterCatalog compacts the catalog and keeps the nodes keep returns true
// for, in their order, and all types.
func filterCatalog(catalogJSON []byte, keep func(nodeType string) bool) (string, error) {
	// Only nodes and types are copied, so anything else in the catalog would
	// be lost.
	var catalog map[string]json.RawMessage
	if err := json.Unmarshal(catalogJSON, &catalog); err != nil {
		return "", err
	}
	for key := range catalog {
		if key != "nodes" && key != "types" {
			return "", fmt.Errorf("unexpected catalog key: %s", key)
		}
	}

	dec := json.NewDecoder(bytes.NewReader(catalog["nodes"]))
	if _, err := dec.Token(); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(`{"nodes":{`)
	first := true
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return "", err
		}
		nodeType := t.(string)

		var node json.RawMessage
		if err := dec.Decode(&node); err != nil {
			return "", err
		}
		if !keep(nodeType) {
			continue
		}

		if !first {
			buf.WriteByte(',')
		}
		first = false
		key, _ := json.Marshal(nodeType)
		buf.Write(key)
		buf.WriteByte(':')
		if err := json.Compact(&buf, node); err != nil {
			return "", err
		}
	}
	buf.WriteByte('}')
	if types, ok := catalog["types"]; ok {
		buf.WriteString(`,"types":`)
		if err := json.Compact(&buf, types); err != nil {
			return "", err
		}
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

// instructionsFile is written from kite-web/src/lib/flow/aiInstructions.ts by
// pnpm test -u in kite-web.
//
//go:embed instructions.txt
var instructionsFile string

// instructionsText leaves out the file's final newline, so the model gets the
// same text as before it was a file.
var instructionsText = strings.TrimSuffix(instructionsFile, "\n")
