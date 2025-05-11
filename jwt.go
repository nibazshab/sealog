package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"gorm.io/gorm"
)

var jwtSecret []byte

func initializeJwtKey() {
	var secret JwtSecret

	err := db.Last(&secret).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jwtSecret = generateHMACKey()
			secret.Str = base64.RawURLEncoding.EncodeToString(jwtSecret)

			db.Create(&secret)
		} else {
			panic("database error: " + err.Error())
		}

		return
	}

	jwtSecret, _ = base64.RawURLEncoding.DecodeString(secret.Str)
}

func generateHMACKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)

	return key
}
