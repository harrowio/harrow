package domain

import (
	"encoding/base32"

	"crypto/rand"
)

func RandomTotpSecret() string {
	secret := make([]byte, 10)
	_, err := rand.Read(secret)
	if err != nil {
		panic("domain.RandomTotpSecret: " + err.Error())
	}

	return base32.StdEncoding.EncodeToString(secret)
}
