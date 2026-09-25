package flow

import _ "embed"

// CatalogJSON describes every block: its settings as JSON Schema, its outputs
// and the flow types it can be used in. It's generated from the editor's
// schemas, run `pnpm test -u` in kite-web to update it.
//
//go:embed catalog.json
var CatalogJSON []byte

// CatalogSummary lists the blocks that can be added with the settings they
// need, for the cheaper model that checks prompts. It's generated with the
// catalog.
//
//go:embed catalog_summary.txt
var CatalogSummary string
