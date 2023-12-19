package app

import (
	"crypto/sha1"
	"encoding/hex"
)

func GenerateShortURLID(url string) string {
	hasher := sha1.New()
	hasher.Write([]byte(url))
	return hex.EncodeToString(hasher.Sum(nil))[:8]
}
