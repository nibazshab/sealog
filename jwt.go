package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// Key 签名密钥
type Key struct {
	Id  int `gorm:"primaryKey"`
	Str string
}

var jwtSecret []byte

func initializeJwtSecret() {
	var k Key

	err := k.getSignKey()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = k.addSignKey()
			if err != nil {
				log.Fatalln("error:", err)
			}
			return
		} else {
			log.Fatalln("error:", err)
		}
	}

	jwtSecret, err = base64.RawURLEncoding.DecodeString(k.Str)
	if err != nil {
		log.Fatalln("error:", err)
	}
}

func (k *Key) getSignKey() error {
	// SELECT * FROM `keys` ORDER BY `keys`.`id` DESC LIMIT 1
	return db.Last(k).Error
}

func (k *Key) addSignKey() error {
	jwtSecret = generateHMACKey()
	k.Str = base64.RawURLEncoding.EncodeToString(jwtSecret)

	// INSERT INTO `keys` (`str`) VALUES ("Qk7WIJ70b4Xds6S5L944pU8DmUSYxx5EXojyTRV9S7I") RETURNING `id`
	return db.Create(k).Error
}

func generateHMACKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatalln("error:", err)
	}

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
