package hashutil

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func HashSHA256(payload []byte) string {
	h := sha256.New()
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func HmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
