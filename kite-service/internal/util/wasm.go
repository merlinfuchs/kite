package util

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashModuleWASMBytes(wasmBytes []byte) string {
	hasher := sha256.New()
	hasher.Write(wasmBytes)
	return hex.EncodeToString(hasher.Sum(nil))
}
