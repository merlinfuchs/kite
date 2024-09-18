package util

import (
	"crypto/sha256"
	"fmt"
)

func HashBytes(b []byte) string {
	hasher := sha256.New()
	hasher.Write(b)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
