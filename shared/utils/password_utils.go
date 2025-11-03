package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateOAuthDummyPassword() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "OAUTH_" + hex.EncodeToString(b)
}
