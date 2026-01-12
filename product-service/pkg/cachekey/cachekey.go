package cachekey

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
)

func Hash(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha1.Sum(b)
	return hex.EncodeToString(sum[:]), nil
}
