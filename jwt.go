package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

var jwtSecret []byte

func initializeJwtSecret() {
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

type claims struct {
	jwt.RegisteredClaims
}

func encodeToken() (string, error) {
	claim := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	return token.SignedString(jwtSecret)
}

func decodeToken(str string) bool {
	token, err := jwt.Parse(str, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	return err == nil && token != nil && token.Valid
}
