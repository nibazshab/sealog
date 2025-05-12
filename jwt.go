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
	var k Key

	// SELECT * FROM `keys` ORDER BY `keys`.`id` DESC LIMIT 1
	err := db.Last(&k).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			jwtSecret = generateHMACKey()
			k.Str = base64.RawURLEncoding.EncodeToString(jwtSecret)

			// INSERT INTO `keys` (`str`) VALUES ("Qk7WIJ70b4Xds6S5L944pU8DmUSYxx5EXojyTRV9S7I") RETURNING `id`
			db.Create(&k)
		} else {
			panic("database error: " + err.Error())
		}

		return
	}

	jwtSecret, _ = base64.RawURLEncoding.DecodeString(k.Str)
}

func generateHMACKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)

	return key
}

func encodeToken() (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func decodeToken(str string) bool {
	token, err := jwt.Parse(str, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	return err == nil && token != nil && token.Valid
}
