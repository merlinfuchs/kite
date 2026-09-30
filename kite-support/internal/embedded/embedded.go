package embedded

import "embed"

// knowledge.md is written by `kite-support index` and not committed. Embedding
// the whole directory keeps the build working when it doesn't exist yet.
//
//go:embed all:assets
var assets embed.FS

// Knowledge returns the embedded knowledge, or "" if the binary was built
// without it.
func Knowledge() string {
	data, _ := assets.ReadFile("assets/knowledge.md")
	return string(data)
}
