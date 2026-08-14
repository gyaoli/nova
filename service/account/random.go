package account

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

type RandomIDGenerator struct{}

func (RandomIDGenerator) NewID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

type RandomTokenGenerator struct{}

func (RandomTokenGenerator) NewToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}
