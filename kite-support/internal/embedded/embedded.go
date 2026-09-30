package embedded

import "embed"

// index.gob is written by `kite-support index` and not committed. Embedding
// the whole directory keeps the build working when it doesn't exist yet.
//
//go:embed all:assets
var assets embed.FS

// Index returns the embedded index, or nil if the binary was built without one.
func Index() []byte {
	data, _ := assets.ReadFile("assets/index.gob")
	return data
}
