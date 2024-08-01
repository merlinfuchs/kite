package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func SecureKey() string {
	bytes := make([]byte, 64)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(bytes)
}

func HashKey(token string) string {
	hasher := sha256.New()
	hasher.Write([]byte(token))
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
