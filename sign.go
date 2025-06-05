package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type (
	// Auth 认证密码
	Auth struct {
		Id   int `gorm:"primaryKey"`
		Hash string
	}

	// Hmac 签名密钥
	Hmac struct {
		Id  int `gorm:"primaryKey"`
		Key string
	}
)

var hmacKey []byte

func initializeAuth() {
	_, err := getAuthHash()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			password := generatePassword()
			log.Println("default password:", password)
		} else {
			log.Fatalln("error:", err)
		}
	}
}

func getAuthHash() (string, error) {
	var a Auth

	// SELECT * FROM `auths` ORDER BY `auths`.`id` DESC LIMIT 1
	err := db.Last(&a).Error
	return a.Hash, err
}

func addAuthHash(str string) error {
	hash, err := cryptoPassword(str)
	if err != nil {
		return err
	}

	a := Auth{
		Hash: hash,
	}

	// INSERT INTO `auths` (`hash`) VALUES ("$2a$10$zqars0w5SAkRBDPnhrBEses1lmy15hfKkGZfO21jx/mi959v3aEfq") RETURNING `id`
	return db.Create(&a).Error
}

func generatePassword() string {
	b := make([]byte, 3)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalln("error:", err)
	}

	password := hex.EncodeToString(b)
	err = addAuthHash(password)
	if err != nil {
		log.Fatalln("error:", err)
	}

	return password
}

func cryptoPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func verifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func initializeHmac() {
	key, err := getHmacKey()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = addHmacKey()
			if err != nil {
				log.Fatalln("error:", err)
			}
			return
		} else {
			log.Fatalln("error:", err)
		}
	}

	hmacKey, err = base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		log.Fatalln("error:", err)
	}
}

func getHmacKey() (string, error) {
	var h Hmac

	// SELECT * FROM `hmacs` ORDER BY `hmacs`.`id` DESC LIMIT 1
	err := db.Last(&h).Error
	return h.Key, err
}

func addHmacKey() error {
	hmacKey = generateHmacKey()

	h := Hmac{
		Key: base64.RawURLEncoding.EncodeToString(hmacKey),
	}

	// INSERT INTO `hmacs` (`key`) VALUES ("Qk7WIJ70b4Xds6S5L944pU8DmUSYxx5EXojyTRV9S7I") RETURNING `id`
	return db.Create(&h).Error
}

func generateHmacKey() []byte {
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
	return token.SignedString(hmacKey)
}

func decodeToken(str string) bool {
	token, err := jwt.Parse(str, func(token *jwt.Token) (interface{}, error) {
		return hmacKey, nil
	})

	return err == nil && token != nil && token.Valid
}
