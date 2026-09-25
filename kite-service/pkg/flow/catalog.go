package flow

import _ "embed"

// CatalogJSON describes every block: its settings as JSON Schema, its outputs
// and the flow types it can be used in. It's generated from the editor's
// schemas, run `pnpm test -u` in kite-web to update it.
//
//go:embed catalog.json
var CatalogJSON []byte
