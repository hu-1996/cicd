package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha256(data []byte) string {
	hash := sha256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}
